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
		Short: "Get EVM token balances for a wallet address across multiple chains",
		Long: "Return native and ERC20 token balances for the given wallet address across\n" +
			"supported EVM chains. Each balance entry includes the token identity, raw\n" +
			"balance amount, current USD price, total USD value, and liquidity indicators.\n" +
			"Data comes from Dune's real-time index.\n\n" +
			"By default, queries all chains tagged 'default'. Use --chain-ids to restrict\n" +
			"to specific networks. Run 'dune sim evm supported-chains' to see valid IDs.\n\n" +
			"For a single-token balance lookup, use 'dune sim evm balance' instead.\n" +
			"For stablecoin-only balances, use 'dune sim evm stablecoins'.\n\n" +
			"Results are paginated; use --offset with the next_offset value from a\n" +
			"previous response to retrieve additional pages.\n\n" +
			"Examples:\n" +
			"  dune sim evm balances 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm balances 0xd8da... --chain-ids 1,8453 --exclude-spam\n" +
			"  dune sim evm balances 0xd8da... --filters erc20 --metadata logo,url\n" +
			"  dune sim evm balances 0xd8da... --historical-prices 720,168,24 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalances,
	}

	addBalanceFlags(cmd)
	cmd.Flags().String("asset-class", "", "Filter by asset classification: 'stablecoin' (returns only stablecoins like USDC, USDT, DAI); prefer 'dune sim evm stablecoins' as a shorthand")

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
	ErrorMessage string          `json:"error_message,omitempty"`
	TokenErrors  []apiChainError `json:"token_errors,omitempty"`
}

// apiChainError is a per-chain error returned by several Sim API endpoints
// (balances, transactions, etc.). It is intentionally shared across commands.
type apiChainError struct {
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
	return runBalancesEndpoint(cmd, args, "/v1/evm/balances/", "")
}

// addBalanceFlags registers the common flags shared by the balances and
// stablecoins commands.
func addBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().String("chain-ids", "", "Restrict to specific chains by numeric ID or tag name (comma-separated, e.g. '1,8453' or 'default'); defaults to all chains tagged 'default'. Run 'dune sim evm supported-chains' for valid values")
	cmd.Flags().String("filters", "", "Filter by token standard: 'erc20' (only ERC20 tokens) or 'native' (only native chain assets like ETH)")
	cmd.Flags().String("metadata", "", "Request additional metadata fields in the response (comma-separated): 'logo' (token icon URL), 'url' (project website), 'pools' (liquidity pool details)")
	cmd.Flags().Bool("exclude-spam", false, "Exclude low-liquidity tokens (less than $100 USD pool size) commonly associated with spam airdrops")
	cmd.Flags().Bool("exclude-unpriced", true, "Exclude tokens without a USD price (default: true); pass --exclude-unpriced=false to include them")
	cmd.Flags().String("historical-prices", "", "Include historical USD prices at the specified hour offsets from now (comma-separated, e.g. '720,168,24' for 30d, 7d, 1d ago)")
	cmd.Flags().Int("limit", 0, "Maximum number of balance entries to return per page (1-1000, default: server-determined)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
	output.AddFormatFlag(cmd, "text")
}

// runBalancesEndpoint is the shared run implementation for the balances and
// stablecoins commands. The final API path is built as:
//
//	pathPrefix + address + pathSuffix
//
// For example "/v1/evm/balances/" + addr + "" for balances,
// or "/v1/evm/balances/" + addr + "/stablecoins" for stablecoins.
func runBalancesEndpoint(cmd *cobra.Command, args []string, pathPrefix, pathSuffix string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetString("filters"); v != "" {
		params.Set("filters", v)
	}
	// asset-class is only registered on the balances command; silently ignored
	// when the flag is absent.
	if v, _ := cmd.Flags().GetString("asset-class"); v != "" {
		params.Set("asset_class", v)
	}
	if v, _ := cmd.Flags().GetString("metadata"); v != "" {
		params.Set("metadata", v)
	}
	if v, _ := cmd.Flags().GetBool("exclude-spam"); v {
		params.Set("exclude_spam_tokens", "true")
	}
	if v, _ := cmd.Flags().GetBool("exclude-unpriced"); v {
		params.Set("exclude_unpriced", "true")
	} else {
		params.Set("exclude_unpriced", "false")
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

	data, err := client.Get(cmd.Context(), pathPrefix+address+pathSuffix, params)
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
				output.FormatUSD(b.PriceUSD),
				output.FormatUSD(b.ValueUSD),
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

// printBalanceErrors writes balance-level errors to stderr.
func printBalanceErrors(cmd *cobra.Command, errs *balanceErrors) {
	if errs == nil {
		return
	}
	printAPIChainErrors(cmd, errs.ErrorMessage, errs.TokenErrors)
}

// printAPIChainErrors is a shared helper that writes per-chain API errors to
// stderr. It is used by both balance and transaction commands to avoid
// duplicating the same formatting logic.
func printAPIChainErrors(cmd *cobra.Command, msg string, errs []apiChainError) {
	if msg == "" && len(errs) == 0 {
		return
	}
	stderr := cmd.ErrOrStderr()
	if msg != "" {
		fmt.Fprintf(stderr, "Error: %s\n", msg)
	}
	for _, e := range errs {
		fmt.Fprintf(stderr, "  chain_id=%d address=%s", e.ChainID, e.Address)
		if e.Description != "" {
			fmt.Fprintf(stderr, " — %s", e.Description)
		}
		fmt.Fprintln(stderr)
	}
	fmt.Fprintln(stderr)
}
