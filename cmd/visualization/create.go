package visualization

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
		Short: "Create a new visualization on an existing query",
		Long: `Create a visualization attached to an existing saved query.

IMPORTANT: The --options flag is required for a working visualization. Without
proper options, the visualization will fail to render. The options format depends
on the visualization type.

Visualization types: chart, table, counter, pivot, cohort, funnel, choropleth,
sankey, sunburst_sequence, word_cloud.

For chart type, set globalSeriesType to: column, line, area, scatter, or pie.

COUNTER (simplest — shows a single number):
  --type counter --options '{
    "counterColName": "<column>",
    "rowNumber": 1,
    "stringDecimal": 0,
    "stringPrefix": "",
    "stringSuffix": "",
    "counterLabel": "My Label",
    "coloredPositiveValues": false,
    "coloredNegativeValues": false
  }'

TABLE (displays query results as a table):
  --type table --options '{
    "itemsPerPage": 25,
    "columns": [
      {"name": "<column>", "title": "Display Name", "type": "normal",
       "alignContent": "left", "isHidden": false}
    ]
  }'

COLUMN/LINE/AREA/SCATTER CHART:
  --type chart --options '{
    "globalSeriesType": "line",
    "sortX": true,
    "legend": {"enabled": true},
    "series": {"stacking": null},
    "xAxis": {"title": {"text": "Date"}},
    "yAxis": [{"title": {"text": "Value"}}],
    "columnMapping": {"<x_column>": "x", "<y_column>": "y"},
    "seriesOptions": {
      "<y_column>": {"type": "line", "yAxis": 0, "zIndex": 0}
    }
  }'

PIE CHART:
  --type chart --options '{
    "globalSeriesType": "pie",
    "sortX": true,
    "showDataLabels": true,
    "columnMapping": {"<category_column>": "x", "<value_column>": "y"},
    "seriesOptions": {
      "<value_column>": {"type": "pie", "yAxis": 0, "zIndex": 0}
    }
  }'

Column names in options must match actual query result columns exactly.

Examples:
  # Counter showing row count
  dune viz create --query-id 12345 --name "Total Count" --type counter \
    --options '{"counterColName":"count","rowNumber":1,"stringDecimal":0}'

  # Line chart of daily volume
  dune viz create --query-id 12345 --name "Daily Volume" --type chart \
    --options '{"globalSeriesType":"line","sortX":true,"columnMapping":{"day":"x","volume":"y"},"seriesOptions":{"volume":{"type":"line","yAxis":0,"zIndex":0}},"xAxis":{"title":{"text":"Day"}},"yAxis":[{"title":{"text":"Volume"}}],"legend":{"enabled":true},"series":{"stacking":null}}'

  # Simple table
  dune viz create --query-id 12345 --name "Results" --type table \
    --options '{"itemsPerPage":25,"columns":[{"name":"address","title":"Address","type":"normal","alignContent":"left","isHidden":false},{"name":"balance","title":"Balance","type":"normal","alignContent":"right","isHidden":false}]}'`,
		RunE: runCreate,
	}

	cmd.Flags().Int("query-id", 0, "ID of the query to attach the visualization to (required)")
	cmd.Flags().String("name", "", "visualization name, max 300 characters (required)")
	cmd.Flags().String("type", "table", "visualization type: chart, table, counter")
	cmd.Flags().String("description", "", "visualization description, max 1000 characters")
	cmd.Flags().String("options", "", `visualization options JSON (required for working visualizations, see --help for format per type)`)
	_ = cmd.MarkFlagRequired("query-id")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("options")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runCreate(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	queryID, _ := cmd.Flags().GetInt("query-id")
	name, _ := cmd.Flags().GetString("name")
	vizType, _ := cmd.Flags().GetString("type")
	description, _ := cmd.Flags().GetString("description")
	optionsStr, _ := cmd.Flags().GetString("options")

	var options map[string]any
	if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
		return fmt.Errorf("invalid --options JSON: %w", err)
	}

	resp, err := client.CreateVisualization(models.CreateVisualizationRequest{
		QueryID:     queryID,
		Name:        name,
		Type:        vizType,
		Description: description,
		Options:     options,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://dune.com/embeds/%d/%d", queryID, resp.ID)

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, map[string]any{
			"id":  resp.ID,
			"url": url,
		})
	default:
		fmt.Fprintf(w, "Created visualization %d on query %d\n%s\n", resp.ID, queryID, url)
		return nil
	}
}
