package folder

import (
	"errors"
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newMoveContentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move-content",
		Short: "Move queries and dashboards into a folder or out of folders",
		Long: "Move up to 100 queries and 100 dashboards into a folder, or out of any folder\n" +
			"with --to-root. Everything listed must belong to the same library as the API key\n" +
			"(the team for a team key, you for a personal key). The move is all or nothing.\n\n" +
			"Find IDs with 'dune query search', 'dune dashboard search', or 'dune folder get'.\n\n" +
			"Examples:\n" +
			"  dune folder move-content --query-ids 4512345,4512346 --to dp180Cu3jW3L\n" +
			"  dune folder move-content --dashboard-ids 98765 --to-root",
		Args: cobra.NoArgs,
		RunE: runMoveContent,
	}
	cmd.Flags().Int64Slice("query-ids", nil, "IDs of the queries to move (comma-separated)")
	cmd.Flags().Int64Slice("dashboard-ids", nil, "IDs of the dashboards to move (comma-separated)")
	addDestinationFlags(cmd, "the content")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runMoveContent(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"query-ids":     "query_ids",
		"dashboard-ids": "dashboard_ids",
	})
	if len(args) == 0 {
		return errors.New("provide --query-ids, --dashboard-ids, or both")
	}
	args["target_folder_id"] = destination(cmd)

	var res struct {
		MovedQueryIDs     []int64        `json:"moved_query_ids"`
		MovedDashboardIDs []int64        `json:"moved_dashboard_ids"`
		TargetFolder      *folderDetails `json:"target_folder"`
	}
	raw, err := cmdutil.CallTool(cmd, "moveContentToFolder", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	where := "out of folders"
	if res.TargetFolder != nil {
		where = "to " + res.TargetFolder.Path
	}
	fmt.Fprintf(w, "Moved %d queries and %d dashboards %s\n",
		len(res.MovedQueryIDs), len(res.MovedDashboardIDs), where)
	return nil
}
