package dataset

import "github.com/spf13/cobra"

// NewDatasetCmd returns the `dataset` parent command.
func NewDatasetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Discover and explore datasets across the Dune catalog",
	}
	cmd.AddCommand(newSearchCmd())
	return cmd
}
