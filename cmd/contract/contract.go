package contract

import "github.com/spf13/cobra"

// NewContractCmd returns the `contract` parent command.
func NewContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Submit smart contracts for decoding and track the submissions",
		Long: "Submit smart contracts for decoding on Dune, which creates decoded event and call\n" +
			"tables (project_<chain>.Contract_evt_Event, project_<chain>.Contract_call_Function),\n" +
			"and track the status of your submissions.\n\n" +
			"Subcommands:\n" +
			"  submit       - Submit one contract, or up to 100 with --batch\n" +
			"  submissions  - List your submissions and their status",
	}
	cmd.AddCommand(newSubmitCmd())
	cmd.AddCommand(newSubmissionsCmd())
	return cmd
}
