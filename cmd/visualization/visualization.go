package visualization

import "github.com/spf13/cobra"

func NewVisualizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "visualization",
		Aliases: []string{"viz"},
		Short:   "Create and manage Dune visualizations",
		Long: "Create and manage visualizations on Dune queries.\n\n" +
			"Visualizations are charts, tables, counters, and other visual representations\n" +
			"of query results. Each visualization is attached to a saved query.",
	}

	cmd.AddCommand(newCreateCmd())

	return cmd
}
