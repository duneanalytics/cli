package execution

import "github.com/spf13/cobra"

// NewExecutionCmd returns the `execution` parent command.
func NewExecutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution",
		Short: "Manage query executions",
	}
	cmd.AddCommand(newResultsCmd())
	return cmd
}
