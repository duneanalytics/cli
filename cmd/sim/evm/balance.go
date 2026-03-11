package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewBalanceCmd returns the `sim evm balance` command (single token).
func NewBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance <wallet_address>",
		Short: "Get balance for a single token",
		Long: "Return the balance of a single token for the given wallet address on one chain.\n" +
			"Use \"native\" as the --token value to query the native asset (e.g. ETH).\n\n" +
			"Examples:\n" +
			"  dune sim evm balance 0xd8da... --token native --chain-ids 1\n" +
			"  dune sim evm balance 0xd8da... --token 0xa0b8...eb48 --chain-ids 8453\n" +
			"  dune sim evm balance 0xd8da... --token native --chain-ids 1 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalance,
	}

	cmd.Flags().String("token", "", "Token contract address or \"native\" (required)")
	cmd.Flags().String("chain-ids", "", "Chain ID (required)")
	cmd.Flags().String("metadata", "", "Extra metadata fields: logo,url,pools")
	cmd.Flags().String("historical-prices", "", "Hour offsets for historical prices (e.g. 720,168,24)")
	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("chain-ids")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runBalance(cmd *cobra.Command, args []string) error {
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

	address := args[0]
	tokenAddress, _ := cmd.Flags().GetString("token")

	params := url.Values{}
	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetString("metadata"); v != "" {
		params.Set("metadata", v)
	}
	if v, _ := cmd.Flags().GetString("historical-prices"); v != "" {
		params.Set("historical_prices", v)
	}

	path := fmt.Sprintf("/v1/evm/balances/%s/token/%s", address, tokenAddress)
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
		var resp balancesResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if len(resp.Balances) == 0 {
			fmt.Fprintln(w, "No balance found.")
			return nil
		}

		b := resp.Balances[0]
		fmt.Fprintf(w, "Chain:     %s (ID: %d)\n", b.Chain, b.ChainID)
		fmt.Fprintf(w, "Token:     %s\n", b.Address)
		fmt.Fprintf(w, "Symbol:    %s\n", b.Symbol)
		if b.Name != "" {
			fmt.Fprintf(w, "Name:      %s\n", b.Name)
		}
		fmt.Fprintf(w, "Decimals:  %d\n", b.Decimals)
		fmt.Fprintf(w, "Amount:    %s\n", formatAmount(b.Amount, b.Decimals))
		fmt.Fprintf(w, "Price USD: %s\n", formatUSD(b.PriceUSD))
		fmt.Fprintf(w, "Value USD: %s\n", formatUSD(b.ValueUSD))

		return nil
	}
}
