package folder

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <folder-id>",
		Short: "Delete a folder",
		Long: "Delete a folder. An empty folder is deleted right away. A folder that still has\n" +
			"subfolders, queries, or dashboards is refused, and the error says how many of\n" +
			"each it holds.\n\n" +
			"--recursive deletes every subfolder and ARCHIVES every query and dashboard\n" +
			"inside: they leave the library and can be restored from the archive, but are not\n" +
			"moved anywhere. To keep content, move it out first with 'dune folder move-content'.\n\n" +
			"Examples:\n" +
			"  dune folder delete dp180Cu3jW3L\n" +
			"  dune folder delete dp180Cu3jW3L --recursive",
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}
	cmd.Flags().Bool("recursive", false, "delete a folder that is not empty, archiving its queries and dashboards")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runDelete(cmd *cobra.Command, posArgs []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{"recursive": "recursive"})
	args["folder_id"] = posArgs[0]

	var res struct {
		Path                   string `json:"path"`
		DeletedFolderCount     int    `json:"deleted_folder_count"`
		ArchivedQueryCount     int    `json:"archived_query_count"`
		ArchivedDashboardCount int    `json:"archived_dashboard_count"`
	}
	raw, err := cmdutil.CallTool(cmd, "deleteDuneFolder", args, &res)
	if err != nil {
		// The server's refusal names the MCP argument's effect, not the CLI flag.
		if args["recursive"] != true && strings.Contains(err.Error(), "is not empty") {
			return fmt.Errorf("%w\nTo delete it anyway, pass --recursive", err)
		}
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	fmt.Fprintf(w, "Deleted folder %s: %d folders deleted, %d queries and %d dashboards archived\n",
		res.Path, res.DeletedFolderCount, res.ArchivedQueryCount, res.ArchivedDashboardCount)
	return nil
}
