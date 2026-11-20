package folder

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the folders under a path in your library",
		Long: "List the folders directly under a path, ordered by name, with the number of\n" +
			"queries and dashboards in each. Use the folder IDs with 'dune folder get',\n" +
			"'dune query create --folder-id', or 'dune folder move-content --to'.\n\n" +
			"Examples:\n" +
			"  dune folder list\n" +
			"  dune folder list --parent-path /Research\n" +
			"  dune folder list --name-contains rwa -o json",
		Args: cobra.NoArgs,
		RunE: runList,
	}
	cmd.Flags().String("parent-path", "/", "path of the folder whose subfolders to list; '/' lists the top level")
	cmd.Flags().String("name-contains", "", "case-insensitive substring filter on the folder name")
	cmd.Flags().Int("limit", 50, "maximum number of folders to return (1-100)")
	cmd.Flags().Int("offset", 0, "folders to skip; use the next offset from the previous page")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)
	parentPath, _ := cmd.Flags().GetString("parent-path")
	parentID, err := resolvePath(client, parentPath)
	if err != nil {
		return err
	}
	nameContains, _ := cmd.Flags().GetString("name-contains")
	res, err := client.ListFolders(models.ListFoldersOptions{
		ParentFolderID: parentID,
		NameContains:   nameContains,
		Limit:          changedInt(cmd, "limit"),
		Offset:         changedInt(cmd, "offset"),
	})
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintJSON(w, res)
	}
	if len(res.Folders) == 0 {
		fmt.Fprintln(w, "No folders found.")
		return nil
	}
	printFolderSummaries(w, res.Folders)
	if res.NextOffset != nil {
		fmt.Fprintf(w, "\nMore folders: --offset %d\n", *res.NextOffset)
	}
	return nil
}

// resolvePath returns the ID of the folder at path, or "" for the library root ("/"). The API
// lists folders by parent ID only, so it walks the path one segment at a time.
func resolvePath(client dune.DuneClient, path string) (string, error) {
	parentID := ""
	walked := ""
	for _, name := range strings.Split(path, "/") {
		if name == "" {
			continue
		}
		walked += "/" + name
		id, err := findChild(client, parentID, name)
		if err != nil {
			return "", fmt.Errorf("resolving %s: %w", walked, err)
		}
		if id == "" {
			return "", fmt.Errorf("no folder at %s", walked)
		}
		parentID = id
	}
	return parentID, nil
}

// findChild returns the ID of the folder under parentID whose name is exactly name, or "" when
// there is none. The name filter is a substring match, so it pages until it finds the exact name.
func findChild(client dune.DuneClient, parentID, name string) (string, error) {
	opts := models.ListFoldersOptions{ParentFolderID: parentID, NameContains: name, Limit: 100}
	for {
		res, err := client.ListFolders(opts)
		if err != nil {
			return "", err
		}
		for _, f := range res.Folders {
			if f.Name == name {
				return f.FolderID, nil
			}
		}
		if res.NextOffset == nil {
			return "", nil
		}
		opts.Offset = *res.NextOffset
	}
}
