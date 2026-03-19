package evm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewDefiPositionsCmd returns the `sim evm defi-positions` command.
func NewDefiPositionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defi-positions <address>",
		Short: "Get DeFi positions for a wallet address",
		Long: "Return DeFi positions for the given wallet address including USD values,\n" +
			"position-specific metadata, and aggregation summaries across supported protocols.\n\n" +
			"Supported position types: Erc4626 (vaults), Tokenized (lending, e.g. aTokens),\n" +
			"UniswapV2 (AMM LP), Nft (Uniswap V3 NFT), NftV4 (Uniswap V4 NFT).\n\n" +
			"Examples:\n" +
			"  dune sim evm defi-positions 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm defi-positions 0xd8da... --chain-ids 1,8453\n" +
			"  dune sim evm defi-positions 0xd8da... -o json",
		Args: cobra.ExactArgs(1),
		RunE: runDefiPositions,
	}

	cmd.Flags().String("chain-ids", "", "Comma-separated chain IDs or tags (default: all default chains)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// --- Response types ---

type defiPositionsResponse struct {
	Positions    []defiPosition    `json:"positions"`
	Aggregations *defiAggregations `json:"aggregations,omitempty"`
	Warnings     []warningEntry    `json:"warnings,omitempty"`
}

type defiAggregations struct {
	TotalUSDValue float64            `json:"total_usd_value"`
	TotalByChain  map[string]float64 `json:"total_by_chain,omitempty"`
}

// defiPosition is a flat struct matching the polymorphic DefiPosition schema.
// Fields are optional depending on the `type` discriminator.
type defiPosition struct {
	Type    string  `json:"type"`
	ChainID int64   `json:"chain_id"`
	USDVal  float64 `json:"usd_value"`
	Logo    *string `json:"logo,omitempty"`

	// Erc4626 / Tokenized fields
	TokenType               string `json:"token_type,omitempty"`
	Token                   string `json:"token,omitempty"`
	TokenName               string `json:"token_name,omitempty"`
	TokenSymbol             string `json:"token_symbol,omitempty"`
	UnderlyingToken         string `json:"underlying_token,omitempty"`
	UnderlyingTokenName     string `json:"underlying_token_name,omitempty"`
	UnderlyingTokenSymbol   string `json:"underlying_token_symbol,omitempty"`
	UnderlyingTokenDecimals int    `json:"underlying_token_decimals,omitempty"`

	// Erc4626 / Tokenized / UniswapV2 fields
	CalculatedBalance float64 `json:"calculated_balance,omitempty"`
	PriceInUSD        float64 `json:"price_in_usd,omitempty"`

	// UniswapV2 / Nft / NftV4 fields
	Protocol       string  `json:"protocol,omitempty"`
	Pool           string  `json:"pool,omitempty"`
	PoolID         []int   `json:"pool_id,omitempty"`
	PoolManager    string  `json:"pool_manager,omitempty"`
	Salt           []int   `json:"salt,omitempty"`
	Token0         string  `json:"token0,omitempty"`
	Token0Name     string  `json:"token0_name,omitempty"`
	Token0Symbol   string  `json:"token0_symbol,omitempty"`
	Token0Decimals int     `json:"token0_decimals,omitempty"`
	Token1         string  `json:"token1,omitempty"`
	Token1Name     string  `json:"token1_name,omitempty"`
	Token1Symbol   string  `json:"token1_symbol,omitempty"`
	Token1Decimals int     `json:"token1_decimals,omitempty"`
	LPBalance      string  `json:"lp_balance,omitempty"`
	Token0Price    float64 `json:"token0_price,omitempty"`
	Token1Price    float64 `json:"token1_price,omitempty"`

	// Nft / NftV4 concentrated liquidity positions
	Positions []nftPositionDetails `json:"positions,omitempty"`
}

type nftPositionDetails struct {
	TickLower      int     `json:"tick_lower"`
	TickUpper      int     `json:"tick_upper"`
	TokenID        string  `json:"token_id"`
	Token0Price    float64 `json:"token0_price"`
	Token0Holdings float64 `json:"token0_holdings,omitempty"`
	Token0Rewards  float64 `json:"token0_rewards,omitempty"`
	Token1Price    float64 `json:"token1_price"`
	Token1Holdings float64 `json:"token1_holdings,omitempty"`
	Token1Rewards  float64 `json:"token1_rewards,omitempty"`
}

func runDefiPositions(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}

	data, err := client.Get(cmd.Context(), "/beta/evm/defi/positions/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp defiPositionsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Print warnings to stderr.
		printWarnings(cmd, resp.Warnings)

		if len(resp.Positions) == 0 {
			fmt.Fprintln(w, "No DeFi positions found.")
			return nil
		}

		columns := []string{"TYPE", "CHAIN_ID", "PROTOCOL", "USD_VALUE", "DETAILS"}
		rows := make([][]string, len(resp.Positions))
		for i, p := range resp.Positions {
			rows[i] = []string{
				p.Type,
				fmt.Sprintf("%d", p.ChainID),
				p.Protocol,
				output.FormatUSD(p.USDVal),
				positionDetails(p),
			}
		}
		output.PrintTable(w, columns, rows)

		// Print aggregation summary.
		printAggregations(w, resp.Aggregations)

		return nil
	}
}

