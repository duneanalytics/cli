package query

import "github.com/spf13/cobra"

// NewQueryCmd returns the `query` parent command.
func NewQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Manage Dune queries",
	}
}
