package folder

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

const (
	iconHelp = "icon: folder, confetti, factory, buildings, bird, cookie, armchair, exam, bug, mountains, " +
		"paintbrush, strategy, rocket, parachute, ghost, sun, sunhorizon, sunglasses, coffee, leaf, planet, " +
		"club, gift, flyingsaucer, heart, crownsimple, moonstars, or rocketlaunch"
	colorHelp = "color: a preset (green, yellow, red, blue-light, gray, orange, blue, pink, mauve, amber) " +
		"or a hex value like #E5477D"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a folder in your library",
		Long: "Create a folder at the top level of your library, or inside --parent-id.\n" +
			"If a folder already exists at that path the command fails and the error names\n" +
			"the existing folder's ID, so you can use that folder instead.\n\n" +
			"Save new queries into it with 'dune query create --folder-id', or move existing\n" +
			"ones with 'dune folder move-content'.\n\n" +
			"Examples:\n" +
			"  dune folder create --name Research\n" +
			"  dune folder create --name RWA --parent-id dp180Cu3jW3L --icon rocket --color blue",
		Args: cobra.NoArgs,
		RunE: runCreate,
	}
	cmd.Flags().String("name", "", "folder name; must not contain '/' (required)")
	cmd.Flags().String("parent-id", "", "create the folder inside this folder; omit for the top level")
	cmd.Flags().String("description", "", "short text that tells what the folder holds")
	cmd.Flags().String("icon", "folder", iconHelp)
	cmd.Flags().String("color", "gray", colorHelp)
	_ = cmd.MarkFlagRequired("name")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	req := models.CreateFolderRequest{
		Name:           name,
		ParentFolderID: changedString(cmd, "parent-id"),
		Description:    changedString(cmd, "description"),
		Icon:           changedString(cmd, "icon"),
		Color:          changedString(cmd, "color"),
	}
	f, err := cmdutil.ClientFromCmd(cmd).CreateFolder(req)
	if err != nil {
		if apiErr, details := folderAPIError(err); apiErr != nil && apiErr.Code == "folder_already_exists" &&
			details.ExistingFolderID != nil {
			return fmt.Errorf("%w\nThe existing folder's ID is %s", err, *details.ExistingFolderID)
		}
		return err
	}
	return printFolder(cmd, f, "Created")
}
