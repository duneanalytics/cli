package query

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new saved Dune query and return the query ID",
		Long: "Create a new SQL query on Dune. Returns the query ID on success.\n\n" +
			"The query is written in DuneSQL dialect. If the query targets tables with\n" +
			"known partition columns, include a WHERE filter on those columns\n" +
			"(e.g. WHERE block_date >= CURRENT_DATE - INTERVAL '7' DAY) to enable\n" +
			"partition pruning and reduce query cost.\n\n" +
			"Use --temp to create a temporary query that won't appear in the dune.com\n" +
			"library or be accessible when shared. Temporary queries are auto-deleted.\n\n" +
			"Use --folder-id or --folder-name to save the query into a library folder, and\n" +
			"--parameters to declare {{parameters}} the SQL uses.\n\n" +
			"Examples:\n" +
			"  dune query create --name \"Recent Blocks\" --sql \"SELECT * FROM ethereum.blocks LIMIT 10\"\n" +
			"  dune query create --name \"My Query\" --sql \"SELECT 1\" --private --description \"Test query\"\n" +
			"  dune query create --name \"Throwaway\" --sql \"SELECT 1\" --temp\n" +
			"  dune query create --name \"Blocks\" --sql \"SELECT * FROM ethereum.blocks LIMIT {{n}}\" \\\n" +
			"    --folder-name Research --parameters '[{\"key\":\"n\",\"type\":\"number\",\"value\":\"10\"}]'",
		RunE: runCreate,
	}

	addCreateFlags(cmd)
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// addCreateFlags registers the flags that describe a new query.
func addCreateFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "human-readable query title, max 600 characters (required)")
	cmd.Flags().String("sql", "", "the SQL query text in DuneSQL dialect, max 500,000 characters (required)")
	cmd.Flags().String("description", "", "short description of what the query does, max 1,000 characters")
	cmd.Flags().Bool("private", false, "make the query private; may be forced by team privacy settings")
	cmd.Flags().Bool("temp", false, "create a temporary query that won't appear in the dune.com library or be accessible when shared; auto-deleted")
	cmd.Flags().String("folder-id", "", "save the query into this folder (see 'dune folder list')")
	cmd.Flags().String("folder-name", "",
		"save the query into the folder with this name; fails if several folders share it")
	cmd.Flags().String("parameters", "", parametersHelp)
	cmd.MarkFlagsMutuallyExclusive("folder-id", "folder-name")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("sql")
}

// parametersHelp documents the --parameters flag of query create, update, and create-and-run.
const parametersHelp = `query parameters as a JSON array, @path, or @-: ` +
	`[{"key":"days","type":"number","value":"30"}]; ` +
	`type is text, number, date, or enum (with "enumOptions"); reference them in SQL as {{key}}`

// createArgs returns the createDuneQuery arguments for the flags of query create
// and create-and-run.
func createArgs(cmd *cobra.Command) (map[string]any, error) {
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"name":        "name",
		"sql":         "query",
		"description": "description",
		"private":     "is_private",
		"folder-id":   "folder_id",
		"folder-name": "folder_name",
	})
	// The MCP tool creates temporary queries by default; the CLI saves them.
	args["is_temp"], _ = cmd.Flags().GetBool("temp")
	if cmd.Flags().Changed("parameters") {
		var params []any
		if err := cmdutil.JSONFlag(cmd, "parameters", &params); err != nil {
			return nil, err
		}
		args["parameters"] = params
	}
	return args, nil
}

func runCreate(cmd *cobra.Command, _ []string) error {
	args, err := createArgs(cmd)
	if err != nil {
		return err
	}
	var resp models.CreateQueryResponse
	if _, err := cmdutil.CallTool(cmd, "createDuneQuery", args, &resp); err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Created query %d\n", resp.QueryID)
		return nil
	}
}
