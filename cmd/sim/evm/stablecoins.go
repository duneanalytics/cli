package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewStablecoinsCmd returns the `sim evm stablecoins` command.
func NewStablecoinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stablecoins <address>",
		Short: "Get stablecoin balances for a wallet address",
		Long: "Return stablecoin balances for the given wallet address across supported\n" +
			"EVM chains, including USD valuations.\n\n" +
			"Examples:\n" +
			"  dune sim evm stablecoins 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm stablecoins 0xd8da... --chain-ids 1,8453\n" +
			"  dune sim evm stablecoins 0xd8da... -o json",
		Args: cobra.ExactArgs(1),
		RunE: runStablecoins,
	}

	cmd.Flags().String("chain-ids", "", "Comma-separated chain IDs or tags (default: all default chains)")
	cmd.Flags().String("filters", "", "Token filter: erc20 or native")
	cmd.Flags().String("metadata", "", "Extra metadata fields: logo,url,pools")
	cmd.Flags().Bool("exclude-spam", false, "Exclude tokens with <100 USD liquidity")
	cmd.Flags().String("historical-prices", "", "Hour offsets for historical prices (e.g. 720,168,24)")
	cmd.Flags().Int("limit", 0, "Max results (1-1000)")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runStablecoins(cmd *cobra.Command, args []string) error {
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

	data, err := client.Get(cmd.Context(), "/v1/evm/balances/"+address+"/stablecoins", params)
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
