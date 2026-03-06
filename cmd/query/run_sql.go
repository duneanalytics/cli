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
		Short: "Execute a raw DuneSQL query inline and display results",
		Long: "Execute an inline SQL statement in DuneSQL dialect without saving it as a\n" +
			"query on Dune. Ideal for ad-hoc exploration and one-off analysis.\n\n" +
			"By default, waits for completion (polling every 2 seconds) and displays result rows.\n" +
			"Use --no-wait to submit the execution and exit immediately with just the\n" +
			"execution ID. Credits are consumed based on actual compute resources used.\n\n" +
			"Important: if the SQL targets tables with known partition columns (returned by\n" +
			"'dune dataset search' or 'dune dataset search-by-contract'), include a WHERE filter\n" +
			"on those partition columns (e.g. WHERE block_date >= CURRENT_DATE - INTERVAL '7' DAY).\n" +
			"This enables partition pruning and significantly reduces query cost.\n\n" +
			"Examples:\n" +
			"  dune query run-sql --sql \"SELECT block_number, block_time FROM ethereum.blocks ORDER BY block_number DESC LIMIT 5\"\n" +
			"  dune query run-sql --sql \"SELECT * FROM ethereum.transactions WHERE block_number = {{block_num}}\" --param block_num=20000000\n" +
			"  dune query run-sql --sql \"SELECT COUNT(*) FROM ethereum.transactions\" --performance large\n" +
			"  dune query run-sql --sql \"SELECT 1\" --no-wait",
		Args: cobra.NoArgs,
		RunE: runRunSQL,
	}

	cmd.Flags().String("sql", "", "the SQL query text in DuneSQL dialect (required)")
	_ = cmd.MarkFlagRequired("sql")
	cmd.Flags().StringArray("param", nil, "typed query parameter in key=value format (repeatable); supported types: text, number (stringified, e.g. '30'), datetime (YYYY-MM-DD HH:mm:ss), enum")
	cmd.Flags().String("performance", "medium", `engine size for the execution: "medium" (default) or "large"; credits are consumed based on actual compute resources used`)
	cmd.Flags().Int("limit", 0, "maximum number of result rows to return (0 = all available rows)")
	cmd.Flags().Bool("no-wait", false, "submit the execution and exit immediately, printing only the execution ID and state")
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
