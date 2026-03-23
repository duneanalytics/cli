package svm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewTransactionsCmd returns the `sim svm transactions` command.
func NewTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions <address>",
		Short: "Get Solana transaction history for an SVM wallet address",
		Long: "Return transaction history for the given SVM (Solana Virtual Machine) wallet\n" +
			"address. Transactions are returned in reverse-chronological order by block slot.\n\n" +
			"Note: This endpoint is in beta (served under /beta/svm/transactions/*).\n\n" +
			"Response fields per transaction:\n" +
			"  - chain: network name (e.g. 'solana')\n" +
			"  - block_slot: Solana slot number of the block\n" +
			"  - block_time: block timestamp (microseconds since Unix epoch; displayed\n" +
			"    as UTC datetime in text mode)\n" +
			"  - address: the queried wallet address\n" +
			"  - raw_transaction: full Solana transaction object including signatures,\n" +
			"    instructions, and account keys (only in JSON output with -o json)\n\n" +
			"The text table shows chain, block slot, block time, and the first transaction\n" +
			"signature. Use -o json to access the complete raw_transaction structure\n" +
			"with all instructions and log messages.\n\n" +
			"Results are paginated; use --offset with the next_offset value from a\n" +
			"previous response to retrieve additional pages.\n\n" +
			"Examples:\n" +
			"  dune sim svm transactions 86xCnPeV69n6t3DnyGvkKobf9FdN2H9oiVDdaMpo2MMY\n" +
			"  dune sim svm transactions 86xCnPeV... --limit 20\n" +
			"  dune sim svm transactions 86xCnPeV... -o json",
		Args: cobra.ExactArgs(1),
		RunE: runTransactions,
	}

	cmd.Flags().Int("limit", 0, "Maximum number of transactions to return per page (1-1000, default: 100)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// --- Response types ---

type svmTransactionsResponse struct {
	NextOffset   string           `json:"next_offset,omitempty"`
	Transactions []svmTransaction `json:"transactions"`
}

type svmTransaction struct {
	Address        string          `json:"address"`
	BlockSlot      json.Number     `json:"block_slot"`
	BlockTime      json.Number     `json:"block_time"`
	Chain          string          `json:"chain"`
	RawTransaction json.RawMessage `json:"raw_transaction,omitempty"`
}

func runTransactions(cmd *cobra.Command, args []string) error {
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/beta/svm/transactions/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp svmTransactionsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if len(resp.Transactions) == 0 {
			fmt.Fprintln(w, "No transactions found.")
			return nil
		}

		columns := []string{"CHAIN", "BLOCK_SLOT", "BLOCK_TIME", "TX_SIGNATURE"}
		rows := make([][]string, len(resp.Transactions))
		for i, tx := range resp.Transactions {
			rows[i] = []string{
				tx.Chain,
				tx.BlockSlot.String(),
				formatBlockTime(tx.BlockTime),
				extractSignature(tx.RawTransaction),
			}
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}

// formatBlockTime converts a block_time (microseconds since epoch) to a
// human-readable UTC timestamp.
func formatBlockTime(bt json.Number) string {
	us, err := bt.Int64()
	if err != nil {
		return bt.String()
	}
	// block_time is in microseconds.
	t := time.Unix(0, us*int64(time.Microsecond))
	return t.UTC().Format("2006-01-02 15:04:05")
}

// extractSignature pulls the first transaction signature from raw_transaction.
// Returns the signature string or an empty string if unavailable.
func extractSignature(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var rt struct {
		Transaction struct {
			Signatures []string `json:"signatures"`
		} `json:"transaction"`
	}
	if err := json.Unmarshal(raw, &rt); err != nil {
		return ""
	}
	if len(rt.Transaction.Signatures) > 0 {
		return rt.Transaction.Signatures[0]
	}
	return ""
}
