package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewTransactionsCmd returns the `sim evm transactions` command.
func NewTransactionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions <address>",
		Short: "Get EVM transactions for a wallet address",
		Long: "Return transaction history for the given wallet address across supported EVM chains.\n" +
			"Use --decode with -o json to include decoded function calls and event logs.\n\n" +
			"Examples:\n" +
			"  dune sim evm transactions 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm transactions 0xd8da... --chain-ids 1 --decode -o json\n" +
			"  dune sim evm transactions 0xd8da... --limit 50 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runTransactions,
	}

	cmd.Flags().String("chain-ids", "", "Comma-separated chain IDs or tags (default: all default chains)")
	cmd.Flags().Bool("decode", false, "Include decoded transaction data and logs (use with -o json)")
	cmd.Flags().Int("limit", 0, "Max results (1-100)")
	cmd.Flags().String("offset", "", "Pagination cursor from previous response")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type transactionsResponse struct {
	WalletAddress string             `json:"wallet_address"`
	Transactions  []transactionTx    `json:"transactions"`
	Errors        *transactionErrors `json:"errors,omitempty"`
	NextOffset    string             `json:"next_offset,omitempty"`
	Warnings      []warningEntry     `json:"warnings,omitempty"`
	RequestTime   string             `json:"request_time,omitempty"`
	ResponseTime  string             `json:"response_time,omitempty"`
}

type transactionErrors struct {
	ErrorMessage      string                 `json:"error_message,omitempty"`
	TransactionErrors []transactionErrorInfo `json:"transaction_errors,omitempty"`
}

type transactionErrorInfo struct {
	ChainID     int64  `json:"chain_id"`
	Address     string `json:"address"`
	Description string `json:"description,omitempty"`
}

type transactionTx struct {
	Address              string           `json:"address"`
	BlockHash            string           `json:"block_hash"`
	BlockNumber          json.Number      `json:"block_number"`
	BlockTime            string           `json:"block_time"`
	BlockVersion         int              `json:"block_version,omitempty"`
	Chain                string           `json:"chain"`
	From                 string           `json:"from"`
	To                   string           `json:"to"`
	Data                 string           `json:"data,omitempty"`
	GasPrice             string           `json:"gas_price,omitempty"`
	Hash                 string           `json:"hash"`
	Index                json.Number      `json:"index,omitempty"`
	MaxFeePerGas         string           `json:"max_fee_per_gas,omitempty"`
	MaxPriorityFeePerGas string           `json:"max_priority_fee_per_gas,omitempty"`
	Nonce                string           `json:"nonce,omitempty"`
	TransactionType      string           `json:"transaction_type,omitempty"`
	Value                string           `json:"value"`
	Decoded              *decodedCall     `json:"decoded,omitempty"`
	Logs                 []transactionLog `json:"logs,omitempty"`
}

type decodedCall struct {
	Name   string         `json:"name,omitempty"`
	Inputs []decodedInput `json:"inputs,omitempty"`
}

type decodedInput struct {
	Name  string          `json:"name,omitempty"`
	Type  string          `json:"type,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

type transactionLog struct {
	Address string       `json:"address,omitempty"`
	Data    string       `json:"data,omitempty"`
	Topics  []string     `json:"topics,omitempty"`
	Decoded *decodedCall `json:"decoded,omitempty"`
}

func runTransactions(cmd *cobra.Command, args []string) error {
	client, err := requireSimClient(cmd)
	if err != nil {
		return err
	}

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	if v, _ := cmd.Flags().GetBool("decode"); v {
		params.Set("decode", "true")
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/transactions/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp transactionsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Warn if --decode is used in text mode since the table can't show decoded data.
		if decode, _ := cmd.Flags().GetBool("decode"); decode {
			fmt.Fprintln(cmd.ErrOrStderr(), "Note: --decode data is only visible in JSON output. Use -o json to see decoded fields.")
		}

		// Print errors to stderr.
		printTransactionErrors(cmd, resp.Errors)

		// Print warnings to stderr.
		printWarnings(cmd, resp.Warnings)

		columns := []string{"CHAIN", "HASH", "FROM", "TO", "VALUE", "BLOCK_TIME"}
		rows := make([][]string, len(resp.Transactions))
		for i, tx := range resp.Transactions {
			rows[i] = []string{
				tx.Chain,
				truncateHash(tx.Hash),
				truncateHash(tx.From),
				truncateHash(tx.To),
				tx.Value,
				tx.BlockTime,
			}
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}

// printTransactionErrors writes transaction-level errors to stderr.
func printTransactionErrors(cmd *cobra.Command, errs *transactionErrors) {
	if errs == nil {
		return
	}
	stderr := cmd.ErrOrStderr()
	if errs.ErrorMessage != "" {
		fmt.Fprintf(stderr, "Error: %s\n", errs.ErrorMessage)
	}
	for _, e := range errs.TransactionErrors {
		fmt.Fprintf(stderr, "  chain_id=%d address=%s", e.ChainID, e.Address)
		if e.Description != "" {
			fmt.Fprintf(stderr, " — %s", e.Description)
		}
		fmt.Fprintln(stderr)
	}
	if errs.ErrorMessage != "" || len(errs.TransactionErrors) > 0 {
		fmt.Fprintln(stderr)
	}
}
