package dashboard

import (
	"encoding/json"
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new dashboard",
		Long: `Create a new dashboard with optional visualization and text widgets.

Visualizations are placed in a grid layout controlled by --columns-per-row:
  1 = full-width charts, 2 = half-width (default), 3 = compact overview.
Text widgets always span the full width above visualizations.

The --visualization-ids flag accepts a comma-separated list of visualization IDs
(from 'dune viz create' or 'dune viz list' output).

The --text-widgets flag accepts a JSON array of text widget objects:
  --text-widgets '[{"text":"# Dashboard Title"},{"text":"Description here"}]'

Examples:
  # Empty dashboard
  dune dashboard create --name "My Dashboard" -o json

  # Dashboard with visualizations
  dune dashboard create --name "DEX Overview" --visualization-ids 111,222,333 -o json

  # Dashboard with text header and visualizations
  dune dashboard create --name "ETH Analysis" \
    --text-widgets '[{"text":"# Ethereum Analysis\nDaily metrics"}]' \
    --visualization-ids 111,222 --columns-per-row 1 -o json

  # Private dashboard
  dune dashboard create --name "Internal Metrics" --private -o json`,
		RunE: runCreate,
	}

	cmd.Flags().String("name", "", "dashboard name (required)")
	cmd.Flags().Bool("private", false, "make the dashboard private")
	cmd.Flags().Int64Slice("visualization-ids", nil, "visualization IDs to add (comma-separated)")
	cmd.Flags().String("text-widgets", "", `text widgets JSON array, e.g. '[{"text":"# Title"}]'`)
	cmd.Flags().Int32("columns-per-row", 2, "visualizations per row: 1, 2, or 3")
	_ = cmd.MarkFlagRequired("name")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runCreate(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	name, _ := cmd.Flags().GetString("name")
	private, _ := cmd.Flags().GetBool("private")
	vizIDs, _ := cmd.Flags().GetInt64Slice("visualization-ids")
	textWidgetsStr, _ := cmd.Flags().GetString("text-widgets")
	columnsPerRow, _ := cmd.Flags().GetInt32("columns-per-row")

	req := models.CreateDashboardRequest{
		Name: name,
	}

	if cmd.Flags().Changed("private") {
		req.IsPrivate = &private
	}
	if len(vizIDs) > 0 {
		req.VisualizationIDs = vizIDs
	}
	if textWidgetsStr != "" {
		var textWidgets []models.TextWidgetInput
		if err := json.Unmarshal([]byte(textWidgetsStr), &textWidgets); err != nil {
			return fmt.Errorf("invalid --text-widgets JSON: %w", err)
		}
		req.TextWidgets = textWidgets
	}
	if cmd.Flags().Changed("columns-per-row") {
		req.ColumnsPerRow = &columnsPerRow
	}

	resp, err := client.CreateDashboard(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Created dashboard %d\n%s\n", resp.DashboardID, resp.DashboardURL)
		return nil
	}
}
