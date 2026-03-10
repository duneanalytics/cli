package svm

import (
	"github.com/spf13/cobra"
)

// NewSvmCmd returns the `sim svm` parent command.
func NewSvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svm",
		Short: "Query SVM chain data (balances, transactions)",
		Long: "Access real-time SVM blockchain data including token balances and\n" +
			"transaction history for Solana and Eclipse chains.",
	}

	// Subcommands will be added here as they are implemented.

	return cmd
}
