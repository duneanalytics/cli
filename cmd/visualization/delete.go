package visualization

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <visualization_id>",
		Short: "Permanently delete a visualization by its ID",
		Long: "Permanently delete a visualization by its ID. This action cannot be undone.\n\n" +
			"Use 'dune viz list --query-id <id>' to find visualization IDs for a given query.\n\n" +
			"Examples:\n" +
			"  dune viz delete 12345\n" +
			"  dune viz delete 12345 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	vizID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid visualization ID %q: must be an integer", args[0])
	}

	client := cmdutil.ClientFromCmd(cmd)

	_, err = client.DeleteVisualization(vizID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, map[string]any{"ok": true})
	default:
		fmt.Fprintf(w, "Deleted visualization %d\n", vizID)
		return nil
	}
}
