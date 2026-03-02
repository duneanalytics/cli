package execution

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <execution-id>",
		Short: "Fetch results of a query execution",
		Args:  cobra.ExactArgs(1),
		RunE:  runResults,
	}

	cmd.Flags().Int("limit", 0, "maximum number of rows to return (0 = all)")
	cmd.Flags().Int("offset", 0, "number of rows to skip")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runResults(cmd *cobra.Command, args []string) error {
	executionID := args[0]

	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")

	opts := models.ResultOptions{}
	if limit > 0 || offset > 0 {
		opts.Page = &models.ResultPageOption{
			Offset: uint64(offset),
			Limit:  uint32(limit),
		}
	}

	client := cmdutil.ClientFromCmd(cmd)
	resp, err := client.QueryResultsV2(executionID, opts)
	if err != nil {
		return err
	}

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