// positionDetails returns a human-readable summary for a DeFi position,
// varying by position type.
func positionDetails(p defiPosition) string {
	switch p.Type {
	case "Erc4626":
		parts := []string{}
		if p.TokenSymbol != "" {
			parts = append(parts, p.TokenSymbol)
		}
		if p.UnderlyingTokenSymbol != "" {
			parts = append(parts, fmt.Sprintf("-> %s", p.UnderlyingTokenSymbol))
		}
		if p.CalculatedBalance != 0 {
			parts = append(parts, fmt.Sprintf("bal=%.6g", p.CalculatedBalance))
		}
		return strings.Join(parts, " ")

	case "Tokenized":
		parts := []string{}
		if p.TokenType != "" {
			parts = append(parts, p.TokenType)
		}
		if p.TokenSymbol != "" {
			parts = append(parts, p.TokenSymbol)
		}
		if p.CalculatedBalance != 0 {
			parts = append(parts, fmt.Sprintf("bal=%.6g", p.CalculatedBalance))
		}
		return strings.Join(parts, " ")

	case "UniswapV2":
		pair := formatPair(p.Token0Symbol, p.Token1Symbol)
		if p.CalculatedBalance != 0 {
			return fmt.Sprintf("%s bal=%.6g", pair, p.CalculatedBalance)
		}
		return pair

	case "Nft", "NftV4":
		pair := formatPair(p.Token0Symbol, p.Token1Symbol)
		nPos := len(p.Positions)
		if nPos == 1 {
			return fmt.Sprintf("%s (1 position)", pair)
		}
		if nPos > 1 {
			return fmt.Sprintf("%s (%d positions)", pair, nPos)
		}
		return pair

	default:
		return ""
	}
}

// formatPair returns "SYM0/SYM1" or falls back to individual symbols.
func formatPair(sym0, sym1 string) string {
	if sym0 != "" && sym1 != "" {
		return sym0 + "/" + sym1
	}
	if sym0 != "" {
		return sym0
	}
	return sym1
}

// printAggregations prints the aggregation summary after the positions table.
func printAggregations(w io.Writer, agg *defiAggregations) {
	if agg == nil {
		return
	}

	fmt.Fprintf(w, "\nTotal USD Value: %s\n", output.FormatUSD(agg.TotalUSDValue))

	if len(agg.TotalByChain) > 0 {
		fmt.Fprintln(w, "Breakdown by chain:")

		// Sort chain IDs numerically for natural display order.
		chainIDs := make([]string, 0, len(agg.TotalByChain))
		for k := range agg.TotalByChain {
			chainIDs = append(chainIDs, k)
		}
		sort.Slice(chainIDs, func(i, j int) bool {
			a, errA := strconv.Atoi(chainIDs[i])
			b, errB := strconv.Atoi(chainIDs[j])
			if errA != nil || errB != nil {
				return chainIDs[i] < chainIDs[j] // fallback to lexicographic
			}
			return a < b
		})

		for _, cid := range chainIDs {
			fmt.Fprintf(w, "  Chain %s: %s\n", cid, output.FormatUSD(agg.TotalByChain[cid]))
		}
	}
}
