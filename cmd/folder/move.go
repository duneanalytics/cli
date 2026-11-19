package folder

import (
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <folder-id>",
		Short: "Move a folder under another folder or to the top level",
		Long: "Move a folder, with everything inside it, under another folder in the same\n" +
			"library, or to the top level with --to-root. A folder cannot move into itself or\n" +
			"one of its own subfolders, and the destination must not already hold a folder\n" +
			"with the same name.\n\n" +
			"Examples:\n" +
			"  dune folder move dp180Cu3jW3L --to aZ3kP0qLm9Xc\n" +
			"  dune folder move dp180Cu3jW3L --to-root",
		Args: cobra.ExactArgs(1),
		RunE: runMove,
	}
	addDestinationFlags(cmd, "the folder")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runMove(cmd *cobra.Command, posArgs []string) error {
	req := models.MoveFolderRequest{ParentFolderID: destination(cmd)}
	f, err := cmdutil.ClientFromCmd(cmd).MoveFolder(posArgs[0], req)
	if err != nil {
		return err
	}
	return printFolder(cmd, f, "Moved")
}
