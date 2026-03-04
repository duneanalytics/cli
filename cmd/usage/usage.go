package usage

import (
	"fmt"
	"time"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// NewUsageCmd returns the top-level "usage" command.
func NewUsageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show credit and resource usage for your Dune account",
		RunE:  runUsage,
	}

	cmd.Flags().String("start-date", "", "filter start date (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "filter end date (YYYY-MM-DD)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUsage(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")

	if err := validateDateFlag(startDate, "start-date"); err != nil {
		return err
	}
	if err := validateDateFlag(endDate, "end-date"); err != nil {
		return err
	}

	var (
		resp *models.UsageResponse
		err  error
	)
	if startDate != "" || endDate != "" {
		resp, err = client.GetUsageForDates(startDate, endDate)
	} else {
		resp, err = client.GetUsage()
	}
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Private Queries:      %d\n", resp.PrivateQueries)
		fmt.Fprintf(w, "Private Dashboards:   %d\n", resp.PrivateDashboards)
		fmt.Fprintf(w, "Storage Used:         %s / %s\n",
			output.FormatBytes(resp.BytesUsed), output.FormatBytes(resp.BytesAllowed))

		if len(resp.BillingPeriods) > 0 {
			fmt.Fprintln(w)
			columns := []string{"START DATE", "END DATE", "CREDITS USED", "CREDITS INCLUDED"}
			rows := make([][]string, len(resp.BillingPeriods))
			for i, bp := range resp.BillingPeriods {
				rows[i] = []string{
					bp.StartDate,
					bp.EndDate,
					fmt.Sprintf("%.2f", bp.CreditsUsed),
					fmt.Sprintf("%d", bp.CreditsIncluded),
				}
			}
			output.PrintTable(w, columns, rows)
		}
		return nil
	}
}

func validateDateFlag(value, name string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid --%s: expected YYYY-MM-DD format", name)
	}
	return nil
}
