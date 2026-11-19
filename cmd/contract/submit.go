package contract

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// singleFlags maps the flags that describe one contract to its submission fields.
var singleFlags = map[string]string{
	"blockchain":         "blockchain_name",
	"address":            "address",
	"project":            "project_name",
	"contract-name":      "contract_name",
	"proxy":              "is_proxy",
	"multiple-instances": "has_multiple_instances",
	"factory":            "is_created_by_factory",
	"manual-abi":         "is_manual_abi",
	"type":               "submission_type",
	"reason":             "resubmission_reason",
	"old-project":        "old_project_name",
	"old-contract-name":  "old_contract_name",
	"idempotency-key":    "idempotency_key",
}

func newSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit smart contracts for decoding",
		Long: "Submit a smart contract for decoding, or up to 100 with --batch. Each contract is\n" +
			"validated and queued on its own, so a bad ABI in one item does not fail the others;\n" +
			"the command exits with an error if any item failed.\n\n" +
			"Before submitting:\n" +
			"  - Use 'dune dataset blockchains' for valid --blockchain values and\n" +
			"    'dune dataset search-by-contract' to check that the address is not decoded yet.\n" +
			"  - Provide the ABI. For a proxy, submit the implementation's ABI and set --proxy.\n\n" +
			"--project and --contract-name are case-sensitive and become part of the table names;\n" +
			"reuse the same values when submitting the same contract on more chains. Free plans\n" +
			"can submit new contracts for one blockchain per call; upgrades, renames, deletions,\n" +
			"and other requests need a paid plan and always go to manual review.\n\n" +
			"Set --idempotency-key when you may retry: resubmitting a used key returns the\n" +
			"existing submission instead of creating a duplicate.\n\n" +
			"--batch takes a JSON array of submissions with the fields blockchain_name, address,\n" +
			"project_name, contract_name, abi, and optionally is_proxy, has_multiple_instances,\n" +
			"is_created_by_factory, is_manual_abi, submission_type, resubmission_reason,\n" +
			"old_project_name, old_contract_name, and idempotency_key.\n\n" +
			"--abi and --batch take inline JSON, @path to read a file, or @- to read stdin.\n\n" +
			"Track progress with 'dune contract submissions'.\n\n" +
			"Examples:\n" +
			"  dune contract submit --blockchain ethereum --address 0x1f98...F984 \\\n" +
			"    --project uniswap --contract-name UniswapToken --abi @uni.abi.json\n" +
			"  dune contract submit --batch @submissions.json -o json",
		Args: cobra.NoArgs,
		RunE: runSubmit,
	}
	cmd.Flags().String("blockchain", "", "blockchain the contract is deployed on, e.g. ethereum or base")
	cmd.Flags().String("address", "", "contract address; 0x-prefixed for EVM chains, base58 for tron")
	cmd.Flags().String("project", "", "project (namespace) the decoded tables are grouped under, e.g. uniswap")
	cmd.Flags().String("contract-name", "", "contract name used in the decoded table names, e.g. UniswapToken")
	cmd.Flags().String("abi", "", "contract ABI as a JSON array, @path, or @-")
	cmd.Flags().Bool("proxy", false, "the address is a proxy and the ABI belongs to its implementation")
	cmd.Flags().Bool("multiple-instances", false, "the ABI is shared by several deployed instances")
	cmd.Flags().Bool("factory", false, "the instances are created by a factory contract")
	cmd.Flags().Bool("manual-abi", false, "the ABI was written or edited by hand rather than fetched from an explorer")
	cmd.Flags().String("type", "", `submission type: "new" (default), "upgrade", "rename", "delete", or "other"`)
	cmd.Flags().String("reason", "", `why the contract is resubmitted; required for "delete" and "other"`)
	cmd.Flags().String("old-project", "", `current project name; required for "rename"`)
	cmd.Flags().String("old-contract-name", "", `current contract name; required for "rename"`)
	cmd.Flags().String("idempotency-key", "", "key, unique per account, that makes a retry return the existing submission")
	cmd.Flags().String("batch", "",
		"JSON array of up to 100 submissions, @path, or @-, instead of the single-contract flags")
	for name := range singleFlags {
		cmd.MarkFlagsMutuallyExclusive("batch", name)
	}
	cmd.MarkFlagsMutuallyExclusive("batch", "abi")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSubmit(cmd *cobra.Command, _ []string) error {
	var submissions []any
	if cmd.Flags().Changed("batch") {
		if err := cmdutil.JSONFlag(cmd, "batch", &submissions); err != nil {
			return err
		}
	} else {
		for _, f := range []string{"blockchain", "address", "project", "contract-name", "abi"} {
			if !cmd.Flags().Changed(f) {
				return fmt.Errorf("--%s is required unless you pass --batch", f)
			}
		}
		item := cmdutil.ToolArgs(cmd, singleFlags)
		var abi any
		if err := cmdutil.JSONFlag(cmd, "abi", &abi); err != nil {
			return err
		}
		item["abi"] = abi
		submissions = []any{item}
	}

	var res struct {
		Results []struct {
			Index        int    `json:"index"`
			SubmissionID string `json:"submission_id"`
			Status       string `json:"status"`
			Replayed     bool   `json:"replayed"`
			Error        string `json:"error"`
		} `json:"results"`
		StatusURL string `json:"status_url"`
	}
	raw, err := cmdutil.CallTool(cmd, "submitContractsForDecoding", map[string]any{"submissions": submissions}, &res)
	if err != nil {
		return err
	}

	failed := 0
	for _, r := range res.Results {
		if r.Error != "" {
			failed++
		}
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		if err := output.PrintRawJSON(w, raw); err != nil {
			return err
		}
	} else {
		rows := make([][]string, len(res.Results))
		for i, r := range res.Results {
			status := r.Status
			if r.Replayed {
				status += " (existing)"
			}
			rows[i] = []string{fmt.Sprintf("%d", r.Index), r.SubmissionID, status, r.Error}
		}
		output.PrintTable(w, []string{"INDEX", "SUBMISSION_ID", "STATUS", "ERROR"}, rows)
		if res.StatusURL != "" {
			fmt.Fprintf(w, "\nTrack progress at %s or with 'dune contract submissions'\n", res.StatusURL)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d submissions failed", failed, len(res.Results))
	}
	return nil
}
