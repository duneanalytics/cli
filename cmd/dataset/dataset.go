package dataset

import "github.com/spf13/cobra"

// NewDatasetCmd returns the `dataset` parent command.
func NewDatasetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Search and discover Dune datasets",
		Long: "Search the Dune dataset catalog to discover tables and their schemas.\n\n" +
			"Use 'dataset search' for keyword-based discovery across all tables, or\n" +
			"'dataset search-by-contract' to find decoded tables for a specific contract address.",
	}
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newSearchByContractCmd())
	return cmd
}
