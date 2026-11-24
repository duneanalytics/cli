package dashboard

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find dashboards by name",
		Long: "Find dashboards whose name contains the given text (not case-sensitive), to get a\n" +
			"dashboard ID, or the owner and slug for 'dune dashboard get'. Results include public\n" +
			"dashboards and the private ones you or your team can access; --owned-only keeps only\n" +
			"the ones the API key's user or team owns.\n\n" +
			"Examples:\n" +
			"  dune dashboard search --query dex\n" +
			"  dune dashboard search --query \"stablecoin supply\" --owned-only -o json",
		Args: cobra.NoArgs,
		RunE: runSearch,
	}
	cmd.Flags().String("query", "", "text to find in the dashboard name (required)")
	cmd.Flags().Bool("owned-only", false, "only return dashboards owned by the API key's user or team")
	cmd.Flags().Int("limit", 20, "maximum number of dashboards to return (1-50)")
	_ = cmd.MarkFlagRequired("query")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSearch(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{"query": "search_string", "limit": "limit"})
	if owned, _ := cmd.Flags().GetBool("owned-only"); owned {
		args["scope"] = "owned_only"
	}
	var res struct {
		Results []struct {
			DashboardID int    `json:"dashboard_id"`
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Owner       struct {
				Handle string `json:"handle"`
			} `json:"owner"`
		} `json:"results"`
	}
	raw, err := cmdutil.CallTool(cmd, "searchDuneDashboards", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(cmd, raw)
	}
	if len(res.Results) == 0 {
		fmt.Fprintln(w, "No dashboards found.")
		return nil
	}
	rows := make([][]string, len(res.Results))
	for i, r := range res.Results {
		rows[i] = []string{fmt.Sprintf("%d", r.DashboardID), r.Name, r.Owner.Handle, r.Slug}
	}
	output.PrintTable(w, []string{"DASHBOARD_ID", "NAME", "OWNER", "SLUG"}, rows)
	return nil
}
