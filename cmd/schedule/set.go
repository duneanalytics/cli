package schedule

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Create or replace the schedule of a saved query or dashboard",
		Long: "Set the schedule of a saved query (scheduled execution) or a dashboard (scheduled\n" +
			"refresh of all its queries). Creates the schedule when there is none and replaces\n" +
			"every setting when there is one, so pass the settings you want to keep\n" +
			"('dune schedule list' shows them).\n\n" +
			"Every run is charged credits at --performance; prefer the smallest tier and the\n" +
			"lowest frequency that meet the need. You must own the query or dashboard or be an\n" +
			"editor or admin of the team that owns it. On the free plan a schedule expires\n" +
			"after at most 3 months.\n\n" +
			"Supported cron expressions (5-part, UTC):\n" +
			"  */15 * * * *, */20 * * * *, */30 * * * *   every 15, 20 or 30 minutes\n" +
			"  0 * * * *                                  hourly\n" +
			"  0 */N * * *   (N in 2, 3, 4, 6, 8, 12)     every N hours\n" +
			"  M H * * *                                  daily at H:M\n" +
			"  M H * * D                                  weekly on day D (0 = Sunday) at H:M\n" +
			"where M is 0 or 30 and H is 0-23.\n\n" +
			"Examples:\n" +
			"  dune schedule set --query-id 12345 --cron \"0 6 * * *\" --performance medium\n" +
			"  dune schedule set --dashboard-id 678 --cron \"0 */6 * * *\" --performance small \\\n" +
			"    --expires-at 2027-01-01T00:00:00Z",
		Args: cobra.NoArgs,
		RunE: runSet,
	}
	addTargetFlags(cmd)
	cmd.Flags().String("cron", "", "when to run, as a supported 5-part UTC cron expression (required)")
	cmd.Flags().String("performance", "", `engine tier for each run: "small", "medium", or "large" (required)`)
	cmd.Flags().String("expires-at", "", "RFC 3339 time after which the schedule stops; omit for no expiration")
	_ = cmd.MarkFlagRequired("cron")
	_ = cmd.MarkFlagRequired("performance")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSet(cmd *cobra.Command, _ []string) error {
	args := targetArgs(cmd)
	for k, v := range cmdutil.ToolArgs(cmd, map[string]string{
		"cron":        "cron_expression",
		"performance": "performance",
		"expires-at":  "expires_at",
	}) {
		args[k] = v
	}

	var res struct {
		schedule
		Created bool `json:"created"`
	}
	raw, err := cmdutil.CallTool(cmd, "setSchedule", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	verb := "Replaced"
	if res.Created {
		verb = "Created"
	}
	fmt.Fprintf(w, "%s schedule for %s: %q on %s\n", verb, res.target(), res.CronExpression, res.Performance)
	fmt.Fprintf(w, "Next run:   %s\n", orDash(res.NextExecutionTime))
	fmt.Fprintf(w, "Expires at: %s\n", orDash(res.ExpiresAt))
	return nil
}
