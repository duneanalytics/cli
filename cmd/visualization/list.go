package visualization

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all visualizations attached to a query",
		Long: "List all visualizations attached to a given query.\n\n" +
			"Use this to discover what charts, tables, and counters already exist\n" +
			"for a query before creating new ones. Each result includes the\n" +
			"visualization ID, name, type, and timestamps.\n\n" +
			"Use 'dune viz get <id>' with a specific ID to fetch full details\n" +
			"including chart options.\n\n" +
			"Examples:\n" +
			"  dune viz list --query-id 12345\n" +
			"  dune viz list --query-id 12345 --limit 10 --offset 0 -o json",
		RunE: runList,
	}

	cmd.Flags().Int("query-id", 0, "ID of the query to list visualizations for (required)")
	cmd.Flags().Int("limit", 25, "maximum number of results to return")
	cmd.Flags().Int("offset", 0, "number of results to skip")
	_ = cmd.MarkFlagRequired("query-id")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	queryID, _ := cmd.Flags().GetInt("query-id")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.ListQueryVisualizations(queryID, limit, offset)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		rows := make([][]string, len(resp.Results))
		for i, v := range resp.Results {
			rows[i] = []string{
				fmt.Sprintf("%d", v.ID),
				v.Name,
				v.Type,
				v.CreatedAt,
			}
		}
		output.PrintTable(w, []string{"ID", "Name", "Type", "Created"}, rows)
		return nil
	}
}
