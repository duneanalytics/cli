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
		Short: "Get token metadata and pricing",
		Long: "Return metadata and pricing for a token contract address (or \"native\" for the\n" +
			"chain's native asset) on a specified chain.\n\n" +
			"Examples:\n" +
			"  dune sim evm token-info native --chain-ids 1\n" +
			"  dune sim evm token-info 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 --chain-ids 8453\n" +
			"  dune sim evm token-info native --chain-ids 1 --historical-prices 720,168,24 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runTokenInfo,
	}

	cmd.Flags().String("chain-ids", "", "Chain ID (required)")
	cmd.Flags().String("historical-prices", "", "Hour offsets for historical prices (e.g. 720,168,24)")
	cmd.Flags().Int("limit", 0, "Max results")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
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
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

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
			fmt.Fprintf(w, "Price USD:    %s\n", formatUSD(t.PriceUSD))
			if t.TotalSupply != "" {
				fmt.Fprintf(w, "Total Supply: %s\n", t.TotalSupply)
			}
			if t.MarketCap > 0 {
				fmt.Fprintf(w, "Market Cap:   %s\n", formatUSD(t.MarketCap))
			}
			if t.Logo != "" {
				fmt.Fprintf(w, "Logo:         %s\n", t.Logo)
			}
			for _, hp := range t.HistoricalPrices {
				fmt.Fprintf(w, "Price %dh ago: %s\n", hp.OffsetHours, formatUSD(hp.PriceUSD))
			}
		}

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}
