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
		Short: "Execute a saved query and display results",
		Long: "Execute a saved DuneSQL query by its numeric ID and display results.\n\n" +
			"By default, polls every 5 seconds for up to ~5 minutes waiting for completion.\n" +
			"Use --no-wait to submit the execution and exit immediately; then fetch\n" +
			"results later with 'dune execution results <execution-id>'.\n\n" +
			"Examples:\n" +
			"  dune query run 12345\n" +
			"  dune query run 12345 --param wallet=0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 --param days=30\n" +
			"  dune query run 12345 --performance large --limit 100\n" +
			"  dune query run 12345 --no-wait",
		Args: cobra.ExactArgs(1),
		RunE: runRun,
	}

	cmd.Flags().StringArray("param", nil, "query parameter in key=value format (repeatable)")
	cmd.Flags().String("performance", "medium", `performance tier: "medium" (default) or "large" for higher compute resources`)
	cmd.Flags().Int("limit", 0, "maximum number of rows to display (0 = all)")
	cmd.Flags().Bool("no-wait", false, "submit execution and exit without waiting for results")
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
