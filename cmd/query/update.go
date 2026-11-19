package query

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <query-id>",
		Short: "Update an existing Dune query's SQL, title, description, privacy, tags, or parameters",
		Long: "Modify a query you own or have edit access to via team membership.\n" +
			"Only supply the flags you want to change; unchanged fields are preserved.\n" +
			"At least one flag must be provided.\n\n" +
			"The update uses optimistic locking — if someone else edited the query\n" +
			"concurrently, you'll get a conflict error.\n\n" +
			"Examples:\n" +
			"  dune query update 12345 --sql \"SELECT * FROM ethereum.blocks LIMIT 20\"\n" +
			"  dune query update 12345 --name \"Renamed Query\" --tags defi,ethereum\n" +
			"  dune query update 12345 --private\n" +
			"  dune query update 12345 --archived=false   # restore an archived query",
		Args: cobra.ExactArgs(1),
		RunE: runUpdate,
	}

	cmd.Flags().String("name", "", "new title for the query, max 600 characters")
	cmd.Flags().String("sql", "", "new SQL content in DuneSQL dialect, max 500,000 characters")
	cmd.Flags().String("description", "", "new description for the query, max 1,000 characters; an empty value clears it")
	cmd.Flags().Bool("private", false, "set to true to make the query private, false to make public")
	cmd.Flags().Bool("archived", false, "set to true to archive the query, false to restore it from the archive")
	cmd.Flags().Bool("temp", false, "set to true to make the query temporary, false to save it to the library")
	cmd.Flags().StringSlice("tags", nil, "new set of tags for the query (comma-separated); replaces all existing tags")
	cmd.Flags().String("parameters", "", parametersHelp+"; replaces all existing parameters")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, posArgs []string) error {
	queryID, err := parseQueryID(posArgs[0])
	if err != nil {
		return err
	}

	fs := cmd.Flags()
	changed := false
	for _, f := range []string{"name", "sql", "description", "private", "archived", "temp", "tags", "parameters"} {
		changed = changed || fs.Changed(f)
	}
	if !changed {
		return fmt.Errorf("at least one flag must be provided " +
			"(--name, --sql, --description, --private, --archived, --temp, --tags, or --parameters)")
	}

	var req models.UpdateQueryRequest
	patch := false
	stringFlag := func(name string) *string {
		if !fs.Changed(name) {
			return nil
		}
		v, _ := fs.GetString(name)
		patch = true
		return &v
	}
	req.Name = stringFlag("name")
	req.QuerySQL = stringFlag("sql")
	req.Description = stringFlag("description")
	if fs.Changed("tags") {
		// An empty --tags gives an empty, non-nil slice, which the SDK sends to clear the tags.
		req.Tags, _ = fs.GetStringSlice("tags")
		patch = true
	}
	if fs.Changed("parameters") {
		if err := cmdutil.JSONFlag(cmd, "parameters", &req.Parameters); err != nil {
			return err
		}
		patch = true
	}

	client := cmdutil.ClientFromCmd(cmd)
	archived, _ := fs.GetBool("archived")
	// Restore an archived query before editing it, and archive one only after.
	if fs.Changed("archived") && !archived {
		if _, err := client.UnarchiveQuery(queryID); err != nil {
			return err
		}
	}
	if patch {
		if _, err := client.UpdateQuery(queryID, req); err != nil {
			return err
		}
	}
	// The PATCH endpoint ignores privacy and archive state, so they have their own calls.
	if fs.Changed("private") {
		private, _ := fs.GetBool("private")
		if private {
			_, err = client.MakeQueryPrivate(queryID)
		} else {
			_, err = client.MakeQueryPublic(queryID)
		}
		if err != nil {
			return err
		}
	}
	if fs.Changed("temp") {
		// The REST API has no endpoint for the temp flag, so it alone still goes through the Dune MCP.
		temp, _ := fs.GetBool("temp")
		args := map[string]any{"queryId": queryID, "is_temp": temp}
		if _, err := cmdutil.CallTool(cmd, "updateDuneQuery", args, nil); err != nil {
			return err
		}
	}
	if fs.Changed("archived") && archived {
		if _, err := client.ArchiveQuery(queryID); err != nil {
			return err
		}
	}
	resp := models.UpdateQueryResponse{QueryID: queryID}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Updated query %d\n", resp.QueryID)
		return nil
	}
}
