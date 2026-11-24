package folder

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <folder-id>",
		Short: "Show a folder, its subfolders, and the queries and dashboards in it",
		Long: "Show one folder: its details, its direct subfolders, and the queries and\n" +
			"dashboards inside it that are not archived. Page through the items with\n" +
			"--limit and --offset.\n\n" +
			"Examples:\n" +
			"  dune folder get dp180Cu3jW3L\n" +
			"  dune folder get dp180Cu3jW3L --limit 100 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}
	cmd.Flags().Int("limit", 50, "maximum number of queries and dashboards to return (1-100)")
	cmd.Flags().Int("offset", 0, "items to skip; use the next offset from the previous page")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runGet(cmd *cobra.Command, posArgs []string) error {
	opts := models.GetFolderOptions{Limit: changedInt(cmd, "limit"), Offset: changedInt(cmd, "offset")}
	res, err := cmdutil.ClientFromCmd(cmd).GetFolder(posArgs[0], opts)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintJSON(cmd, res)
	}

	f := res.Folder
	fmt.Fprintf(w, "ID:          %s\n", f.FolderID)
	fmt.Fprintf(w, "Path:        %s\n", f.Path)
	if f.Description != nil && *f.Description != "" {
		fmt.Fprintf(w, "Description: %s\n", *f.Description)
	}
	fmt.Fprintf(w, "Icon:        %s\n", f.Icon)
	fmt.Fprintf(w, "Color:       %s\n", f.Color)

	if len(res.Subfolders) > 0 {
		fmt.Fprintln(w, "\nSubfolders:")
		printFolderSummaries(w, res.Subfolders)
		if res.HasMoreSubfolders {
			fmt.Fprintf(w, "More subfolders: dune folder list --parent-path %q\n", f.Path)
		}
	}

	fmt.Fprintf(w, "\nItems (%d):\n", res.TotalItems)
	rows := make([][]string, len(res.Items))
	for i, it := range res.Items {
		rows[i] = []string{it.Type, fmt.Sprintf("%d", it.ID), it.Name, fmt.Sprintf("%t", it.IsPrivate)}
	}
	output.PrintTable(w, []string{"TYPE", "ID", "NAME", "PRIVATE"}, rows)
	if res.NextOffset != nil {
		fmt.Fprintf(w, "\nMore items: --offset %d\n", *res.NextOffset)
	}
	return nil
}
