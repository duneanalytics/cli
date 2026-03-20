package whoami

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// NewWhoAmICmd returns the "whoami" command.
func NewWhoAmICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the identity associated with the current API key",
		Long: "Display the handle, customer ID, and API key ID of the currently\n" +
			"configured API key. Useful for verifying which account is active.\n\n" +
			"Examples:\n" +
			"  dune whoami\n" +
			"  dune whoami --output json",
		RunE: runWhoAmI,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runWhoAmI(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.WhoAmI()
	if err != nil {
		return fmt.Errorf("identifying API key: %w", err)
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "Handle:        %s\n", resp.Handle)
		fmt.Fprintf(w, "Customer ID:   %s\n", resp.CustomerID)
		fmt.Fprintf(w, "API Key ID:    %s\n", resp.APIKeyID)
		return nil
	}
}
