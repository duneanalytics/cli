package visualization

import (
	"encoding/json"
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
			"  dune viz update 12345 --type counter --options '{\"counterColName\":\"count\"}'",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().String("name", "", "new name for the visualization")
	cmd.Flags().String("type", "", "new visualization type")
	cmd.Flags().String("description", "", "new description for the visualization")
	cmd.Flags().String("options", "", "new visualization options JSON")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	vizID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid visualization ID %q: must be an integer", args[0])
	}

	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("type") &&
		!cmd.Flags().Changed("description") && !cmd.Flags().Changed("options") {
		return fmt.Errorf("at least one flag must be provided (--name, --type, --description, or --options)")
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
	if cmd.Flags().Changed("options") {
		optionsStr, _ := cmd.Flags().GetString("options")
		var options map[string]any
		if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
			return fmt.Errorf("invalid --options JSON: %w", err)
		}
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
