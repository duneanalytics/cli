package dataset

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newSearchByContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search-by-contract",
		Short: "Search for decoded tables associated with a specific contract address",
		Long: `Search for decoded tables (events and calls) associated with a specific contract address.

This command finds all available decoded tables for a given smart contract address.
Decoded tables contain blockchain data that has been parsed according to a contract's ABI,
providing a structured view of contract interactions.

Table types returned:
  - Event tables: contain parsed event logs emitted by smart contracts
  - Call tables: contain parsed function calls made to smart contracts

When to use this command:
  - You have a specific contract address (EVM or Tron) and want to find its decoded tables
  - You want to analyze on-chain activity for a particular smart contract
  - You need to find event or call tables for a known contract

For broader table discovery by keyword or project name, use 'dune dataset search' instead.

Examples:
  dune dataset search-by-contract --contract-address 0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984
  dune dataset search-by-contract --contract-address 0x1f98... --blockchains ethereum --blockchains arbitrum
  dune dataset search-by-contract --contract-address 0x1f98... --include-schema
  dune dataset search-by-contract --contract-address 0x1f98... --limit 50 --offset 20`,
		RunE: runSearchByContract,
	}

	cmd.Flags().String("contract-address", "",
		"The contract address to search for. Accepts EVM addresses (starting with 0x) or Tron addresses. "+
			"Example: '0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984'.")
	_ = cmd.MarkFlagRequired("contract-address")
	cmd.Flags().StringArray("blockchains", nil,
		"Filter results to specific blockchains (e.g. ethereum, arbitrum). "+
			"Can be specified multiple times. If not provided, searches across all blockchains where the contract exists.")
	cmd.Flags().Bool("include-schema", false,
		"If set, include the column-level schema (name, type, nullable) for every result. "+
			"Enable when preparing SQL generation.")
	cmd.Flags().Int32("limit", 20,
		"Maximum number of results to return. "+
			"Use smaller values (5-10) for quick checks, larger values (50-100) for comprehensive discovery.")
	cmd.Flags().Int32("offset", 0,
		"Number of results to skip for pagination. Use for paginating through large result sets.")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runSearchByContract(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	contractAddress, _ := cmd.Flags().GetString("contract-address")

	req := models.SearchDatasetsByContractAddressRequest{
		ContractAddress: contractAddress,
	}

	if cmd.Flags().Changed("blockchains") {
		v, _ := cmd.Flags().GetStringArray("blockchains")
		req.Blockchains = v
	}
	if cmd.Flags().Changed("include-schema") {
		v, _ := cmd.Flags().GetBool("include-schema")
		req.IncludeSchema = &v
	}
	if cmd.Flags().Changed("limit") {
		v, _ := cmd.Flags().GetInt32("limit")
		req.Limit = &v
	}
	if cmd.Flags().Changed("offset") {
		v, _ := cmd.Flags().GetInt32("offset")
		req.Offset = &v
	}

	resp, err := client.SearchDatasetsByContractAddress(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		columns := []string{"FULL_NAME", "CATEGORY", "DATASET_TYPE", "BLOCKCHAINS"}
		rows := make([][]string, len(resp.Results))
		for i, r := range resp.Results {
			dt := ""
			if r.DatasetType != nil {
				dt = *r.DatasetType
			}
			rows[i] = []string{
				r.FullName,
				r.Category,
				dt,
				strings.Join(r.Blockchains, ", "),
			}
		}
		output.PrintTable(w, columns, rows)
		fmt.Fprintf(w, "\n%d of %d results\n", len(resp.Results), resp.Total)
		return nil
	}
}
