package folder

import (
	"fmt"
	"io"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// NewFolderCmd returns the `folder` parent command.
func NewFolderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folder",
		Short: "Organize saved queries and dashboards into library folders",
		Long: "Create, browse, rename, move, and delete folders in your Dune library, and move\n" +
			"queries and dashboards between them.\n\n" +
			"A team API key works on the team's library and needs an editor or admin behind\n" +
			"the key to change it; a personal key works on your own library.\n\n" +
			"Subcommands:\n" +
			"  list          - List the folders under a path\n" +
			"  get           - Show a folder, its subfolders, and the queries and dashboards in it\n" +
			"  create        - Create a folder\n" +
			"  update        - Rename a folder or change its description, icon, or color\n" +
			"  move          - Move a folder under another folder or to the top level\n" +
			"  move-content  - Move queries and dashboards into a folder or out of folders\n" +
			"  delete        - Delete a folder",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newMoveCmd())
	cmd.AddCommand(newMoveContentCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

type folderSummary struct {
	FolderID  string `json:"folder_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	ItemCount int    `json:"item_count"`
}

type folderDetails struct {
	FolderID       string  `json:"folder_id"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	ParentFolderID *string `json:"parent_folder_id"`
	Description    *string `json:"description"`
	Icon           string  `json:"icon"`
	Color          string  `json:"color"`
}

type folderResult struct {
	Folder folderDetails `json:"folder"`
}

// runFolderTool calls a tool that returns {folder} and prints "<verb> folder <path> (<id>)".
func runFolderTool(cmd *cobra.Command, tool string, args map[string]any, verb string) error {
	var res folderResult
	raw, err := cmdutil.CallTool(cmd, tool, args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	fmt.Fprintf(w, "%s folder %s (%s)\n", verb, res.Folder.Path, res.Folder.FolderID)
	return nil
}

func printFolderSummaries(w io.Writer, folders []folderSummary) {
	rows := make([][]string, len(folders))
	for i, f := range folders {
		rows[i] = []string{f.FolderID, f.Path, fmt.Sprintf("%d", f.ItemCount)}
	}
	output.PrintTable(w, []string{"FOLDER_ID", "PATH", "ITEMS"}, rows)
}

// addDestinationFlags registers --to and --to-root, exactly one of which is required.
func addDestinationFlags(cmd *cobra.Command, what string) {
	cmd.Flags().String("to", "", "ID of the folder to move "+what+" into")
	cmd.Flags().Bool("to-root", false, "move "+what+" to the top level of the library instead")
	cmd.MarkFlagsMutuallyExclusive("to", "to-root")
	cmd.MarkFlagsOneRequired("to", "to-root")
}

// destination returns the folder ID from --to, or nil for --to-root.
func destination(cmd *cobra.Command) any {
	if root, _ := cmd.Flags().GetBool("to-root"); root {
		return nil
	}
	to, _ := cmd.Flags().GetString("to")
	return to
}
