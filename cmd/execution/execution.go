package execution

import "github.com/spf13/cobra"

// NewExecutionCmd returns the `execution` parent command.
func NewExecutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution",
		Short: "Retrieve and inspect query execution results",
		Long: "Retrieve and inspect the results of query executions.\n\n" +
			"Use 'execution results <execution-id>' to fetch results from a previously\n" +
			"submitted query execution. The execution ID is returned when running queries\n" +
			"with 'dune query run --no-wait' or 'dune query run-sql --no-wait'.",
	}
	cmd.AddCommand(newResultsCmd())
	return cmd
}
