package visualization

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <visualization_id>",
		Short: "Update an existing visualization",
		Long: "Replace a visualization's properties. Fetches the current visualization first,\n" +
			"applies your changes, and sends the full updated object.\n\n" +
			"Only supply the flags you want to change; unchanged fields are preserved.\n" +
			"At least one flag must be provided.\n\n" +
			"Examples:\n" +
			"  dune viz update 12345 --name \"New Name\"\n" +
			"  dune viz update 12345 --type counter --options '{\"counterColName\":\"count\"}'\n" +
			"  dune viz update 12345 --chart-type counter --config @counter.json\n" +
			"  dune viz update 12345 --counter-label \"Total volume\"\n\n" +
			"--counter-label sets the caption shown with a counter's number; an empty\n" +
			"string removes it. It only applies to counter visualizations.\n\n" +
			"--chart-type and --config replace --type and --options with a compact chart\n" +
			"description. When the chart type stays the same, options the config does not\n" +
			"cover (such as custom colors) are kept.\n\n" + ChartConfigHelp,
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().String("name", "", "new name for the visualization")
	cmd.Flags().String("type", "", "new visualization type")
	cmd.Flags().String("description", "", "new description for the visualization")
	cmd.Flags().String("options", "", "new raw visualization options JSON")
	cmd.Flags().String("counter-label", "", "new caption for a counter; an empty string removes it")
	AddChartFlags(cmd)
	cmd.MarkFlagsMutuallyExclusive("options", "config")
	cmd.MarkFlagsMutuallyExclusive("type", "chart-type")
	cmd.MarkFlagsMutuallyExclusive("counter-label", "options")
	cmd.MarkFlagsMutuallyExclusive("counter-label", "type")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	vizID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid visualization ID %q: must be an integer", args[0])
	}

	chart, err := ChartArgs(cmd)
	if err != nil {
		return err
	}
	if chart != nil || cmd.Flags().Changed("counter-label") {
		if chart == nil {
			chart = map[string]any{}
		}
		chart["visualizationId"] = vizID
		for flag, arg := range map[string]string{
			"name": "visualizationName", "description": "description", "counter-label": "counterLabel",
		} {
			if cmd.Flags().Changed(flag) {
				chart[arg], _ = cmd.Flags().GetString(flag)
			}
		}
		return runChartTool(cmd, "updateVisualization", chart, "Updated")
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("type") &&
		!cmd.Flags().Changed("description") && !cmd.Flags().Changed("options") {
		return fmt.Errorf("at least one flag must be provided " +
			"(--name, --type, --description, --options, --counter-label, or --chart-type with --config)")
	}

	// Parsed before any call, so bad options fail without touching the visualization.
	var options map[string]any
	if cmd.Flags().Changed("options") {
		if err := cmdutil.JSONFlag(cmd, "options", &options); err != nil {
			return err
		}
		// "null" decodes to a nil map, which would replace the current options with null.
		if options == nil {
			return fmt.Errorf("--options must be a JSON object")
		}
	}

	client := cmdutil.ClientFromCmd(cmd)

	// Fetch current visualization to preserve unchanged fields
	current, err := client.GetVisualization(vizID)
	if err != nil {
		return fmt.Errorf("failed to fetch current visualization: %w", err)
	}

	// Start with current values, override with changed flags
	req := models.UpdateVisualizationRequest{
		Name:        current.Name,
		Type:        current.Type,
		Description: current.Description,
		Options:     current.Options,
	}

	if cmd.Flags().Changed("name") {
		req.Name, _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("type") {
		req.Type, _ = cmd.Flags().GetString("type")
	}
	if cmd.Flags().Changed("description") {
		req.Description, _ = cmd.Flags().GetString("description")
	}
	if options != nil {
		req.Options = options
	}

	resp, err := client.UpdateVisualization(vizID, req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Updated visualization %d\n", resp.ID)
		return nil
	}
}
