package visualization

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <visualization_id>",
		Short: "Get details of a visualization including type, options, and timestamps",
		Long: "Retrieve detailed information about a visualization by its ID.\n\n" +
			"Returns the visualization name, type, description, options JSON, query ID,\n" +
			"and timestamps. Use this to inspect an existing visualization before\n" +
			"updating it, or to review the options format for a given type.\n\n" +
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

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		output.PrintTable(w,
			[]string{"Field", "Value"},
			[][]string{
				{"ID", fmt.Sprintf("%d", resp.ID)},
				{"Query ID", fmt.Sprintf("%d", resp.QueryID)},
				{"Name", resp.Name},
				{"Description", resp.Description},
				{"Type", resp.Type},
				{"Created At", resp.CreatedAt},
				{"Updated At", resp.UpdatedAt},
			},
		)
		return nil
	}
}
