package execution

import "github.com/spf13/cobra"

// NewExecutionCmd returns the `execution` parent command.
func NewExecutionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execution",
		Short: "Retrieve query execution results and manage the team's runs",
		Long: "Retrieve the results of query executions, and list or stop the team's runs.\n\n" +
			"Use 'execution results <execution-id>' to fetch results from a previously\n" +
			"submitted query execution. The execution ID is returned when running queries\n" +
			"with 'dune query run --no-wait' or 'dune query run-sql --no-wait'.\n\n" +
			"Use 'execution list' to see the team's active or recently finished runs, and\n" +
			"'execution cancel <execution-id>' to stop an active run.",
	}
	cmd.AddCommand(newResultsCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCancelCmd())
	return cmd
}
