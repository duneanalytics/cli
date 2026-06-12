package matview

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Permanently delete a materialized view",
		Long: "Permanently delete a materialized view and its refresh schedule. This drops the\n" +
			"underlying table — it can no longer be queried. This action cannot be undone.\n\n" +
			"The name must be fully qualified, e.g. dune.my_team.result_token_summary.\n" +
			"You must own the matview.\n\n" +
			"Examples:\n" +
			"  dune matview delete dune.my_team.result_token_summary\n" +
			"  dune mv delete dune.my_team.result_token_summary -o json",
		Args: cobra.ExactArgs(1),
		RunE: runDelete,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.DeleteMaterializedView(args[0])
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Deleted materialized view %s\n", args[0])
		return nil
	}
}
