package query

import (
	"errors"
	"fmt"

	"github.com/duneanalytics/cli/cmd/visualization"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newCreateAndRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-and-run",
		Short: "Save a new query, run it, optionally chart it, and show the first rows",
		Long: "Create a saved query, execute it, optionally add a visualization, wait for the\n" +
			"results, and show the first rows, in one step. It takes the same flags as\n" +
			"'dune query create', plus the execution and chart flags below.\n\n" +
			"If the execution is still running after --timeout seconds, the command prints the\n" +
			"query and execution IDs; fetch the rows later with 'dune execution results'. If the\n" +
			"query, execution, or chart fails, everything that was created is still printed and\n" +
			"the command exits with an error.\n\n" +
			"If the SQL targets tables with partition columns (shown by 'dune dataset search'),\n" +
			"filter on them to reduce query cost.\n\n" +
			"Examples:\n" +
			"  dune query create-and-run --name \"Daily blocks\" --sql \"SELECT ...\" --performance small\n" +
			"  dune query create-and-run --name \"Daily volume\" --sql @volume.sql --chart-type line \\\n" +
			"    --config @line.json --limit 5 -o json\n\n" + visualization.ChartConfigHelp,
		Args: cobra.NoArgs,
		RunE: runCreateAndRun,
	}
	addCreateFlags(cmd)
	cmd.Flags().String("performance", "",
		`engine tier override: "small", "medium", or "large"; omit to let the API auto-select based on your plan`)
	cmd.Flags().Int("timeout", 60, "maximum seconds to wait for the execution (0 returns right after it starts)")
	cmd.Flags().Int("limit", 20, "maximum number of result rows to show (0-32000)")
	cmd.Flags().String("viz-name", "", "name of the visualization; defaults to the query name")
	visualization.AddChartFlags(cmd)
	output.AddFormatFlag(cmd, "text")
	return cmd
}

type toolError struct {
	Message string `json:"message"`
}

func runCreateAndRun(cmd *cobra.Command, _ []string) error {
	args, err := createArgs(cmd)
	if err != nil {
		return err
	}
	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}
	if performance != "" {
		args["performance"] = performance
	}
	for k, v := range cmdutil.ToolArgs(cmd, map[string]string{"timeout": "timeout", "limit": "max_rows_returned"}) {
		args[k] = v
	}
	chart, err := visualization.ChartArgs(cmd)
	if err != nil {
		return err
	}
	if chart != nil {
		if cmd.Flags().Changed("viz-name") {
			chart["visualizationName"], _ = cmd.Flags().GetString("viz-name")
		}
		args["visualization"] = chart
	} else if cmd.Flags().Changed("viz-name") {
		return errors.New("--viz-name needs --chart-type and --config")
	}

	var res struct {
		Query struct {
			QueryID int    `json:"query_id"`
			URL     string `json:"url"`
		} `json:"query"`
		Execution *struct {
			ExecutionID string `json:"execution_id"`
			State       string `json:"state"`
			EngineUsed  string `json:"engine_used"`
		} `json:"execution"`
		ExecutionError *toolError `json:"execution_error"`
		Visualization  *struct {
			VisualizationID     int    `json:"visualizationId"`
			InteractiveChartURL string `json:"interactiveChartUrl"`
			ChartJpegURL        string `json:"chartJpegUrl"`
		} `json:"visualization"`
		VisualizationError *toolError              `json:"visualization_error"`
		ResultPreview      *output.ExecutionResult `json:"result_preview"`
		ResultPreviewError *toolError              `json:"result_preview_error"`
	}
	raw, err := cmdutil.CallTool(cmd, "createAndExecuteQuery", args, &res)
	if err != nil {
		return err
	}

	var errs []error
	if e := res.ExecutionError; e != nil {
		errs = append(errs, fmt.Errorf("running query %d: %s", res.Query.QueryID, e.Message))
	}
	if e := res.VisualizationError; e != nil {
		errs = append(errs, fmt.Errorf("creating the visualization: %s", e.Message))
	}
	if e := res.ResultPreviewError; e != nil {
		errs = append(errs, fmt.Errorf("reading the results: %s", e.Message))
	}
	if p := res.ResultPreview; p != nil {
		errs = append(errs, p.Err())
	}

	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		if err := output.PrintRawJSON(w, raw); err != nil {
			return err
		}
		return errors.Join(errs...)
	}
	fmt.Fprintf(w, "Created query %d: %s\n", res.Query.QueryID, res.Query.URL)
	if e := res.Execution; e != nil {
		fmt.Fprintf(w, "Execution %s: %s (engine %s)\n", e.ExecutionID, e.State, e.EngineUsed)
	}
	if v := res.Visualization; v != nil {
		fmt.Fprintf(w, "Visualization %d: %s\n", v.VisualizationID, v.InteractiveChartURL)
		fmt.Fprintf(w, "Image: %s\n", v.ChartJpegURL)
	}
	if p := res.ResultPreview; p != nil && p.Err() == nil {
		fmt.Fprintln(w)
		output.PrintExecutionResult(w, *p)
		if p.State != "COMPLETED" {
			fmt.Fprintf(w, "\nFetch the rows when it finishes: dune execution results %s\n", p.ExecutionID)
		}
	}
	return errors.Join(errs...)
}
