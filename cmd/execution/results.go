package execution

import (
	"fmt"
	"time"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// PollInterval controls the polling interval when waiting for execution results.
var PollInterval = 2 * time.Second

func newResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <execution-id>",
		Short: "Fetch results of a query execution",
		Args:  cobra.ExactArgs(1),
		RunE:  runResults,
	}

	cmd.Flags().Int("limit", 0, "maximum number of rows to return (0 = all)")
	cmd.Flags().Int("offset", 0, "number of rows to skip")
	cmd.Flags().Bool("no-wait", false, "fetch current state without waiting for completion")
	cmd.Flags().Int("timeout", 300, "maximum seconds to wait for completion")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runResults(cmd *cobra.Command, args []string) error {
	executionID := args[0]

	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	noWait, _ := cmd.Flags().GetBool("no-wait")

	if limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative, got %d", offset)
	}

	opts := models.ResultOptions{}
	if limit > 0 || offset > 0 {
		opts.Page = &models.ResultPageOption{
			Offset: uint64(offset),
			Limit:  uint32(limit),
		}
	}

	client := cmdutil.ClientFromCmd(cmd)

	if noWait {
		resp, err := client.QueryResultsV2(executionID, opts)
		if err != nil {
			return err
		}
		return handleResultsResponse(cmd, executionID, resp)
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	intervalSec := int(PollInterval.Seconds())
	maxRetries := timeout
	if intervalSec > 0 {
		maxRetries = timeout / intervalSec
	}
	if maxRetries < 1 {
		maxRetries = 1
	}

	return waitForResults(cmd, client, executionID, opts, PollInterval, maxRetries)
}

func waitForResults(
	cmd *cobra.Command,
	client dune.DuneClient,
	executionID string,
	opts models.ResultOptions,
	interval time.Duration,
	maxRetries int,
) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := client.QueryResultsV2(executionID, opts)
		if err != nil {
			return err
		}

		switch resp.State {
		case "QUERY_STATE_PENDING", "QUERY_STATE_EXECUTING":
			// still running, wait and retry
		default:
			return handleResultsResponse(cmd, executionID, resp)
		}

		if i < maxRetries-1 {
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("timed out waiting for execution %s to complete", executionID)
}

func handleResultsResponse(cmd *cobra.Command, executionID string, resp *models.ResultsResponse) error {
	switch resp.State {
	case "QUERY_STATE_COMPLETED":
		return output.DisplayResults(cmd, resp)
	case "QUERY_STATE_PENDING", "QUERY_STATE_EXECUTING":
		w := cmd.OutOrStdout()
		switch output.FormatFromCmd(cmd) {
		case output.FormatJSON:
			return output.PrintJSON(w, resp)
		default:
			fmt.Fprintf(w, "Execution ID: %s\n", executionID)
			fmt.Fprintf(w, "State:        %s\n", resp.State)
			return nil
		}
	case "QUERY_STATE_FAILED":
		msg := "execution failed"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return fmt.Errorf("%s", msg)
	case "QUERY_STATE_CANCELLED":
		return fmt.Errorf("execution was cancelled")
	default:
		return fmt.Errorf("unexpected execution state: %s", resp.State)
	}
}
