package folder

import (
	"errors"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
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
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"name":        "name",
		"description": "description",
		"icon":        "icon",
		"color":       "color",
	})
	if len(args) == 0 {
		return errors.New("at least one flag must be provided (--name, --description, --icon, or --color)")
	}
	if d, ok := args["description"]; ok && d == "" {
		args["description"] = nil
	}
	args["folder_id"] = posArgs[0]
	return runFolderTool(cmd, "updateDuneFolder", args, "Updated")
}
