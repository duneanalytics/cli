package contract

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newSubmissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submissions",
		Short: "List your contract decoding submissions and their status",
		Long: "List the contract decoding submissions made by the user behind the API key, newest\n" +
			"first, including those made on dune.com. Filters combine with AND.\n\n" +
			"Statuses: pending (queued, picked up within about a minute), approved (about to be\n" +
			"decoded), processed (decoded tables are available), needs_manual_review (waiting for\n" +
			"the Dune team; always the case for upgrades, renames, deletions, and protected\n" +
			"namespaces), rejected (see the comment for why), in_progress and cancelled.\n\n" +
			"Decoding takes minutes and manual review can take days, so check back no sooner\n" +
			"than a few minutes later rather than polling.\n\n" +
			"Examples:\n" +
			"  dune contract submissions\n" +
			"  dune contract submissions --status pending --blockchain ethereum\n" +
			"  dune contract submissions --project uniswap --cursor <next-cursor> -o json",
		Args: cobra.NoArgs,
		RunE: runSubmissions,
	}
	cmd.Flags().String("blockchain", "", "filter by blockchain, e.g. ethereum")
	cmd.Flags().String("address", "", "filter by the full contract address; not case-sensitive")
	cmd.Flags().String("project", "", "filter by project (namespace) name; case-sensitive")
	cmd.Flags().String("contract-name", "", "filter by contract name; case-sensitive")
	cmd.Flags().String("status", "", "filter by status: pending, rejected, approved, processed, in_progress, "+
		"cancelled, or needs_manual_review")
	cmd.Flags().Int("limit", 50, "maximum number of submissions to return (1-250)")
	cmd.Flags().String("cursor", "", "next cursor from the previous page")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSubmissions(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"blockchain":    "blockchain_name",
		"address":       "address",
		"project":       "project_name",
		"contract-name": "contract_name",
		"status":        "status",
		"limit":         "limit",
		"cursor":        "cursor",
	})
	var res struct {
		Submissions []struct {
			ID             string `json:"id"`
			BlockchainName string `json:"blockchain_name"`
			Address        string `json:"address"`
			ProjectName    string `json:"project_name"`
			ContractName   string `json:"contract_name"`
			Status         string `json:"status"`
			SubmissionType string `json:"submission_type"`
			CreatedAt      string `json:"created_at"`
		} `json:"submissions"`
		Total      int    `json:"total"`
		NextCursor string `json:"next_cursor"`
	}
	raw, err := cmdutil.CallTool(cmd, "listContractSubmissions", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	if len(res.Submissions) == 0 {
		fmt.Fprintln(w, "No submissions found.")
		return nil
	}
	rows := make([][]string, len(res.Submissions))
	for i, s := range res.Submissions {
		rows[i] = []string{s.ID, s.BlockchainName, s.Address, s.ProjectName + "." + s.ContractName,
			s.SubmissionType, s.Status, s.CreatedAt}
	}
	output.PrintTable(w, []string{"ID", "BLOCKCHAIN", "ADDRESS", "CONTRACT", "TYPE", "STATUS", "CREATED"}, rows)
	fmt.Fprintf(w, "\n%d of %d submissions\n", len(res.Submissions), res.Total)
	if res.NextCursor != "" {
		fmt.Fprintf(w, "More submissions: --cursor %s\n", res.NextCursor)
	}
	return nil
}
