package evm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewTokenHoldersCmd returns the `sim evm token-holders` command.
func NewTokenHoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-holders <token_address>",
		Short: "Get token holders ranked by balance",
		Long: "Return a list of holders for a given ERC20 token contract on a specified chain,\n" +
			"ranked by balance descending.\n\n" +
			"Examples:\n" +
			"  dune sim evm token-holders 0x63706e401c06ac8513145b7687A14804d17f814b --chain-id 8453\n" +
			"  dune sim evm token-holders 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 --chain-id 1 --limit 50\n" +
			"  dune sim evm token-holders 0x63706e... --chain-id 8453 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runTokenHolders,
	}

	cmd.Flags().String("chain-id", "", "Numeric chain ID (required)")
	cmd.Flags().Int("limit", 0, "Max results (1-500, default 500)")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
	_ = cmd.MarkFlagRequired("chain-id")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type tokenHoldersResponse struct {
	TokenAddress string   `json:"token_address"`
	ChainID      int64    `json:"chain_id"`
	Holders      []holder `json:"holders"`
	NextOffset   string   `json:"next_offset,omitempty"`
}

type holder struct {
	WalletAddress        string `json:"wallet_address"`
	Balance              string `json:"balance"`
	FirstAcquired        string `json:"first_acquired,omitempty"`
	HasInitiatedTransfer bool   `json:"has_initiated_transfer"`
}

func runTokenHolders(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	tokenAddress := args[0]
	chainID, _ := cmd.Flags().GetString("chain-id")
	// Validate chain_id is a valid integer.
	if _, err := strconv.Atoi(chainID); err != nil {
		return fmt.Errorf("--chain-id must be a numeric value, got %q", chainID)
	}

	params := url.Values{}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	path := fmt.Sprintf("/v1/evm/token-holders/%s/%s", chainID, tokenAddress)
	data, err := client.Get(cmd.Context(), path, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp tokenHoldersResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if len(resp.Holders) == 0 {
			fmt.Fprintln(w, "No holders found.")
			return nil
		}

		columns := []string{"WALLET_ADDRESS", "BALANCE", "FIRST_ACQUIRED", "HAS_TRANSFERRED"}
		rows := make([][]string, len(resp.Holders))
		for i, h := range resp.Holders {
			transferred := "N"
			if h.HasInitiatedTransfer {
				transferred = "Y"
			}
			rows[i] = []string{
				h.WalletAddress,
				h.Balance,
				h.FirstAcquired,
				transferred,
			}
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}
