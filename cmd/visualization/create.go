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
		Long: "Create a visualization attached to an existing saved query.\n\n" +
			"The visualization type must be one of: chart, table, counter, pivot,\n" +
			"cohort, funnel, choropleth, sankey, sunburst_sequence, word_cloud.\n\n" +
			"The --options flag accepts a JSON string of visualization-specific\n" +
			"configuration (axes, series, formatting, etc.).\n\n" +
			"Examples:\n" +
			"  dune viz create --query-id 12345 --name \"Token Volume\" --type chart\n" +
			"  dune viz create --query-id 12345 --name \"Summary\" --type counter --options '{\"column\":\"total\",\"row_num\":1}'\n" +
			"  dune viz create --query-id 12345 --name \"Results\" --type table -o json",
		RunE: runCreate,
	}

	cmd.Flags().Int("query-id", 0, "ID of the query to attach the visualization to (required)")
	cmd.Flags().String("name", "", "visualization name, max 300 characters (required)")
	cmd.Flags().String("type", "table", "visualization type: chart, table, counter, pivot, cohort, funnel, choropleth, sankey, sunburst_sequence, word_cloud")
	cmd.Flags().String("description", "", "visualization description, max 1000 characters")
	cmd.Flags().String("options", "{}", "JSON string of visualization options")
	_ = cmd.MarkFlagRequired("query-id")
	_ = cmd.MarkFlagRequired("name")
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

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Created visualization %d on query %d\n", resp.ID, queryID)
		return nil
	}
}
