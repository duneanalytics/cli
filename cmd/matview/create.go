package matview

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Materialize a saved query into a table",
		Long: "Create a materialized view from an existing saved query. The query's results are\n" +
			"written into a queryable table and an immediate execution is triggered.\n\n" +
			"Requirements:\n" +
			"  - The source query must be saved (non-temporary) and have no parameters.\n" +
			"  - The query must be owned by the same user/team as the matview.\n" +
			"  - The name must start with \"result_\", be 8-128 lowercase alphanumeric/underscore\n" +
			"    characters, and not end with an underscore. It is immutable after creation.\n\n" +
			"The matview is owned by your authenticated context: a team API key creates a\n" +
			"team-owned matview, a personal key a personal one.\n\n" +
			"If a matview with this name already exists for the same query it is replaced\n" +
			"(re-run). To change settings on an existing matview, use `dune matview update`.\n\n" +
			"With --cron the matview is refreshed periodically (5-field cron, minimum 15-minute\n" +
			"interval).\n\n" +
			"Examples:\n" +
			"  dune matview create --name result_token_summary --query-id 12345 --performance medium\n" +
			"  dune matview create --name result_daily --query-id 12345 --cron \"0 */6 * * *\"\n" +
			"  dune matview create --name result_public --query-id 12345 --private=false",
		RunE: runCreate,
	}

	cmd.Flags().String("name", "", "SQL-safe matview name, e.g. \"result_token_summary\" (required)")
	cmd.Flags().Int("query-id", 0, "ID of the source query to materialize (required)")
	cmd.Flags().Bool("private", true, "make the matview private; private matviews require a supporting plan")
	cmd.Flags().String("performance", "", "execution tier: \"small\", \"medium\", or \"large\" (default: account default)")
	cmd.Flags().String("cron", "", "5-field cron for periodic refresh, min 15-minute interval (default: none)")
	cmd.Flags().String("expires-at", "", "RFC3339 time when the refresh schedule expires (requires --cron)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("query-id")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runCreate(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	if err := validateMatviewName(name); err != nil {
		return err
	}
	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}
	if cmd.Flags().Changed("expires-at") && !cmd.Flags().Changed("cron") {
		return fmt.Errorf("--expires-at requires --cron (expiry applies to the refresh schedule)")
	}

	queryID, _ := cmd.Flags().GetInt("query-id")
	private, _ := cmd.Flags().GetBool("private")

	req := models.UpsertMaterializedViewRequest{
		Name:        name,
		QueryID:     queryID,
		IsPrivate:   private,
		Performance: performance,
	}
	if cmd.Flags().Changed("cron") {
		cron, _ := cmd.Flags().GetString("cron")
		req.CronExpression = &cron
	}
	if cmd.Flags().Changed("expires-at") {
		expiresAt, _ := cmd.Flags().GetString("expires-at")
		req.ExpiresAt = &expiresAt
	}

	client := cmdutil.ClientFromCmd(cmd)
	resp, err := client.UpsertMaterializedView(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Created materialized view %s\n", resp.SQLID)
		fmt.Fprintf(w, "Execution ID: %s\n", resp.ExecutionID)
		fmt.Fprintf(w, "\nThe matview is being built. Check results with:\n")
		fmt.Fprintf(w, "  dune execution results %s\n", resp.ExecutionID)
		return nil
	}
}
