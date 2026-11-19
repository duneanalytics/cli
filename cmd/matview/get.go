package matview

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name-or-id>",
		Short: "Fetch a materialized view by its fully-qualified SQL name or its ID",
		Long: "Retrieve a matview's metadata: source query, privacy, table size, recent execution\n" +
			"IDs, and refresh schedule (if any). Only matviews visible to you are returned.\n\n" +
			"Pass the fully-qualified name, e.g. dune.my_team.result_token_summary, or the matview\n" +
			"ID that 'dune matview list -o json' returns as id.\n\n" +
			"Examples:\n" +
			"  dune matview get dune.my_team.result_token_summary\n" +
			"  dune matview get 01M7T4DTEJ2A3S86NKAMXV50D8\n" +
			"  dune mv get dune.my_team.result_token_summary -o json",
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !strings.Contains(name, ".") {
		// A name is always catalog.schema.table, so this is a matview ID. The REST API finds
		// matviews only by name, so look the name up on the MCP server.
		var mv struct {
			SQLNames []string `json:"sql_names"`
		}
		if _, err := cmdutil.CallTool(cmd, "getMaterializedView", map[string]any{"matview_id": name}, &mv); err != nil {
			return err
		}
		if len(mv.SQLNames) == 0 {
			return fmt.Errorf("materialized view %s has no SQL name", name)
		}
		name = mv.SQLNames[0]
	}

	client := cmdutil.ClientFromCmd(cmd)
	resp, err := client.GetMaterializedView(name)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Name:        %s\n", resp.SQLID)
		fmt.Fprintf(w, "Query ID:    %d\n", resp.QueryID)
		fmt.Fprintf(w, "Private:     %t\n", resp.IsPrivate)
		fmt.Fprintf(w, "Table size:  %s\n", output.FormatBytes(resp.TableSizeBytes))
		if len(resp.LastExecutionIDs) > 0 {
			fmt.Fprintf(w, "Executions:  %s\n", strings.Join(resp.LastExecutionIDs, ", "))
		}
		fmt.Fprintln(w)
		if resp.Schedule == nil {
			fmt.Fprintln(w, "Refresh schedule: none")
			return nil
		}
		fmt.Fprintln(w, "Refresh schedule:")
		fmt.Fprintf(w, "  Cron:        %s\n", resp.Schedule.CronExpression)
		if resp.Schedule.Performance != "" {
			fmt.Fprintf(w, "  Performance: %s\n", resp.Schedule.Performance)
		}
		if resp.Schedule.NextExecutionTime != nil {
			fmt.Fprintf(w, "  Next run:    %s\n", *resp.Schedule.NextExecutionTime)
		}
		if resp.Schedule.ExpiresAt != nil {
			fmt.Fprintf(w, "  Expires at:  %s\n", *resp.Schedule.ExpiresAt)
		}
		return nil
	}
}
