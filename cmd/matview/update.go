package matview

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a materialized view's settings or refresh schedule",
		Long: "Change a matview's privacy, performance tier, or refresh schedule. Updating\n" +
			"re-executes the source query. The name and source query are immutable.\n" +
			"At least one flag must be provided.\n\n" +
			"Settings you don't pass are preserved (read-modify-write). In particular, the\n" +
			"existing refresh schedule is kept unless you change it with --cron or remove it\n" +
			"with --no-schedule — so a plain `update --private=false` will not silently drop\n" +
			"a schedule.\n\n" +
			"The name must be fully qualified, e.g. dune.my_team.result_token_summary.\n\n" +
			"Examples:\n" +
			"  dune matview update dune.my_team.result_x --performance large\n" +
			"  dune matview update dune.my_team.result_x --cron \"0 0 * * *\"\n" +
			"  dune matview update dune.my_team.result_x --no-schedule\n" +
			"  dune matview update dune.my_team.result_x --private=false",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().Bool("private", false, "set the matview's privacy")
	cmd.Flags().String("performance", "", "execution tier: \"small\", \"medium\", or \"large\"")
	cmd.Flags().String("cron", "", "set or replace the refresh schedule (5-field cron, min 15-minute interval)")
	cmd.Flags().Bool("no-schedule", false, "remove the refresh schedule")
	cmd.Flags().String("expires-at", "", "RFC3339 time when the refresh schedule expires")
	cmd.MarkFlagsMutuallyExclusive("cron", "no-schedule")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fqName := args[0]
	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("private") && !cmd.Flags().Changed("performance") &&
		!cmd.Flags().Changed("cron") && !cmd.Flags().Changed("no-schedule") &&
		!cmd.Flags().Changed("expires-at") {
		return fmt.Errorf("at least one flag must be provided (--private, --performance, --cron, --no-schedule, or --expires-at); use 'matview refresh' to rebuild without changing settings")
	}

	client := cmdutil.ClientFromCmd(cmd)

	// Read-modify-write: fetch the current state so unspecified settings — especially the
	// refresh schedule — are preserved. The upsert endpoint overwrites everything, and omitting
	// the cron expression would silently drop an existing schedule.
	current, err := client.GetMaterializedView(fqName)
	if err != nil {
		return err
	}

	req := models.UpsertMaterializedViewRequest{
		Name:      bareName(current.SQLID),
		QueryID:   current.QueryID,
		IsPrivate: current.IsPrivate,
	}
	if cmd.Flags().Changed("private") {
		req.IsPrivate, _ = cmd.Flags().GetBool("private")
	}

	// Resolve the schedule, defaulting to preserving the current one.
	noSchedule, _ := cmd.Flags().GetBool("no-schedule")
	switch {
	case noSchedule:
		// leave CronExpression nil → removes the schedule
	case cmd.Flags().Changed("cron"):
		cron, _ := cmd.Flags().GetString("cron")
		req.CronExpression = &cron
	case current.Schedule != nil:
		cron := current.Schedule.CronExpression
		req.CronExpression = &cron
	}

	if cmd.Flags().Changed("expires-at") && req.CronExpression == nil {
		return fmt.Errorf("--expires-at requires a refresh schedule (set one with --cron)")
	}

	// Performance: an explicit flag wins; otherwise keep the scheduled tier if there is one.
	switch {
	case performance != "":
		req.Performance = performance
	case current.Schedule != nil:
		req.Performance = current.Schedule.Performance
	}

	// Expiry: an explicit flag wins; otherwise preserve the current expiry while keeping a schedule.
	switch {
	case req.CronExpression == nil:
		// no schedule → no expiry
	case cmd.Flags().Changed("expires-at"):
		expiresAt, _ := cmd.Flags().GetString("expires-at")
		req.ExpiresAt = &expiresAt
	case current.Schedule != nil:
		req.ExpiresAt = current.Schedule.ExpiresAt
	}

	resp, err := client.UpsertMaterializedView(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Updated materialized view %s\n", resp.SQLID)
		fmt.Fprintf(w, "Execution ID: %s\n", resp.ExecutionID)
		return nil
	}
}
