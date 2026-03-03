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
		Short: "Create a new saved query",
		RunE:  runCreate,
	}

	cmd.Flags().String("name", "", "query name (required)")
	cmd.Flags().String("sql", "", "query SQL (required)")
	cmd.Flags().String("description", "", "query description")
	cmd.Flags().Bool("private", false, "make the query private")
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

	resp, err := client.CreateQuery(models.CreateQueryRequest{
		Name:        name,
		QuerySQL:    sql,
		Description: description,
		IsPrivate:   private,
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
