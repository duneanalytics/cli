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
		Short: "Execute raw DuneSQL and display results",
		Long: "Execute a DuneSQL query directly without creating a saved query.\n" +
			"Ideal for ad-hoc exploration and one-off analysis.\n\n" +
			"By default, polls every 5 seconds for up to ~5 minutes waiting for completion.\n" +
			"Use --no-wait to submit and exit immediately.\n\n" +
			"Examples:\n" +
			"  dune query run-sql --sql \"SELECT block_number, block_time FROM ethereum.blocks ORDER BY block_number DESC LIMIT 5\"\n" +
			"  dune query run-sql --sql \"SELECT * FROM ethereum.transactions WHERE block_number = {{block_num}}\" --param block_num=20000000\n" +
			"  dune query run-sql --sql \"SELECT COUNT(*) FROM ethereum.transactions\" --performance large",
		Args: cobra.NoArgs,
		RunE: runRunSQL,
	}

	cmd.Flags().String("sql", "", "DuneSQL query to execute (required)")
	_ = cmd.MarkFlagRequired("sql")
	cmd.Flags().StringArray("param", nil, "query parameter in key=value format (repeatable)")
	cmd.Flags().String("performance", "medium", `performance tier: "medium" (default) or "large" for higher compute resources`)
	cmd.Flags().Int("limit", 0, "maximum number of rows to display (0 = all)")
	cmd.Flags().Bool("no-wait", false, "submit execution and exit without waiting for results")
	cmd.Flags().Int("timeout", 300, "maximum seconds to wait for the execution to complete before timing out")
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

	timeout, _ := cmd.Flags().GetInt("timeout")
	return runSQLWait(cmd, req, timeout)
}

func runSQLNoWait(cmd *cobra.Command, req models.ExecuteSQLRequest) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.SQLExecute(req)
	if err != nil {
		return err
	}

	return displayExecuteResponse(cmd, resp)
}

func runSQLWait(cmd *cobra.Command, req models.ExecuteSQLRequest, timeout int) error {
	client := cmdutil.ClientFromCmd(cmd)

	exec, err := client.RunSQL(req)
	if err != nil {
		return err
	}

	return waitAndDisplay(cmd, exec, timeout)
}
