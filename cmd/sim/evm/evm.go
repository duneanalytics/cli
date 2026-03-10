package evm

import (
	"github.com/spf13/cobra"
)

// NewEvmCmd returns the `sim evm` parent command.
func NewEvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "Query EVM chain data (balances, activity, transactions, etc.)",
		Long: "Access real-time EVM blockchain data including token balances, activity feeds,\n" +
			"transaction history, NFT collectibles, token metadata, token holders,\n" +
			"and DeFi positions.",
	}

	// Subcommands will be added here as they are implemented.

	return cmd
}
