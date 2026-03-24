package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewActivityCmd returns the `sim evm activity` command.
func NewActivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity <address>",
		Short: "Get a decoded activity feed for a wallet address across EVM chains",
		Long: "Return a reverse-chronological feed of human-readable on-chain activity for\n" +
			"the given wallet address. Unlike raw transactions, activity entries are decoded\n" +
			"and classified into semantic types: sends, receives, mints, burns, token swaps,\n" +
			"approvals, and contract calls.\n\n" +
			"Activity types:\n" +
			"  - send/receive: native or token transfers to/from the wallet\n" +
			"  - mint/burn: token creation or destruction involving the wallet\n" +
			"  - swap: DEX token exchanges (includes from/to token details)\n" +
			"  - approve: ERC20/ERC721 spending approvals\n" +
			"  - call: contract interactions with decoded function name and inputs\n\n" +
			"Asset types: native, erc20, erc721, erc1155.\n\n" +
			"Each activity item includes the transaction context, transfer amounts with\n" +
			"USD values, and token metadata. Swap entries include both sides of the trade.\n" +
			"Call entries include the decoded function name and inputs.\n\n" +
			"By default, returns all activity types across all default chains.\n" +
			"Run 'dune sim evm supported-chains' to see which chains support activity.\n\n" +
			"For raw transaction data (hashes, gas, calldata), use 'dune sim evm transactions'.\n\n" +
			"Examples:\n" +
			"  dune sim evm activity 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm activity 0xd8da... --activity-type send,receive --chain-ids 1\n" +
			"  dune sim evm activity 0xd8da... --asset-type erc20 --limit 50 -o json\n" +
			"  dune sim evm activity 0xd8da... --token-address 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		Args: cobra.ExactArgs(1),
		RunE: runActivity,
	}

	cmd.Flags().String("chain-ids", "", "Restrict to specific chains by numeric ID or tag name (comma-separated, e.g. '1,8453' or 'default'); defaults to all chains tagged 'default'")
	cmd.Flags().String("token-address", "", "Filter activities involving specific token contracts (comma-separated ERC20/ERC721/ERC1155 addresses)")
	cmd.Flags().String("activity-type", "", "Filter by activity classification (comma-separated): send, receive, mint, burn, swap, approve, call; defaults to all types")
	cmd.Flags().String("asset-type", "", "Filter by token standard (comma-separated): native, erc20, erc721, erc1155; defaults to all standards")
	cmd.Flags().Int("limit", 0, "Maximum number of activity items to return per page (1-100, default: server-determined)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type activityResponse struct {
	Activity   []activityItem `json:"activity"`
	NextOffset string         `json:"next_offset,omitempty"`
	Warnings   []warningEntry `json:"warnings,omitempty"`
}

type activityItem struct {
	ChainID      int64          `json:"chain_id"`
	BlockNumber  int64          `json:"block_number"`
	BlockTime    string         `json:"block_time"`
	TxHash       string         `json:"tx_hash"`
	Type         string         `json:"type"`
	AssetType    string         `json:"asset_type"`
	TokenAddress string         `json:"token_address,omitempty"`
	From         string         `json:"from,omitempty"`
	To           string         `json:"to,omitempty"`
	Value        string         `json:"value,omitempty"`
	ValueUSD     float64        `json:"value_usd"`
	ID           string         `json:"id,omitempty"` // ERC721/ERC1155 token ID
	Spender      string         `json:"spender,omitempty"`
	TokenMeta    *tokenMetadata `json:"token_metadata,omitempty"`

	// Swap-specific fields.
	FromTokenAddress  string         `json:"from_token_address,omitempty"`
	FromTokenValue    string         `json:"from_token_value,omitempty"`
	FromTokenMetadata *tokenMetadata `json:"from_token_metadata,omitempty"`
	ToTokenAddress    string         `json:"to_token_address,omitempty"`
	ToTokenValue      string         `json:"to_token_value,omitempty"`
	ToTokenMetadata   *tokenMetadata `json:"to_token_metadata,omitempty"`

	// Contract call fields.
	Function         *functionInfo    `json:"function,omitempty"`
	ContractMetadata *contractMetaObj `json:"contract_metadata,omitempty"`
}

type tokenMetadata struct {
	Symbol   string  `json:"symbol"`
	Decimals int     `json:"decimals"`
	Name     string  `json:"name,omitempty"`
	Logo     string  `json:"logo,omitempty"`
	PriceUSD float64 `json:"price_usd"`
	PoolSize float64 `json:"pool_size,omitempty"`
	Standard string  `json:"standard,omitempty"`
}

type functionInfo struct {
	Signature string          `json:"signature,omitempty"`
	Name      string          `json:"name,omitempty"`
	Inputs    []functionInput `json:"inputs,omitempty"`
}

type functionInput struct {
	Name  string          `json:"name,omitempty"`
	Type  string          `json:"type,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

type contractMetaObj struct {
	Name string `json:"name,omitempty"`
}

func runActivity(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetString("token-address"); v != "" {
		params.Set("token_address", v)
	}
	if v, _ := cmd.Flags().GetString("activity-type"); v != "" {
		params.Set("activity_type", v)
	}
	if v, _ := cmd.Flags().GetString("asset-type"); v != "" {
		params.Set("asset_type", v)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/activity/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp activityResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Print warnings to stderr.
		printWarnings(cmd, resp.Warnings)

		columns := []string{"CHAIN_ID", "TYPE", "ASSET_TYPE", "SYMBOL", "VALUE_USD", "TX_HASH", "BLOCK_TIME"}
		rows := make([][]string, len(resp.Activity))
		for i, a := range resp.Activity {
			rows[i] = []string{
				fmt.Sprintf("%d", a.ChainID),
				a.Type,
				a.AssetType,
				activitySymbol(a),
				output.FormatUSD(a.ValueUSD),
				truncateHash(a.TxHash),
				a.BlockTime,
			}
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}

// activitySymbol returns the best symbol to display for the activity.
// For swaps it shows "FROM -> TO", for regular activities it uses token_metadata.
func activitySymbol(a activityItem) string {
	if a.Type == "swap" {
		from := ""
		to := ""
		if a.FromTokenMetadata != nil {
			from = a.FromTokenMetadata.Symbol
		}
		if a.ToTokenMetadata != nil {
			to = a.ToTokenMetadata.Symbol
		}
		if from != "" || to != "" {
			return from + " -> " + to
		}
		return ""
	}
	if a.TokenMeta != nil {
		return a.TokenMeta.Symbol
	}
	// Native transfers may not have token_metadata.
	if a.AssetType == "native" {
		return "ETH"
	}
	return ""
}

// truncateHash shortens a hex hash for table display.
func truncateHash(hash string) string {
	if len(hash) <= 14 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-4:]
}
