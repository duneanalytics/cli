package folder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
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

// printFolder prints a folder returned by a create, update, or move as "<verb> folder <path> (<id>)".
func printFolder(cmd *cobra.Command, f *models.Folder, verb string) error {
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintJSON(w, f)
	}
	fmt.Fprintf(w, "%s folder %s (%s)\n", verb, f.Path, f.FolderID)
	return nil
}

// folderAPIError returns the API error in err with its folder error details, or nil when err
// is not an API error.
func folderAPIError(err error) (*dune.APIError, models.FolderErrorDetails) {
	var details models.FolderErrorDetails
	var apiErr *dune.APIError
	if !errors.As(err, &apiErr) {
		return nil, details
	}
	if len(apiErr.Details) > 0 {
		_ = json.Unmarshal(apiErr.Details, &details)
	}
	return apiErr, details
}

// changedInt returns an int flag's value, or 0 (left to the server default) when it was not set.
func changedInt(cmd *cobra.Command, name string) int {
	if !cmd.Flags().Changed(name) {
		return 0
	}
	v, _ := cmd.Flags().GetInt(name)
	return v
}

// changedString returns a string flag's value, or nil when it was not set.
func changedString(cmd *cobra.Command, name string) *string {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v, _ := cmd.Flags().GetString(name)
	return &v
}

func printFolderSummaries(w io.Writer, folders []models.FolderSummary) {
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
func destination(cmd *cobra.Command) *string {
	if root, _ := cmd.Flags().GetBool("to-root"); root {
		return nil
	}
	to, _ := cmd.Flags().GetString("to")
	return &to
}
