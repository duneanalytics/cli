package query

import "github.com/spf13/cobra"

// NewQueryCmd returns the `query` parent command.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Manage Dune queries",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newRunCmd())
	return cmd
}
