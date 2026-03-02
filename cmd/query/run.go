package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <query-id>",
		Short: "Execute a saved query and display results",
		Args:  cobra.ExactArgs(1),
		RunE:  runRun,
	}

	cmd.Flags().StringArray("param", nil, "query parameter in key=value format (repeatable)")
	cmd.Flags().String("performance", "medium", `performance tier: "medium" or "large"`)
	cmd.Flags().Int("limit", 0, "maximum number of rows to display (0 = all)")
	cmd.Flags().Bool("no-wait", false, "submit execution and exit without waiting for results")
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

	performance, _ := cmd.Flags().GetString("performance")
	if performance != "medium" && performance != "large" {
		return fmt.Errorf("invalid performance tier %q: must be \"medium\" or \"large\"", performance)
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
	return runWait(cmd, req)
}

func runNoWait(cmd *cobra.Command, req models.ExecuteRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.QueryExecute(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Execution ID: %s\n", resp.ExecutionID)
		fmt.Fprintf(w, "State:        %s\n", resp.State)
		return nil
	}
}

func runWait(cmd *cobra.Command, req models.ExecuteRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	exec, err := client.RunQuery(req)
	if err != nil {
		return err
	}

	resp, err := exec.WaitGetResults(5*time.Second, 60)
	if err != nil {
		return err
	}

	if resp.State != "QUERY_STATE_COMPLETED" {
		msg := fmt.Sprintf("query execution failed with state %s", resp.State)
		if resp.Error != nil {
			msg += fmt.Sprintf(": %s", resp.Error.Message)
		}
		return fmt.Errorf("%s", msg)
	}

	return output.DisplayResults(cmd, resp)
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

