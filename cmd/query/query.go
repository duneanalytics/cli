package query

import "github.com/spf13/cobra"

// NewQueryCmd returns the `query` parent command.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Manage Dune queries",
		Long: "Create, retrieve, update, archive, and execute DuneSQL queries.\n\n" +
			"Use 'query create' to save a reusable query, 'query run' to execute a saved query,\n" +
			"or 'query run-sql' to execute raw DuneSQL without saving.",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newRunSQLCmd())
	return cmd
}
