package folder

import (
	"errors"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <folder-id>",
		Short: "Rename a folder or change its description, icon, or color",
		Long: "Change a folder's name, description, icon, or color. Flags you leave out keep\n" +
			"their current value. Renaming updates the paths of every subfolder; queries and\n" +
			"dashboards stay where they are. To move a folder use 'dune folder move'.\n\n" +
			"Examples:\n" +
			"  dune folder update dp180Cu3jW3L --name \"RWA (2026)\"\n" +
			"  dune folder update dp180Cu3jW3L --description \"\"   # clears the description\n" +
			"  dune folder update dp180Cu3jW3L --icon rocket --color #E5477D",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}
	cmd.Flags().String("name", "", "new name; must not contain '/'")
	cmd.Flags().String("description", "", "new description; an empty value clears it")
	cmd.Flags().String("icon", "", "new "+iconHelp)
	cmd.Flags().String("color", "", "new "+colorHelp)
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runUpdate(cmd *cobra.Command, posArgs []string) error {
	req := models.UpdateFolderRequest{
		Name:        changedString(cmd, "name"),
		Description: changedString(cmd, "description"),
		Icon:        changedString(cmd, "icon"),
		Color:       changedString(cmd, "color"),
	}
	if req == (models.UpdateFolderRequest{}) {
		return errors.New("at least one flag must be provided (--name, --description, --icon, or --color)")
	}
	f, err := cmdutil.ClientFromCmd(cmd).UpdateFolder(posArgs[0], req)
	if err != nil {
		return err
	}
	return printFolder(cmd, f, "Updated")
}
