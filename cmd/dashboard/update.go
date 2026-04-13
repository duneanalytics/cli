package dashboard

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <dashboard_id>",
		Short: "Update an existing dashboard",
		Long: `Update dashboard metadata or replace widgets.

Only supply the flags you want to change; omitted fields are preserved.
At least one flag must be provided.

IMPORTANT: Widget updates use all-or-nothing replacement. When you provide
--visualization-widgets or --text-widgets, ALL existing widgets are replaced.
To preserve widgets you want to keep, first fetch the current state with
'dune dashboard get <id> -o json', modify what you need, and pass the complete
widget state back.

The --visualization-widgets flag accepts a JSON array:
  --visualization-widgets '[{"visualization_id":111},{"visualization_id":222}]'

Each widget can include an optional position from the get output:
  --visualization-widgets '[{"visualization_id":111,"position":{"row":0,"col":0,"size_x":3,"size_y":8}}]'

Examples:
  dune dashboard update 12345 --name "New Name" -o json
  dune dashboard update 12345 --tags blockchain,defi,ethereum -o json
  dune dashboard update 12345 --private -o json
  dune dashboard update 12345 --visualization-widgets '[{"visualization_id":111}]' -o json`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().String("name", "", "new dashboard name")
	cmd.Flags().String("slug", "", "new URL slug")
	cmd.Flags().Bool("private", false, "set dashboard privacy")
	cmd.Flags().StringSlice("tags", nil, "replace all tags (comma-separated)")
	cmd.Flags().String("visualization-widgets", "", "visualization widgets JSON array (replaces all)")
	cmd.Flags().String("text-widgets", "", "text widgets JSON array (replaces all)")
	cmd.Flags().String("param-widgets", "", "param widgets JSON array (replaces all, from get output)")
	cmd.Flags().Int32("columns-per-row", 2, "visualizations per row: 1, 2, or 3")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid dashboard ID %q: must be an integer", args[0])
	}

	changedFlags := []string{"name", "slug", "private", "tags", "visualization-widgets", "text-widgets", "param-widgets", "columns-per-row"}
	hasChange := false
	for _, f := range changedFlags {
		if cmd.Flags().Changed(f) {
			hasChange = true
			break
		}
	}
	if !hasChange {
		return fmt.Errorf("at least one flag must be provided (--name, --slug, --private, --tags, --visualization-widgets, --text-widgets, --param-widgets, or --columns-per-row)")
	}

	client := cmdutil.ClientFromCmd(cmd)

	var req models.UpdateDashboardRequest

	if cmd.Flags().Changed("name") {
		v, _ := cmd.Flags().GetString("name")
		req.Name = &v
	}
	if cmd.Flags().Changed("slug") {
		v, _ := cmd.Flags().GetString("slug")
		req.Slug = &v
	}
	if cmd.Flags().Changed("private") {
		v, _ := cmd.Flags().GetBool("private")
		req.IsPrivate = &v
	}
	if cmd.Flags().Changed("tags") {
		v, _ := cmd.Flags().GetStringSlice("tags")
		// Normalize: trim whitespace from each tag
		for i := range v {
			v[i] = strings.TrimSpace(v[i])
		}
		req.Tags = &v
	}
	if cmd.Flags().Changed("visualization-widgets") {
		v, _ := cmd.Flags().GetString("visualization-widgets")
		var widgets []models.VisualizationWidgetInput
		if err := json.Unmarshal([]byte(v), &widgets); err != nil {
			return fmt.Errorf("invalid --visualization-widgets JSON: %w", err)
		}
		req.VisualizationWidgets = &widgets
	}
	if cmd.Flags().Changed("text-widgets") {
		v, _ := cmd.Flags().GetString("text-widgets")
		var widgets []models.TextWidgetInput
		if err := json.Unmarshal([]byte(v), &widgets); err != nil {
			return fmt.Errorf("invalid --text-widgets JSON: %w", err)
		}
		req.TextWidgets = &widgets
	}
	if cmd.Flags().Changed("param-widgets") {
		v, _ := cmd.Flags().GetString("param-widgets")
		var widgets []models.ParamWidgetInput
		if err := json.Unmarshal([]byte(v), &widgets); err != nil {
			return fmt.Errorf("invalid --param-widgets JSON: %w", err)
		}
		req.ParamWidgets = &widgets
	}
	if cmd.Flags().Changed("columns-per-row") {
		v, _ := cmd.Flags().GetInt32("columns-per-row")
		req.ColumnsPerRow = &v
	}

	resp, err := client.UpdateDashboard(dashboardID, req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Updated dashboard %d\n", resp.DashboardID)
		return nil
	}
}
