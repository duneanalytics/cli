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
		Short: "Get the balance of a single token for a wallet on one chain",
		Long: "Return the balance of a specific token for the given wallet address on a\n" +
			"single EVM chain. This is a targeted lookup that returns exactly one balance\n" +
			"entry, unlike 'dune sim evm balances' which returns all tokens across chains.\n\n" +
			"Both --token and --chain-ids are required. Use the literal string 'native'\n" +
			"as the --token value to query the chain's native asset (e.g. ETH on Ethereum,\n" +
			"MATIC on Polygon), or pass an ERC20 contract address.\n\n" +
			"Response fields:\n" +
			"  - chain, chain_id: network name and numeric ID\n" +
			"  - address, symbol, name, decimals: token identity\n" +
			"  - amount: raw balance (divide by 10^decimals for human-readable)\n" +
			"  - price_usd, value_usd: current pricing and total value\n\n" +
			"For multi-token or multi-chain lookups, use 'dune sim evm balances' instead.\n\n" +
			"Examples:\n" +
			"  dune sim evm balance 0xd8da... --token native --chain-ids 1\n" +
			"  dune sim evm balance 0xd8da... --token 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 --chain-ids 8453\n" +
			"  dune sim evm balance 0xd8da... --token native --chain-ids 1 --historical-prices 168,24 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalance,
	}

	cmd.Flags().String("token", "", "Token to query: an ERC20 contract address (0x...) or the literal string 'native' for the chain's native asset (required)")
	cmd.Flags().String("chain-ids", "", "Numeric EVM chain ID to query (required, single value, e.g. '1' for Ethereum, '8453' for Base)")
	cmd.Flags().String("metadata", "", "Request additional metadata fields in the response (comma-separated): 'logo' (token icon URL), 'url' (project website), 'pools' (liquidity pool details)")
	cmd.Flags().String("historical-prices", "", "Include historical USD prices at the specified hour offsets from now (comma-separated, e.g. '720,168,24' for 30d, 7d, 1d ago)")
	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("chain-ids")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runBalance(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
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

		printBalanceErrors(cmd, resp.Errors)

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
		fmt.Fprintf(w, "Price USD: %s\n", output.FormatUSD(b.PriceUSD))
		fmt.Fprintf(w, "Value USD: %s\n", output.FormatUSD(b.ValueUSD))

		return nil
	}
}
