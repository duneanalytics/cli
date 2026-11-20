package query

import (
	"fmt"
	"strings"

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
	cmd.Flags().String("sql", "", "new SQL content in DuneSQL dialect, @path, or @- (stdin); max 500,000 characters")
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
	if fs.Changed("sql") {
		sql, err := cmdutil.FileFlag(cmd, "sql")
		if err != nil {
			return err
		}
		v := string(sql)
		req.QuerySQL = &v
		patch = true
	}
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
	private, _ := fs.GetBool("private")
	temp, _ := fs.GetBool("temp")

	// The changes take up to five calls, which are not atomic: unarchive first, so an
	// archived query can be edited, and archive last. The PATCH endpoint ignores privacy and
	// archive state, so they have their own endpoints, and the REST API has no endpoint for
	// the temp flag, so it alone still goes through the Dune MCP.
	type step struct {
		change string
		run    func() error
	}
	var steps []step
	if fs.Changed("archived") && !archived {
		steps = append(steps, step{"unarchive the query", func() error {
			_, err := client.UnarchiveQuery(queryID)
			return err
		}})
	}
	if patch {
		steps = append(steps, step{"update the query's fields", func() error {
			_, err := client.UpdateQuery(queryID, req)
			return err
		}})
	}
	if fs.Changed("private") {
		change, call := "make the query public", client.MakeQueryPublic
		if private {
			change, call = "make the query private", client.MakeQueryPrivate
		}
		steps = append(steps, step{change, func() error {
			_, err := call(queryID)
			return err
		}})
	}
	if fs.Changed("temp") {
		steps = append(steps, step{fmt.Sprintf("set temp to %t", temp), func() error {
			_, err := cmdutil.CallTool(cmd, "updateDuneQuery", map[string]any{"queryId": queryID, "is_temp": temp}, nil)
			return err
		}})
	}
	if fs.Changed("archived") && archived {
		steps = append(steps, step{"archive the query", func() error {
			_, err := client.ArchiveQuery(queryID)
			return err
		}})
	}
	var done []string
	for _, st := range steps {
		if err := st.run(); err != nil {
			if len(done) == 0 {
				return fmt.Errorf("could not %s; nothing was changed: %w", st.change, err)
			}
			return fmt.Errorf("could not %s; already applied: %s; later changes were skipped: %w",
				st.change, strings.Join(done, ", "), err)
		}
		done = append(done, st.change)
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
