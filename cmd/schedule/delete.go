package schedule

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Stop a saved query or dashboard from running on a schedule",
		Long: "Delete the schedule of a saved query or dashboard. Alerts that depend on the\n" +
			"query's schedule are deleted with it. Deleting when there is no schedule succeeds\n" +
			"and reports that nothing was deleted.\n\n" +
			"To stop a materialized view refresh, use 'dune matview update --no-schedule'.\n\n" +
			"Examples:\n" +
			"  dune schedule delete --query-id 12345\n" +
			"  dune schedule delete --dashboard-id 678",
		Args: cobra.NoArgs,
		RunE: runDelete,
	}
	addTargetFlags(cmd)
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runDelete(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)
	var (
		res *models.DeleteScheduleResponse
		err error
	)
	if isDashboard, id := dashboardTarget(cmd); isDashboard {
		res, err = client.DeleteDashboardSchedule(id)
	} else {
		res, err = client.DeleteQuerySchedule(id)
	}
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintJSON(cmd, res)
	}
	if res.Deleted {
		fmt.Fprintln(w, "Deleted the schedule")
	} else {
		fmt.Fprintln(w, "There was no schedule to delete")
	}
	return nil
}
