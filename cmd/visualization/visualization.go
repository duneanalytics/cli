package visualization

import "github.com/spf13/cobra"

func NewVisualizationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "visualization",
		Aliases: []string{"viz"},
		Short:   "Create and manage Dune visualizations",
		Long: "Create and manage visualizations on Dune queries.\n\n" +
			"Visualizations are charts, tables, counters, and other visual representations\n" +
			"of query results. Each visualization is attached to a saved query by its query ID.\n\n" +
			"Important: Visualizations require a saved query ID (from 'dune query create').\n" +
			"'dune query run-sql' does not create a saved query and cannot be used with\n" +
			"visualizations.\n\n" +
			"Supported types: chart, table, counter, pivot, cohort, funnel, choropleth,\n" +
			"sankey, sunburst_sequence, word_cloud.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}
