package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [dashboard_id]",
		Short: "Get details of a dashboard",
		Long: `Retrieve the full state of a dashboard including metadata and widgets.

Lookup by ID (positional argument):
  dune dashboard get 12345

Lookup by owner and slug:
  dune dashboard get --owner duneanalytics --slug ethereum-overview

Use -o json to get the full response including all widget details.

Examples:
  dune dashboard get 12345 -o json
  dune dashboard get --owner alice --slug my-dashboard -o json`,
		Args: cobra.MaximumNArgs(1),
		RunE: runGet,
	}

	cmd.Flags().String("owner", "", "owner handle (username or team handle)")
	cmd.Flags().String("slug", "", "dashboard URL slug")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	client := cmdutil.ClientFromCmd(cmd)
	owner, _ := cmd.Flags().GetString("owner")
	slug, _ := cmd.Flags().GetString("slug")

	hasID := len(args) > 0
	hasSlug := owner != "" && slug != ""

	if !hasID && !hasSlug {
		return fmt.Errorf("provide either a dashboard ID or both --owner and --slug")
	}
	if hasID && hasSlug {
		return fmt.Errorf("provide either a dashboard ID or --owner/--slug, not both")
	}

	var resp *models.DashboardResponse
	var err error

	if hasID {
		dashboardID, parseErr := strconv.Atoi(args[0])
		if parseErr != nil {
			return fmt.Errorf("invalid dashboard ID %q: must be an integer", args[0])
		}
		resp, err = client.GetDashboard(dashboardID)
	} else {
		resp, err = client.GetDashboardBySlug(owner, slug)
	}
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		tags := ""
		if len(resp.Tags) > 0 {
			tags = strings.Join(resp.Tags, ", ")
		}
		output.PrintTable(w,
			[]string{"Field", "Value"},
			[][]string{
				{"ID", fmt.Sprintf("%d", resp.DashboardID)},
				{"Name", resp.Name},
				{"Slug", resp.Slug},
				{"Private", fmt.Sprintf("%t", resp.IsPrivate)},
				{"Tags", tags},
				{"URL", resp.DashboardURL},
				{"Visualizations", fmt.Sprintf("%d", len(resp.VisualizationWidgets))},
				{"Text Widgets", fmt.Sprintf("%d", len(resp.TextWidgets))},
			},
		)
		return nil
	}
}
