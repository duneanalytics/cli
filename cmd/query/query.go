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
			"  create-and-run - Save a query, run it, optionally chart it, and show the first rows\n" +
			"  search   - Find saved queries by name\n" +
			"  get      - Fetch a saved query's SQL, metadata, and execution state\n" +
			"  update   - Modify a query's SQL, title, description, privacy, tags, or parameters\n" +
			"  archive  - Hide a query from the library (still retrievable by ID)\n" +
			"  run      - Execute a saved query by ID and display results\n" +
			"  run-sql  - Execute raw DuneSQL inline without saving a query\n" +
			"  alert    - Schedule a query and get notified when a run returns rows",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newCreateAndRunCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newArchiveCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newRunSQLCmd())
	cmd.AddCommand(newAlertCmd())
	return cmd
}
