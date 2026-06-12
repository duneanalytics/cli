package matview

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the materialized views you own",
		Long: "List materialized views owned by the authenticated user or team, paginated.\n\n" +
			"Examples:\n" +
			"  dune matview list\n" +
			"  dune matview list --limit 50 --offset 100\n" +
			"  dune matview list --all -o json",
		RunE: runList,
	}

	cmd.Flags().Int("limit", 100, "max number of matviews to return per page (max 1000)")
	cmd.Flags().Int("offset", 0, "pagination offset")
	cmd.Flags().Bool("all", false, "fetch all pages (ignores --offset)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	all, _ := cmd.Flags().GetBool("all")

	client := cmdutil.ClientFromCmd(cmd)

	result := &models.ListMaterializedViewsResponse{}
	if all {
		for nextOffset := 0; ; {
			resp, err := client.ListMaterializedViews(limit, nextOffset)
			if err != nil {
				return err
			}
			result.MaterializedViews = append(result.MaterializedViews, resp.MaterializedViews...)
			if resp.NextOffset <= 0 {
				break
			}
			nextOffset = int(resp.NextOffset)
		}
	} else {
		resp, err := client.ListMaterializedViews(limit, offset)
		if err != nil {
			return err
		}
		result.MaterializedViews = resp.MaterializedViews
		result.NextOffset = resp.NextOffset
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, result)
	default:
		if len(result.MaterializedViews) == 0 {
			fmt.Fprintln(w, "No materialized views found.")
			return nil
		}
		rows := make([][]string, 0, len(result.MaterializedViews))
		for _, mv := range result.MaterializedViews {
			private := "no"
			if mv.IsPrivate {
				private = "yes"
			}
			rows = append(rows, []string{
				mv.SQLID,
				strconv.Itoa(mv.QueryID),
				private,
				output.FormatBytes(mv.TableSizeBytes),
			})
		}
		output.PrintTable(w, []string{"NAME", "QUERY ID", "PRIVATE", "SIZE"}, rows)
		if result.NextOffset > 0 {
			fmt.Fprintf(w, "\nMore results available. Next: --offset %d (or use --all)\n", result.NextOffset)
		}
		return nil
	}
}
