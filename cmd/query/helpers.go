package query

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func parseQueryID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid query ID %q: must be an integer", arg)
	}
	return id, nil
}

func parsePerformance(cmd *cobra.Command) (string, error) {
	performance, _ := cmd.Flags().GetString("performance")
	switch performance {
	case "", "free", "medium", "large":
		return performance, nil
	default:
		return "", fmt.Errorf(
			"invalid performance tier %q: must be \"free\", \"medium\" or \"large\"",
			performance,
		)
	}
}

func waitAndDisplay(cmd *cobra.Command, exec dune.Execution, timeout int) error {
	maxRetries := timeout / 2
	if maxRetries < 1 {
		maxRetries = 1
	}
	resp, err := exec.WaitGetResults(2*time.Second, maxRetries)
	if err != nil {
		return err
	}

	if resp.State != "QUERY_STATE_COMPLETED" {
		msg := fmt.Sprintf("query execution failed with state %s", resp.State)
		if resp.Error != nil {
			msg += fmt.Sprintf(": %s", resp.Error.Message)
		}
		return errors.New(msg)
	}

	return output.DisplayResults(cmd, resp)
}

func displayExecuteResponse(cmd *cobra.Command, resp *models.ExecuteResponse) error {
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
