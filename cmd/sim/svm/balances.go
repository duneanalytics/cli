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
		Short: "Get SPL token balances for an SVM wallet address with USD valuations",
		Long: "Return SPL token balances for the given SVM (Solana Virtual Machine) wallet\n" +
			"address. Each balance entry includes the token identity, balance (both raw\n" +
			"and human-readable), current USD price, total USD value, and liquidity data.\n" +
			"Data comes from Dune's real-time index.\n\n" +
			"Supported chains: Solana, Eclipse (default: Solana only).\n\n" +
			"Note: This endpoint is in beta (served under /beta/svm/balances/*).\n\n" +
			"Results are paginated; use --offset with the next_offset value from a\n" +
			"previous response to retrieve additional pages.\n\n" +
			"Examples:\n" +
			"  dune sim svm balances 86xCnPeV69n6t3DnyGvkKobf9FdN2H9oiVDdaMpo2MMY\n" +
			"  dune sim svm balances 86xCnPeV... --chains solana,eclipse\n" +
			"  dune sim svm balances 86xCnPeV... --limit 50 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runBalances,
	}

	cmd.Flags().String("chains", "", "Restrict to specific SVM chains (comma-separated): 'solana', 'eclipse' (default: solana only)")
	cmd.Flags().Int("limit", 0, "Maximum number of balance entries to return per page (1-1000, default: 1000)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
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
			bal := b.Balance
			if bal == "" {
				bal = b.Amount
			}
			rows[i] = []string{
				b.Chain,
				b.Symbol,
				bal,
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
