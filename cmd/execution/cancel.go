package execution

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <execution-id>",
		Short: "Stop one of the team's queued or running query runs",
		Long: "Stop one of the team's queued or running query runs. Team admins can stop any\n" +
			"of the team's runs, including API, MCP and scheduled runs that other members\n" +
			"started. Other members can stop only the runs they started.\n\n" +
			"When other users submitted the same SQL on the same engine while the run was\n" +
			"active, they share the run, and it stops for them too. So the command asks\n" +
			"before it cancels; --yes skips the question.\n\n" +
			"Needs a team API key. Find the run with 'dune execution list'.\n\n" +
			"Examples:\n" +
			"  dune execution cancel 01KAQ5ABCDEFGHJKMNPQRSTVWX\n" +
			"  dune execution cancel 01KAQ5ABCDEFGHJKMNPQRSTVWX --yes -o json",
		Args: cobra.ExactArgs(1),
		RunE: runCancel,
	}
	cmd.Flags().BoolP("yes", "y", false, "cancel without asking")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runCancel(cmd *cobra.Command, args []string) error {
	executionID := args[0]
	if yes, _ := cmd.Flags().GetBool("yes"); !yes {
		confirmed, err := confirmCancel(cmd, executionID)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(cmd.ErrOrStderr(), "Not cancelled")
			return nil
		}
	}

	var res struct {
		ExecutionID string `json:"execution_id"`
		Cancelled   bool   `json:"cancelled"`
	}
	raw, err := cmdutil.CallTool(cmd, "cancelTeamExecution", map[string]any{"execution_id": executionID}, &res)
	if err != nil {
		return err
	}
	if !res.Cancelled {
		return fmt.Errorf("run %s was not cancelled", executionID)
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	fmt.Fprintf(w, "Cancelled run %s\n", res.ExecutionID)
	return nil
}

func confirmCancel(cmd *cobra.Command, executionID string) (bool, error) {
	if stdin, ok := cmd.InOrStdin().(*os.File); ok && !term.IsTerminal(int(stdin.Fd())) {
		return false, fmt.Errorf("cancelling a run needs confirmation; pass --yes to cancel without asking")
	}
	fmt.Fprintf(cmd.ErrOrStderr(),
		"Cancel run %s? It also stops for anyone who shares the run. [y/N]: ", executionID)
	answer, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
