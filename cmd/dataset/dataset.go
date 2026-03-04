package dataset

import "github.com/spf13/cobra"

// NewDatasetCmd returns the `dataset` parent command.
func NewDatasetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage Dune datasets",
	}
	cmd.AddCommand(newSearchCmd())
	return cmd
}
