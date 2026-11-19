package query

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <query-id>",
		Short: "Fetch a saved Dune query by ID, including SQL and metadata",
		Long: "Retrieve SQL, metadata, owner, privacy flags, tags, and execution state for\n" +
			"an existing Dune query. Only queries visible to the authenticated user are returned.\n\n" +
			"The latest execution ID reads its results without running the query again:\n" +
			"  dune execution results <latest-execution-id>",
		Args: cobra.ExactArgs(1),
		RunE: runGet,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	queryID, err := parseQueryID(args[0])
	if err != nil {
		return err
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.GetQuery(queryID)
	if err != nil {
		return err
	}
	// The REST API does not report executions; the MCP tool getDuneQuery does.
	var latest struct {
		ID    *string `json:"latest_execution_id"`
		State *string `json:"latest_execution_state"`
	}
	if _, err := cmdutil.CallTool(cmd, "getDuneQuery", map[string]any{"query_id": queryID}, &latest); err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, struct {
			*models.GetQueryResponse
			LatestExecutionID    *string `json:"latest_execution_id"`
			LatestExecutionState *string `json:"latest_execution_state"`
		}{resp, latest.ID, latest.State})
	default:
		fmt.Fprintf(w, "ID:          %d\n", resp.QueryID)
		fmt.Fprintf(w, "Name:        %s\n", resp.Name)
		if resp.Description != "" {
			fmt.Fprintf(w, "Description: %s\n", resp.Description)
		}
		fmt.Fprintf(w, "Owner:       %s\n", resp.Owner)
		fmt.Fprintf(w, "Engine:      %s\n", resp.QueryEngine)
		fmt.Fprintf(w, "Version:     %d\n", resp.Version)
		fmt.Fprintf(w, "Private:     %t\n", resp.IsPrivate)
		fmt.Fprintf(w, "Archived:    %t\n", resp.IsArchived)
		if len(resp.Tags) > 0 {
			fmt.Fprintf(w, "Tags:        %s\n", strings.Join(resp.Tags, ", "))
		}
		switch {
		case latest.ID == nil:
			fmt.Fprintln(w, "Last run:    never")
		case latest.State == nil:
			fmt.Fprintf(w, "Last run:    %s\n", *latest.ID)
		default:
			fmt.Fprintf(w, "Last run:    %s (%s)\n", *latest.ID, strings.TrimPrefix(*latest.State, "QUERY_STATE_"))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "SQL:")
		for _, line := range strings.Split(resp.QuerySQL, "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
		return nil
	}
}
