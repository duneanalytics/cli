package query

import "github.com/spf13/cobra"

// NewQueryCmd returns the `query` parent command.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Create, retrieve, update, execute, and archive Dune queries",
		Long: "Create, retrieve, update, archive, and execute DuneSQL queries.\n\n" +
			"Subcommands:\n" +
			"  create   - Save a new reusable DuneSQL query and get its query ID\n" +
			"  get      - Fetch a saved query's SQL, metadata, and execution state\n" +
			"  update   - Modify a query's SQL, title, description, privacy, or tags\n" +
			"  archive  - Hide a query from the library (still retrievable by ID)\n" +
			"  run      - Execute a saved query by ID and display results\n" +
			"  run-sql  - Execute raw DuneSQL inline without saving a query",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newRunSQLCmd())
	return cmd
}
