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
	TotalUSDValue float64            `json:"total_value_usd"`
	TotalByChain  map[string]float64 `json:"total_by_chain,omitempty"`
}

// defiTokenInfo represents a token object returned by the API with address,
// name, symbol, and optional numeric fields depending on position type.
type defiTokenInfo struct {
	Address  string  `json:"address,omitempty"`
	Name     string  `json:"name,omitempty"`
	Symbol   string  `json:"symbol,omitempty"`
	Decimals int     `json:"decimals,omitempty"`
	Holdings float64 `json:"holdings,omitempty"`
	PriceUSD float64 `json:"price_usd,omitempty"`
}

// nftTokenDetails holds per-token data inside an NFT concentrated-liquidity position.
type nftTokenDetails struct {
	PriceUSD float64 `json:"price_usd"`
	Holdings float64 `json:"holdings,omitempty"`
	Rewards  float64 `json:"rewards,omitempty"`
}

// defiPosition matches the polymorphic DefiPosition schema returned by the API.
// Fields are optional depending on the `type` discriminator.
type defiPosition struct {
	Type    string  `json:"type"`
	Chain   string  `json:"chain,omitempty"`
	ChainID int64   `json:"chain_id"`
	USDVal  float64 `json:"value_usd"`
	Logo    *string `json:"logo,omitempty"`

	// Erc4626 / Tokenized fields
	TokenType       string         `json:"token_type,omitempty"`
	Token           *defiTokenInfo `json:"token,omitempty"`
	UnderlyingToken *defiTokenInfo `json:"underlying_token,omitempty"`
	LendingPool     string         `json:"lending_pool,omitempty"`

	// Erc4626 / Tokenized / UniswapV2 fields
	Balance  float64 `json:"balance,omitempty"`
	PriceUSD float64 `json:"price_usd,omitempty"`

	// UniswapV2 / Nft / NftV4 fields
	Protocol    string         `json:"protocol,omitempty"`
	Pool        string         `json:"pool,omitempty"`
	PoolID      string         `json:"pool_id,omitempty"`
	PoolManager string         `json:"pool_manager,omitempty"`
	Salt        string         `json:"salt,omitempty"`
	Token0      *defiTokenInfo `json:"token0,omitempty"`
	Token1      *defiTokenInfo `json:"token1,omitempty"`
	LPBalance   string         `json:"lp_balance,omitempty"`

	// Nft / NftV4 concentrated liquidity positions
	Positions []nftPositionDetails `json:"positions,omitempty"`
}

type nftPositionDetails struct {
	TickLower int              `json:"tick_lower"`
	TickUpper int              `json:"tick_upper"`
	TokenID   string           `json:"token_id"`
	Token0    *nftTokenDetails `json:"token0,omitempty"`
	Token1    *nftTokenDetails `json:"token1,omitempty"`
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
// tokenSymbol safely extracts the symbol from a token info pointer.
func tokenSymbol(t *defiTokenInfo) string {
	if t == nil {
		return ""
	}
	return t.Symbol
}

func positionDetails(p defiPosition) string {
	switch p.Type {
	case "Erc4626":
		parts := []string{}
		if sym := tokenSymbol(p.Token); sym != "" {
			parts = append(parts, sym)
		}
		if sym := tokenSymbol(p.UnderlyingToken); sym != "" {
			parts = append(parts, fmt.Sprintf("-> %s", sym))
		}
		if p.Balance != 0 {
			parts = append(parts, fmt.Sprintf("bal=%.6g", p.Balance))
		}
		return strings.Join(parts, " ")

	case "Tokenized":
		parts := []string{}
		if p.TokenType != "" {
			parts = append(parts, p.TokenType)
		}
		if sym := tokenSymbol(p.Token); sym != "" {
			parts = append(parts, sym)
		}
		if p.Balance != 0 {
			parts = append(parts, fmt.Sprintf("bal=%.6g", p.Balance))
		}
		return strings.Join(parts, " ")

	case "UniswapV2":
		pair := formatPair(tokenSymbol(p.Token0), tokenSymbol(p.Token1))
		if p.Balance != 0 {
			return fmt.Sprintf("%s bal=%.6g", pair, p.Balance)
		}
		return pair

	case "Nft", "NftV4":
		pair := formatPair(tokenSymbol(p.Token0), tokenSymbol(p.Token1))
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
