package query

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <query-id>",
		Short: "Archive a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  runArchive,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	queryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid query ID %q: must be an integer", args[0])
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.ArchiveQuery(queryID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Archived query %d\n", resp.QueryID)
		return nil
	}
}
