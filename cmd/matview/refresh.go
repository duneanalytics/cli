package matview

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh <name>",
		Short: "Trigger an on-demand refresh of a materialized view",
		Long: "Re-execute a matview's source query and update its table with fresh results.\n" +
			"Refreshing consumes credits based on the compute used. You must own the matview.\n\n" +
			"The name must be fully qualified, e.g. dune.my_team.result_token_summary.\n\n" +
			"Examples:\n" +
			"  dune matview refresh dune.my_team.result_token_summary\n" +
			"  dune matview refresh dune.my_team.result_token_summary --performance large",
		Args: cobra.ExactArgs(1),
		RunE: runRefresh,
	}

	cmd.Flags().String("performance", "", "execution tier: \"small\", \"medium\", or \"large\" (default: account default)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runRefresh(cmd *cobra.Command, args []string) error {
	performance, err := parsePerformance(cmd)
	if err != nil {
		return err
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.RefreshMaterializedView(args[0], models.RefreshMaterializedViewRequest{
		Performance: performance,
	})
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Refreshing materialized view %s\n", resp.SQLID)
		fmt.Fprintf(w, "Execution ID: %s\n", resp.ExecutionID)
		fmt.Fprintf(w, "\nCheck results with:\n  dune execution results %s\n", resp.ExecutionID)
		return nil
	}
}
