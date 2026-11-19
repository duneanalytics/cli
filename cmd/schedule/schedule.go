package schedule

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
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

type schedule struct {
	Type              string  `json:"type"`
	QueryID           *int64  `json:"query_id"`
	DashboardID       *int64  `json:"dashboard_id"`
	CronExpression    string  `json:"cron_expression"`
	Performance       string  `json:"performance"`
	NextExecutionTime *string `json:"next_execution_time"`
	ExpiresAt         *string `json:"expires_at"`
}

// target describes what a schedule runs, e.g. "query 123".
func (s schedule) target() string {
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

func targetArgs(cmd *cobra.Command) map[string]any {
	return cmdutil.ToolArgs(cmd, map[string]string{"query-id": "query_id", "dashboard-id": "dashboard_id"})
}

func orDash(s *string) string {
	if s == nil || *s == "" {
		return "-"
	}
	return *s
}
