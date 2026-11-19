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

	args := cmdutil.ToolArgs(cmd, map[string]string{
		"name":        "name",
		"sql":         "query",
		"description": "description",
		"private":     "is_private",
		"archived":    "is_archived",
		"temp":        "is_temp",
		"tags":        "tags",
	})
	if cmd.Flags().Changed("parameters") {
		var params []any
		if err := cmdutil.JSONFlag(cmd, "parameters", &params); err != nil {
			return err
		}
		args["parameters"] = params
	}
	if len(args) == 0 {
		return fmt.Errorf("at least one flag must be provided " +
			"(--name, --sql, --description, --private, --archived, --temp, --tags, or --parameters)")
	}
	if d, ok := args["description"]; ok && d == "" {
		args["description"] = nil
	}
	args["queryId"] = queryID

	if _, err := cmdutil.CallTool(cmd, "updateDuneQuery", args, nil); err != nil {
		return err
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
