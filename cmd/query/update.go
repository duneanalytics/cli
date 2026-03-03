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
		Short: "Update an existing saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  runUpdate,
	}

	cmd.Flags().String("name", "", "query name")
	cmd.Flags().String("sql", "", "query SQL")
	cmd.Flags().String("description", "", "query description")
	cmd.Flags().Bool("private", false, "make the query private")
	cmd.Flags().StringSlice("tags", nil, "query tags (comma-separated)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	queryID, err := parseQueryID(args[0])
	if err != nil {
		return err
	}

	var req models.UpdateQueryRequest
	changed := false

	if cmd.Flags().Changed("name") {
		v, _ := cmd.Flags().GetString("name")
		req.Name = &v
		changed = true
	}
	if cmd.Flags().Changed("sql") {
		v, _ := cmd.Flags().GetString("sql")
		req.QuerySQL = &v
		changed = true
	}
	if cmd.Flags().Changed("description") {
		v, _ := cmd.Flags().GetString("description")
		req.Description = &v
		changed = true
	}
	if cmd.Flags().Changed("private") {
		v, _ := cmd.Flags().GetBool("private")
		req.IsPrivate = &v
		changed = true
	}
	if cmd.Flags().Changed("tags") {
		v, _ := cmd.Flags().GetStringSlice("tags")
		req.Tags = v
		changed = true
	}

	if !changed {
		return fmt.Errorf("at least one flag must be provided (--name, --sql, --description, --private, or --tags)")
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.UpdateQuery(queryID, req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Updated query %d\n", resp.QueryID)
		return nil
	}
}
