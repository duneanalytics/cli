package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewTokenInfoCmd returns the `sim evm token-info` command.
func NewTokenInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-info <address>",
		Short: "Get token metadata, pricing, supply, and market cap for a token address",
		Long: "Return metadata and real-time pricing for a token contract address on a\n" +
			"specified EVM chain. Use the literal string 'native' as the address to query\n" +
			"the chain's native asset (e.g. ETH on chain 1, MATIC on chain 137).\n\n" +
			"The --chain-ids flag is required and should specify a single chain ID.\n\n" +
			"Response fields per token:\n" +
			"  - chain, chain_id: network name and numeric ID\n" +
			"  - symbol, name, decimals: token identity and precision\n" +
			"  - price_usd: current USD price from on-chain DEX liquidity pools\n" +
			"  - total_supply: total token supply (raw, divide by 10^decimals)\n" +
			"  - market_cap: estimated market capitalization in USD\n" +
			"  - logo: URL to the token's icon image\n" +
			"  - historical_prices: past USD prices at requested hour offsets\n\n" +
			"This command is useful for looking up token details before querying\n" +
			"balances or activity. For wallet-scoped token data, use\n" +
			"'dune sim evm balances' or 'dune sim evm balance'.\n\n" +
			"Examples:\n" +
			"  dune sim evm token-info native --chain-ids 1\n" +
			"  dune sim evm token-info 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 --chain-ids 8453\n" +
			"  dune sim evm token-info native --chain-ids 1 --historical-prices 720,168,24 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runTokenInfo,
	}

	cmd.Flags().String("chain-ids", "", "Numeric EVM chain ID to query (required, single value, e.g. '1' for Ethereum, '8453' for Base)")
	cmd.Flags().String("historical-prices", "", "Include historical USD prices at the specified hour offsets from now (comma-separated, e.g. '720,168,24' for 30d, 7d, 1d ago)")
	cmd.Flags().Int("limit", 0, "Maximum number of token entries to return (default: server-determined)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
	_ = cmd.MarkFlagRequired("chain-ids")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type tokensResponse struct {
	ContractAddress string         `json:"contract_address"`
	Tokens          []tokenInfo    `json:"tokens"`
	Warnings        []warningEntry `json:"warnings,omitempty"`
	NextOffset      string         `json:"next_offset,omitempty"`
}

type tokenInfo struct {
	Chain            string            `json:"chain"`
	ChainID          int64             `json:"chain_id"`
	Symbol           string            `json:"symbol,omitempty"`
	Name             string            `json:"name,omitempty"`
	Decimals         int               `json:"decimals,omitempty"`
	PriceUSD         float64           `json:"price_usd"`
	HistoricalPrices []historicalPrice `json:"historical_prices,omitempty"`
	TotalSupply      string            `json:"total_supply,omitempty"`
	MarketCap        float64           `json:"market_cap,omitempty"`
	Logo             string            `json:"logo,omitempty"`
}

func runTokenInfo(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetString("historical-prices"); v != "" {
		params.Set("historical_prices", v)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/token-info/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp tokensResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Print warnings to stderr.
		printWarnings(cmd, resp.Warnings)

		if len(resp.Tokens) == 0 {
			fmt.Fprintln(w, "No token info found.")
			return nil
		}

		// Key-value display for each token entry.
		for i, t := range resp.Tokens {
			if i > 0 {
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "Chain:        %s (ID: %d)\n", t.Chain, t.ChainID)
			if t.Symbol != "" {
				fmt.Fprintf(w, "Symbol:       %s\n", t.Symbol)
			}
			if t.Name != "" {
				fmt.Fprintf(w, "Name:         %s\n", t.Name)
			}
			fmt.Fprintf(w, "Decimals:     %d\n", t.Decimals)
			fmt.Fprintf(w, "Price USD:    %s\n", output.FormatUSD(t.PriceUSD))
			if t.TotalSupply != "" {
				fmt.Fprintf(w, "Total Supply: %s\n", t.TotalSupply)
			}
			if t.MarketCap > 0 {
				fmt.Fprintf(w, "Market Cap:   %s\n", output.FormatUSD(t.MarketCap))
			}
			if t.Logo != "" {
				fmt.Fprintf(w, "Logo:         %s\n", t.Logo)
			}
			for _, hp := range t.HistoricalPrices {
				fmt.Fprintf(w, "Price %dh ago: %s\n", hp.OffsetHours, output.FormatUSD(hp.PriceUSD))
			}
		}

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}
