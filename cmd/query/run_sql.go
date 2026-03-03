package query

import (
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newRunSQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-sql",
		Short: "Execute raw SQL and display results",
		Args:  cobra.NoArgs,
		RunE:  runRunSQL,
	}

	cmd.Flags().String("sql", "", "SQL query to execute (required)")
	_ = cmd.MarkFlagRequired("sql")
	cmd.Flags().StringArray("param", nil, "query parameter in key=value format (repeatable)")
	cmd.Flags().String("performance", "medium", `performance tier: "medium" or "large"`)
	cmd.Flags().Int("limit", 0, "maximum number of rows to display (0 = all)")
	cmd.Flags().Bool("no-wait", false, "submit execution and exit without waiting for results")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runRunSQL(cmd *cobra.Command, _ []string) error {
	sql, _ := cmd.Flags().GetString("sql")

	paramFlags, _ := cmd.Flags().GetStringArray("param")
	params, err := parseParams(paramFlags)
	if err != nil {
		return err
	}

	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}

	req := models.ExecuteSQLRequest{
		SQL:         sql,
		Performance: performance,
	}
	if len(params) > 0 {
		req.QueryParameters = params
	}

	noWait, _ := cmd.Flags().GetBool("no-wait")
	if noWait {
		return runSQLNoWait(cmd, req)
	}
	return runSQLWait(cmd, req)
}

func runSQLNoWait(cmd *cobra.Command, req models.ExecuteSQLRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.SQLExecute(req)
	if err != nil {
		return err
	}

	return displayExecuteResponse(cmd, resp)
}

func runSQLWait(cmd *cobra.Command, req models.ExecuteSQLRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	exec, err := client.RunSQL(req)
	if err != nil {
		return err
	}

	return waitAndDisplay(cmd, exec)
}
