package contract

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// singleFlags are the flags that describe one contract, which --batch replaces.
var singleFlags = []string{
	"blockchain", "address", "project", "contract-name", "abi", "proxy", "multiple-instances", "factory",
	"manual-abi", "type", "reason", "old-project", "old-contract-name", "idempotency-key",
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
	for _, name := range singleFlags {
		cmd.MarkFlagsMutuallyExclusive("batch", name)
	}
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSubmit(cmd *cobra.Command, _ []string) error {
	var req models.SubmitContractsRequest
	if cmd.Flags().Changed("batch") {
		if err := cmdutil.JSONFlag(cmd, "batch", &req.Submissions); err != nil {
			return err
		}
	} else {
		item, err := singleSubmission(cmd)
		if err != nil {
			return err
		}
		req.Submissions = []models.ContractSubmissionInput{item}
	}

	res, err := cmdutil.ClientFromCmd(cmd).SubmitContracts(req)
	if err != nil {
		return err
	}

	failed := 0
	for _, r := range res.Results {
		if r.Error != nil {
			failed++
		}
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		if err := output.PrintJSON(w, res); err != nil {
			return err
		}
	} else {
		rows := make([][]string, len(res.Results))
		for i, r := range res.Results {
			status := deref(r.Status)
			if r.Replayed {
				status += " (existing)"
			}
			rows[i] = []string{fmt.Sprintf("%d", r.Index), deref(r.SubmissionID), status, deref(r.Error)}
		}
		output.PrintTable(w, []string{"INDEX", "SUBMISSION_ID", "STATUS", "ERROR"}, rows)
		fmt.Fprintln(w, "\nTrack progress with 'dune contract submissions'")
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d submissions failed", failed, len(res.Results))
	}
	return nil
}

// singleSubmission builds one submission from the single-contract flags, leaving
// the optional fields the user did not set empty.
func singleSubmission(cmd *cobra.Command) (models.ContractSubmissionInput, error) {
	var item models.ContractSubmissionInput
	for _, f := range []string{"blockchain", "address", "project", "contract-name", "abi"} {
		if !cmd.Flags().Changed(f) {
			return item, fmt.Errorf("--%s is required unless you pass --batch", f)
		}
	}
	if err := cmdutil.JSONFlag(cmd, "abi", &item.ABI); err != nil {
		return item, err
	}
	fs := cmd.Flags()
	item.BlockchainName, _ = fs.GetString("blockchain")
	item.Address, _ = fs.GetString("address")
	item.ProjectName, _ = fs.GetString("project")
	item.ContractName, _ = fs.GetString("contract-name")
	item.IsProxy, _ = fs.GetBool("proxy")
	item.HasMultipleInstances, _ = fs.GetBool("multiple-instances")
	item.IsCreatedByFactory, _ = fs.GetBool("factory")
	item.IsManualABI, _ = fs.GetBool("manual-abi")
	item.SubmissionType = optionalString(cmd, "type")
	item.ResubmissionReason = optionalString(cmd, "reason")
	item.OldProjectName = optionalString(cmd, "old-project")
	item.OldContractName = optionalString(cmd, "old-contract-name")
	item.IdempotencyKey = optionalString(cmd, "idempotency-key")
	return item, nil
}

// optionalString returns the value of a string flag, or nil when the user did not set it.
func optionalString(cmd *cobra.Command, name string) *string {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v, _ := cmd.Flags().GetString(name)
	return &v
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
