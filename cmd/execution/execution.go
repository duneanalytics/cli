package execution

import "github.com/spf13/cobra"

// NewExecutionCmd returns the `execution` parent command.
func NewExecutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution",
		Short: "Retrieve and inspect query execution results",
	}
	cmd.AddCommand(newResultsCmd())
	return cmd
}
