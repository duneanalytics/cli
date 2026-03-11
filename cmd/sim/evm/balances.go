package evm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewBalancesCmd returns the `sim evm balances` command.
func NewBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances <address>",
		Short: "Get EVM token balances for a wallet address",
		Long: "Return native and ERC20 token balances for the given wallet address\n" +
			"across supported EVM chains, including USD valuations.\n\n" +
			"Examples:\n" +
			"  dune sim evm balances 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm balances 0xd8da... --chain-ids 1,8453 --exclude-spam\n" +
			"  dune sim evm balances 0xd8da... --historical-prices 168 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalances,
	}

	cmd.Flags().String("chain-ids", "", "Comma-separated chain IDs or tags (default: all default chains)")
	cmd.Flags().String("filters", "", "Token filter: erc20 or native")
	cmd.Flags().String("asset-class", "", "Asset class filter: stablecoin")
	cmd.Flags().String("metadata", "", "Extra metadata fields: logo,url,pools")
	cmd.Flags().Bool("exclude-spam", false, "Exclude tokens with <100 USD liquidity")
	cmd.Flags().String("historical-prices", "", "Hour offsets for historical prices (e.g. 720,168,24)")
	cmd.Flags().Int("limit", 0, "Max results (1-1000)")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type balancesResponse struct {
	WalletAddress string         `json:"wallet_address"`
	Balances      []balanceEntry `json:"balances"`
	Errors        *balanceErrors `json:"errors,omitempty"`
	NextOffset    string         `json:"next_offset,omitempty"`
	Warnings      []warningEntry `json:"warnings,omitempty"`
	RequestTime   string         `json:"request_time,omitempty"`
	ResponseTime  string         `json:"response_time,omitempty"`
}

type balanceErrors struct {
	ErrorMessage string             `json:"error_message,omitempty"`
	TokenErrors  []balanceErrorInfo `json:"token_errors,omitempty"`
}

type balanceErrorInfo struct {
	ChainID     int64  `json:"chain_id"`
	Address     string `json:"address"`
	Description string `json:"description,omitempty"`
}

type balanceEntry struct {
	Chain            string            `json:"chain"`
	ChainID          int64             `json:"chain_id"`
	Address          string            `json:"address"`
	Amount           string            `json:"amount"`
	Symbol           string            `json:"symbol"`
	Name             string            `json:"name"`
	Decimals         int               `json:"decimals"`
	PriceUSD         float64           `json:"price_usd"`
	ValueUSD         float64           `json:"value_usd"`
	PoolSize         float64           `json:"pool_size"`
	LowLiquidity     bool              `json:"low_liquidity"`
	HistoricalPrices []historicalPrice `json:"historical_prices,omitempty"`
	TokenMetadata    *balanceTokenMeta `json:"token_metadata,omitempty"`
	Pool             *poolMetadata     `json:"pool,omitempty"`
}

type historicalPrice struct {
	OffsetHours int     `json:"offset_hours"`
	PriceUSD    float64 `json:"price_usd"`
}

type balanceTokenMeta struct {
	Logo string `json:"logo,omitempty"`
	URL  string `json:"url,omitempty"`
}

type poolMetadata struct {
	PoolType string `json:"pool_type"`
	Address  string `json:"address"`
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
}

type warningEntry struct {
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	ChainIDs []int64 `json:"chain_ids,omitempty"`
	DocsURL  string  `json:"docs_url,omitempty"`
}

func runBalances(cmd *cobra.Command, args []string) error {
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetString("filters"); v != "" {
		params.Set("filters", v)
	}
	if v, _ := cmd.Flags().GetString("asset-class"); v != "" {
		params.Set("asset_class", v)
	}
	if v, _ := cmd.Flags().GetString("metadata"); v != "" {
		params.Set("metadata", v)
	}
	if v, _ := cmd.Flags().GetBool("exclude-spam"); v {
		params.Set("exclude_spam_tokens", "true")
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

	data, err := client.Get(cmd.Context(), "/v1/evm/balances/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp balancesResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Print errors and warnings to stderr.
		printBalanceErrors(cmd, resp.Errors)
		printWarnings(cmd, resp.Warnings)

		columns := []string{"CHAIN", "SYMBOL", "AMOUNT", "PRICE_USD", "VALUE_USD"}
		rows := make([][]string, len(resp.Balances))
		for i, b := range resp.Balances {
			rows[i] = []string{
				b.Chain,
				b.Symbol,
				formatAmount(b.Amount, b.Decimals),
				formatUSD(b.PriceUSD),
				formatUSD(b.ValueUSD),
			}
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}

// printWarnings writes API warnings to stderr.
func printWarnings(cmd *cobra.Command, warnings []warningEntry) {
	if len(warnings) == 0 {
		return
	}
	stderr := cmd.ErrOrStderr()
	for _, w := range warnings {
		fmt.Fprintf(stderr, "Warning: %s\n", w.Message)
		if len(w.ChainIDs) > 0 {
			ids := make([]string, len(w.ChainIDs))
			for i, id := range w.ChainIDs {
				ids[i] = fmt.Sprintf("%d", id)
			}
			fmt.Fprintf(stderr, "  Unsupported chain IDs: %s\n", strings.Join(ids, ", "))
		}
		if w.DocsURL != "" {
			fmt.Fprintf(stderr, "  See %s\n", w.DocsURL)
		}
	}
	fmt.Fprintln(stderr)
}

// formatAmount converts a raw token amount string with decimals to a
// human-readable decimal representation.
func formatAmount(raw string, decimals int) string {
	if decimals <= 0 || raw == "" || raw == "0" {
		return raw
	}

	// Pad with leading zeros if the raw string is shorter than decimals.
	for len(raw) <= decimals {
		raw = "0" + raw
	}

	intPart := raw[:len(raw)-decimals]
	fracPart := raw[len(raw)-decimals:]

	// Trim trailing zeros from the fractional part, keep up to 6 digits.
	if len(fracPart) > 6 {
		fracPart = fracPart[:6]
	}
	fracPart = strings.TrimRight(fracPart, "0")

	if fracPart == "" {
		return intPart
	}
	return intPart + "." + fracPart
}

// formatUSD formats a USD value for display.
func formatUSD(v float64) string {
	if v == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", v)
}

// printBalanceErrors writes balance-level errors to stderr.
func printBalanceErrors(cmd *cobra.Command, errs *balanceErrors) {
	if errs == nil {
		return
	}
	stderr := cmd.ErrOrStderr()
	if errs.ErrorMessage != "" {
		fmt.Fprintf(stderr, "Error: %s\n", errs.ErrorMessage)
	}
	for _, e := range errs.TokenErrors {
		fmt.Fprintf(stderr, "  chain_id=%d address=%s", e.ChainID, e.Address)
		if e.Description != "" {
			fmt.Fprintf(stderr, " — %s", e.Description)
		}
		fmt.Fprintln(stderr)
	}
	if errs.ErrorMessage != "" || len(errs.TokenErrors) > 0 {
		fmt.Fprintln(stderr)
	}
}
