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
		Short: "Create a new Dune query and return the query ID",
		Long: "Create a new SQL query on Dune. Returns the query ID on success.\n\n" +
			"The query is written in DuneSQL dialect. If the query targets tables with\n" +
			"known partition columns, include a WHERE filter on those columns\n" +
			"(e.g. WHERE block_date >= CURRENT_DATE - INTERVAL '7' DAY) to enable\n" +
			"partition pruning and reduce query cost.",
		RunE: runCreate,
	}

	cmd.Flags().String("name", "", "human-readable query title, max 600 characters (required)")
	cmd.Flags().String("sql", "", "the SQL query text in DuneSQL dialect, max 500,000 characters (required)")
	cmd.Flags().String("description", "", "short description of what the query does, max 1,000 characters")
	cmd.Flags().Bool("private", false, "make the query private; may be forced by team privacy settings")
	cmd.Flags().Bool("temp", false, "create a temporary query that won't appear in the dune.com library or be accessible when shared")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("sql")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runCreate(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	name, _ := cmd.Flags().GetString("name")
	sql, _ := cmd.Flags().GetString("sql")
	description, _ := cmd.Flags().GetString("description")
	private, _ := cmd.Flags().GetBool("private")
	temp, _ := cmd.Flags().GetBool("temp")

	resp, err := client.CreateQuery(models.CreateQueryRequest{
		Name:        name,
		QuerySQL:    sql,
		Description: description,
		IsPrivate:   private,
		IsTemp:      temp,
	})
	if err != nil {
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
