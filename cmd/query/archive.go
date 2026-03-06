package query

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <query-id>",
		Short: "Archive a saved Dune query by ID",
		Long: "Mark a Dune query as archived. Archived queries are hidden from the library\n" +
			"but can still be retrieved by ID. You must own the query or have edit access\n" +
			"via team membership.",
		Args: cobra.ExactArgs(1),
		RunE: runArchive,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	queryID, err := parseQueryID(args[0])
	if err != nil {
		return err
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
