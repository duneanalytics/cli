package dashboard

import "github.com/spf13/cobra"

func NewDashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"dash"},
		Short:   "Create and manage Dune dashboards",
		Long: "Create and manage dashboards on Dune.\n\n" +
			"Dashboards are collections of visualizations and text widgets that display\n" +
			"blockchain and crypto data. Each dashboard has a unique URL and can be\n" +
			"public or private.\n\n" +
			"Visualizations are arranged in a 6-column grid. Use --columns-per-row to\n" +
			"control layout: 1 for full-width, 2 for half-width (default), 3 for compact.",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())

	return cmd
}
