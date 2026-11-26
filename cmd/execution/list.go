package execution

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

type teamRun struct {
	ExecutionID string `json:"execution_id"`
	State       string `json:"state"`
	Source      string `json:"source"`
	QueryID     *int   `json:"query_id"`
	QueryName   string `json:"query_name"`
	StartedBy   *struct {
		Handle string `json:"handle"`
	} `json:"started_by"`
	APIKeyName      string   `json:"api_key_name"`
	Performance     string   `json:"performance"`
	EndedAt         string   `json:"ended_at"`
	DurationSeconds *float64 `json:"duration_seconds"`
	UsesSlot        bool     `json:"uses_concurrency_slot"`
	CanCancel       bool     `json:"can_cancel"`
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the team's active or recently finished query runs",
		Long: "List the team's query runs from every source: the Dune app, the API, MCP, the\n" +
			"CLI, schedules, materialized views and connectors, with who started each run,\n" +
			"where it came from and its engine size.\n\n" +
			"--status active lists queued and running runs, oldest first, at most 100.\n" +
			"--status finished lists runs that finished in the last hour, newest first, at\n" +
			"most 50, with how each ended and how long it ran.\n\n" +
			"Needs a team API key. Use 'dune execution cancel' to stop an active run.\n\n" +
			"Examples:\n" +
			"  dune execution list\n" +
			"  dune execution list --status finished\n" +
			"  dune execution list -o json",
		Args: cobra.NoArgs,
		RunE: runList,
	}
	cmd.Flags().String("status", "active", `which runs to list: "active" or "finished"`)
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	status, _ := cmd.Flags().GetString("status")
	if status != "active" && status != "finished" {
		return fmt.Errorf(`invalid --status %q: use "active" or "finished"`, status)
	}
	var res struct {
		Executions           []teamRun `json:"executions"`
		TotalCount           *int      `json:"total_count"`
		ConcurrencySlotLimit *int      `json:"concurrency_slot_limit"`
	}
	raw, err := cmdutil.CallTool(cmd, "listTeamExecutions", map[string]any{"status": status}, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}

	if status == "finished" {
		rows := make([][]string, len(res.Executions))
		for i, r := range res.Executions {
			rows[i] = []string{r.ExecutionID, r.State, queryLabel(r), startedBy(r), r.Source, duration(r.DurationSeconds), r.EndedAt}
		}
		output.PrintTable(w, []string{"EXECUTION ID", "OUTCOME", "QUERY", "STARTED BY", "SOURCE", "DURATION", "ENDED AT"}, rows)
		fmt.Fprintf(w, "\n%d runs finished in the last hour\n", len(res.Executions))
		return nil
	}

	rows := make([][]string, len(res.Executions))
	for i, r := range res.Executions {
		rows[i] = []string{r.ExecutionID, r.State, queryLabel(r), startedBy(r), r.Source, r.Performance, yesNo(r.UsesSlot), yesNo(r.CanCancel)}
	}
	output.PrintTable(w, []string{"EXECUTION ID", "STATE", "QUERY", "STARTED BY", "SOURCE", "ENGINE", "HOLDS SLOT", "CAN CANCEL"}, rows)
	total := len(res.Executions)
	if res.TotalCount != nil {
		total = *res.TotalCount
	}
	fmt.Fprintf(w, "\n%d of %d queued and running runs", len(res.Executions), total)
	if res.ConcurrencySlotLimit != nil {
		fmt.Fprintf(w, ", %d concurrency slots", *res.ConcurrencySlotLimit)
	}
	fmt.Fprintln(w)
	return nil
}

func queryLabel(r teamRun) string {
	switch {
	case r.QueryName != "":
		return r.QueryName
	case r.QueryID != nil:
		return strconv.Itoa(*r.QueryID)
	default:
		return "ad-hoc SQL"
	}
}

func startedBy(r teamRun) string {
	switch {
	case r.StartedBy != nil:
		return r.StartedBy.Handle
	case r.APIKeyName != "":
		return "API key " + r.APIKeyName
	default:
		return "-"
	}
}

func duration(seconds *float64) string {
	if seconds == nil {
		return "-"
	}
	return fmt.Sprintf("%.1fs", *seconds)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
