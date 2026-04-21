package query

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <query-id>",
		Short: "Execute a saved Dune query by its ID and display results",
		Long: "Execute a saved Dune query by its numeric ID. By default, waits for the\n" +
			"execution to complete (polling every 2 seconds) and displays the result rows.\n" +
			"Use --no-wait to submit the execution and exit immediately with just the\n" +
			"execution ID; then fetch results later with 'dune execution results <execution-id>'.\n\n" +
			"Credits are consumed based on actual compute resources used. Use --performance\n" +
			"to select the engine size (medium or large).\n\n" +
			"Important: if the query targets tables with known partition columns (returned by\n" +
			"'dune dataset search' or 'dune dataset search-by-contract'), ensure the SQL includes\n" +
			"a WHERE filter on those partition columns (e.g. WHERE block_date >= CURRENT_DATE -\n" +
			"INTERVAL '7' DAY). This enables partition pruning and significantly reduces query cost.\n\n" +
			"Examples:\n" +
			"  dune query run 12345\n" +
			"  dune query run 12345 --param wallet=0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 --param days=30\n" +
			"  dune query run 12345 --performance large --limit 100\n" +
			"  dune query run 12345 --no-wait\n" +
			"  dune query run 12345 --timeout 600",
		Args: cobra.ExactArgs(1),
		RunE: runRun,
	}

	cmd.Flags().StringArray("param", nil, "typed query parameter in key=value format (repeatable); supported types: text, number (stringified, e.g. '30'), datetime (YYYY-MM-DD HH:mm:ss), enum")
	cmd.Flags().String("performance", "medium", `engine size for the execution: "small", "medium" (default) or "large"; credits are consumed based on actual compute resources used`)
	cmd.Flags().Int("limit", 0, "maximum number of result rows to return (0 = all available rows)")
	cmd.Flags().Bool("no-wait", false, "submit the execution and exit immediately, printing only the execution ID and state")
	cmd.Flags().Int("timeout", 300, "maximum seconds to wait for the execution to complete before timing out")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runRun(cmd *cobra.Command, args []string) error {
	queryID, err := parseQueryID(args[0])
	if err != nil {
		return err
	}

	paramFlags, _ := cmd.Flags().GetStringArray("param")
	params, err := parseParams(paramFlags)
	if err != nil {
		return err
	}

	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}

	req := models.ExecuteRequest{
		QueryID:     queryID,
		Performance: performance,
	}
	if len(params) > 0 {
		req.QueryParameters = params
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if noWait {
		return runNoWait(cmd, req)
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	return runWait(cmd, req, timeout)
}

func runNoWait(cmd *cobra.Command, req models.ExecuteRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.QueryExecute(req)
	if err != nil {
		return err
	}

	return displayExecuteResponse(cmd, resp)
}

func runWait(cmd *cobra.Command, req models.ExecuteRequest, timeout int) error {
	client := cmdutil.ClientFromCmd(cmd)

	exec, err := client.RunQuery(req)
	if err != nil {
		return err
	}

	return waitAndDisplay(cmd, exec, timeout)
}

func parseParams(raw []string) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	params := make(map[string]any, len(raw))
	for _, s := range raw {
		key, value, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("invalid parameter %q: expected key=value format", s)
		}
		if key == "" {
			return nil, fmt.Errorf("invalid parameter %q: key cannot be empty", s)
		}
		params[key] = value
	}
	return params, nil
}
