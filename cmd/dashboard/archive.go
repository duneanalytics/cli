package dashboard

import (
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <dashboard_id>",
		Short: "Archive a dashboard by its ID",
		Long: `Archive a dashboard by its ID. Archived dashboards are hidden from public view
but can be restored later.

This also deletes any scheduled refresh jobs associated with the dashboard.

Examples:
  dune dashboard archive 12345
  dune dashboard archive 12345 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: runArchive,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runArchive(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid dashboard ID %q: must be an integer", args[0])
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.ArchiveDashboard(dashboardID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Archived dashboard %d\n", dashboardID)
		return nil
	}
}
