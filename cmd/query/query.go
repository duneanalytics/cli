package query

import "github.com/spf13/cobra"

// NewQueryCmd returns the `query` parent command.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Create, retrieve, update, execute, and archive Dune queries",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newRunSQLCmd())
	return cmd
}
