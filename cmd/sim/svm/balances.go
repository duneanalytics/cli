package svm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewBalancesCmd returns the `sim svm balances` command.
func NewBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances <address>",
		Short: "Get SVM token balances for a wallet address",
		Long: "Return token balances for the given SVM wallet address across\n" +
			"Solana and Eclipse chains, including USD valuations.\n\n" +
			"Examples:\n" +
			"  dune sim svm balances 86xCnPeV69n6t3DnyGvkKobf9FdN2H9oiVDdaMpo2MMY\n" +
			"  dune sim svm balances 86xCnPeV... --chains solana,eclipse\n" +
			"  dune sim svm balances 86xCnPeV... --limit 50 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalances,
	}

	cmd.Flags().String("chains", "", "Comma-separated chains: solana, eclipse (default: solana)")
	cmd.Flags().Int("limit", 0, "Max results (1-1000, default 1000)")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// --- Response types ---

type svmBalancesResponse struct {
	ProcessingTimeMs float64           `json:"processing_time_ms,omitempty"`
	WalletAddress    string            `json:"wallet_address"`
	NextOffset       string            `json:"next_offset,omitempty"`
	BalancesCount    float64           `json:"balances_count,omitempty"`
	Balances         []svmBalanceEntry `json:"balances"`
}

type svmBalanceEntry struct {
	Chain         string  `json:"chain"`
	Address       string  `json:"address"`
	Amount        string  `json:"amount"`
	Balance       string  `json:"balance,omitempty"`
	RawBalance    string  `json:"raw_balance,omitempty"`
	ValueUSD      float64 `json:"value_usd,omitempty"`
	ProgramID     *string `json:"program_id,omitempty"`
	Decimals      float64 `json:"decimals,omitempty"`
	TotalSupply   string  `json:"total_supply,omitempty"`
	Name          string  `json:"name,omitempty"`
	Symbol        string  `json:"symbol,omitempty"`
	URI           *string `json:"uri,omitempty"`
	PriceUSD      float64 `json:"price_usd,omitempty"`
	LiquidityUSD  float64 `json:"liquidity_usd,omitempty"`
	PoolType      *string `json:"pool_type,omitempty"`
	PoolAddress   *string `json:"pool_address,omitempty"`
	MintAuthority *string `json:"mint_authority,omitempty"`
}

func runBalances(cmd *cobra.Command, args []string) error {
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chains"); v != "" {
		params.Set("chains", v)
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/beta/svm/balances/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp svmBalancesResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if len(resp.Balances) == 0 {
			fmt.Fprintln(w, "No balances found.")
			return nil
		}

		columns := []string{"CHAIN", "SYMBOL", "BALANCE", "PRICE_USD", "VALUE_USD"}
		rows := make([][]string, len(resp.Balances))
		for i, b := range resp.Balances {
			rows[i] = []string{
				b.Chain,
				b.Symbol,
				b.Balance,
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

// formatUSD formats a USD value for display.
func formatUSD(v float64) string {
	if v == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", v)
}
