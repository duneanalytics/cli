package matview

import "github.com/spf13/cobra"

// NewMatviewCmd returns the `matview` parent command.
func NewMatviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "matview",
		Aliases: []string{"mv"},
		Short:   "Create, inspect, refresh, and delete Dune materialized views",
		Long: "Manage materialized views (matviews) — query results persisted into a queryable\n" +
			"table, optionally refreshed on a schedule.\n\n" +
			"Subcommands:\n" +
			"  create   - Materialize a saved query into a table (optionally scheduled)\n" +
			"  get      - Fetch a matview's metadata, size, and refresh schedule\n" +
			"  list     - List the materialized views you own\n" +
			"  update   - Change a matview's settings or refresh schedule\n" +
			"  refresh  - Trigger an on-demand refresh\n" +
			"  delete   - Permanently delete a matview and its schedule\n\n" +
			"A matview is referenced by its fully-qualified SQL name (e.g.\n" +
			"dune.my_team.result_token_summary) for get, update, refresh, and delete.",
	}
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newRefreshCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}
