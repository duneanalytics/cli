package schedule

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the schedules you or your team own",
		Long: "List the schedules owned by the authenticated user or team: scheduled query\n" +
			"executions, scheduled dashboard refreshes, and materialized view refreshes.\n" +
			"A personal API key lists personal schedules; a team key lists the team's.\n\n" +
			"Examples:\n" +
			"  dune schedule list\n" +
			"  dune schedule list --limit 100 --offset 100 -o json",
		Args: cobra.NoArgs,
		RunE: runList,
	}
	cmd.Flags().Int("limit", 20, "number of schedules to return (1-100)")
	cmd.Flags().Int("offset", 0, "number of schedules to skip, for the next page")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{"limit": "limit", "offset": "offset"})
	var res struct {
		Schedules []schedule `json:"schedules"`
		Total     int        `json:"total"`
	}
	raw, err := cmdutil.CallTool(cmd, "listSchedules", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	rows := make([][]string, len(res.Schedules))
	for i, s := range res.Schedules {
		rows[i] = []string{
			s.Type, s.target(), s.CronExpression, s.Performance, orDash(s.NextExecutionTime), orDash(s.ExpiresAt),
		}
	}
	output.PrintTable(w, []string{"TYPE", "TARGET", "CRON", "PERFORMANCE", "NEXT_RUN", "EXPIRES_AT"}, rows)
	fmt.Fprintf(w, "\n%d of %d schedules\n", len(res.Schedules), res.Total)
	return nil
}
