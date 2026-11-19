package visualization

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <visualization_id>",
		Short: "Get details of a visualization including type, options, and timestamps",
		Long: "Retrieve detailed information about a visualization by its ID.\n\n" +
			"Returns the visualization name, type, chart type, description, options JSON,\n" +
			"query ID, timestamps, and links to the interactive chart and its image. Use this\n" +
			"to inspect an existing visualization before updating it, or to review the options\n" +
			"format for a given type.\n\n" +
			"For a private or temporary query, the chart links carry a read-only API key for\n" +
			"that query, which Dune creates the first time it is needed: do not share them.\n\n" +
			"Use -o json to get the full options object for programmatic use.\n\n" +
			"Examples:\n" +
			"  dune viz get 12345\n" +
			"  dune viz get 12345 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	vizID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid visualization ID %q: must be an integer", args[0])
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.GetVisualization(vizID)
	if err != nil {
		return err
	}
	// The REST API does not return the chart type or links; the MCP tool getVisualization does.
	var extra struct {
		ChartType           *string `json:"chartType"`
		IsPrivate           bool    `json:"isPrivate"`
		InteractiveChartURL string  `json:"interactiveChartUrl"`
		ChartJpegURL        string  `json:"chartJpegUrl"`
	}
	if _, err := cmdutil.CallTool(cmd, "getVisualization", map[string]any{"visualizationId": vizID}, &extra); err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, struct {
			*models.GetVisualizationResponse
			ChartType           *string `json:"chart_type"`
			IsPrivate           bool    `json:"is_private"`
			InteractiveChartURL string  `json:"interactive_chart_url"`
			ChartJpegURL        string  `json:"chart_jpeg_url"`
		}{resp, extra.ChartType, extra.IsPrivate, extra.InteractiveChartURL, extra.ChartJpegURL})
	default:
		chartType := "-"
		if extra.ChartType != nil {
			chartType = *extra.ChartType
		}
		output.PrintTable(w,
			[]string{"Field", "Value"},
			[][]string{
				{"ID", fmt.Sprintf("%d", resp.ID)},
				{"Query ID", fmt.Sprintf("%d", resp.QueryID)},
				{"Name", resp.Name},
				{"Description", resp.Description},
				{"Type", resp.Type},
				{"Chart Type", chartType},
				{"Private", strconv.FormatBool(extra.IsPrivate)},
				{"Created At", resp.CreatedAt},
				{"Updated At", resp.UpdatedAt},
				{"Chart URL", extra.InteractiveChartURL},
				{"Image URL", extra.ChartJpegURL},
			},
		)
		return nil
	}
}
