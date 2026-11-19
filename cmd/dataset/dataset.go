package dataset

import "github.com/spf13/cobra"

// NewDatasetCmd returns the `dataset` parent command.
func NewDatasetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Search and discover Dune datasets",
		Long: "Search the Dune dataset catalog to discover tables and their schemas.\n\n" +
			"Use 'dataset search' for keyword-based discovery across all tables,\n" +
			"'dataset search-by-contract' to find decoded tables for a specific contract address,\n" +
			"'dataset blockchains' to list chain names, and 'dataset size' to see how large a table is.",
	}
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newSearchByContractCmd())
	cmd.AddCommand(newBlockchainsCmd())
	cmd.AddCommand(newSizeCmd())
	return cmd
}
