package folder

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
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
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"parent-path":   "parent_path",
		"name-contains": "name_contains",
		"limit":         "limit",
		"offset":        "offset",
	})
	var res struct {
		Folders    []folderSummary `json:"folders"`
		NextOffset *int            `json:"next_offset"`
	}
	raw, err := cmdutil.CallTool(cmd, "listDuneFolders", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
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
