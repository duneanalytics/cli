package folder

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
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
	recursive, _ := cmd.Flags().GetBool("recursive")
	res, err := cmdutil.ClientFromCmd(cmd).DeleteFolder(posArgs[0], recursive)
	if err != nil {
		if apiErr, details := folderAPIError(err); apiErr != nil && apiErr.Code == "folder_not_empty" && !recursive {
			return fmt.Errorf("%w%s\nTo delete it anyway, pass --recursive", err, notEmptyCounts(details))
		}
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintJSON(cmd, res)
	}
	fmt.Fprintf(w, "Deleted folder %s: %d folders deleted, %d queries and %d dashboards archived\n",
		res.Path, res.DeletedFolderCount, res.ArchivedQueryCount, res.ArchivedDashboardCount)
	return nil
}

// notEmptyCounts describes what a folder_not_empty refusal says the folder holds, or returns ""
// when the error carries no counts.
func notEmptyCounts(d models.FolderErrorDetails) string {
	if d.SubfolderCount == nil && d.QueryCount == nil && d.DashboardCount == nil {
		return ""
	}
	count := func(n *int) int {
		if n == nil {
			return 0
		}
		return *n
	}
	return fmt.Sprintf("\nIt holds %d subfolders, %d queries and %d dashboards.",
		count(d.SubfolderCount), count(d.QueryCount), count(d.DashboardCount))
}
