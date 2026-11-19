package schedule

import (
	"fmt"
	"time"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// NewScheduleCmd returns the `schedule` parent command.
func NewScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Run saved queries and refresh dashboards on a schedule",
		Long: "List, set, and delete the schedules of saved queries and dashboards.\n\n" +
			"A query or a dashboard has at most one schedule, which belongs to the owner of the\n" +
			"query or dashboard. Every scheduled run is charged credits at its performance tier.\n" +
			"Materialized view refreshes are scheduled with 'dune matview create/update --cron'.\n\n" +
			"Subcommands:\n" +
			"  list    - List the schedules you or your team own\n" +
			"  set     - Create or replace the schedule of a query or dashboard\n" +
			"  delete  - Stop a query or dashboard from running on a schedule",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

// target describes what a schedule runs, e.g. "query 123".
func target(s models.Schedule) string {
	switch {
	case s.DashboardID != nil:
		return fmt.Sprintf("dashboard %d", *s.DashboardID)
	case s.QueryID != nil:
		return fmt.Sprintf("query %d", *s.QueryID)
	}
	return s.Type
}

// addTargetFlags registers --query-id and --dashboard-id, exactly one of which is required.
func addTargetFlags(cmd *cobra.Command) {
	cmd.Flags().Int("query-id", 0, "ID of the saved query whose schedule to change")
	cmd.Flags().Int("dashboard-id", 0, "ID of the dashboard whose scheduled refresh to change")
	cmd.MarkFlagsMutuallyExclusive("query-id", "dashboard-id")
	cmd.MarkFlagsOneRequired("query-id", "dashboard-id")
}

// dashboardTarget reports whether the command targets a dashboard rather than a
// query, and the ID of the target.
func dashboardTarget(cmd *cobra.Command) (bool, int) {
	if cmd.Flags().Changed("dashboard-id") {
		id, _ := cmd.Flags().GetInt("dashboard-id")
		return true, id
	}
	id, _ := cmd.Flags().GetInt("query-id")
	return false, id
}

func orDash(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(time.RFC3339)
}
