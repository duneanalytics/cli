package svm

import (
	"context"
	"fmt"
	"net/url"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/spf13/cobra"
)

// SimClient is the interface that svm commands use to talk to the Sim API.
// It is satisfied by *sim.SimClient (stored in the command context by
// the sim parent command's PersistentPreRunE).
type SimClient interface {
	Get(ctx context.Context, path string, params url.Values) ([]byte, error)
}

// SimClientFromCmd extracts the SimClient from the command context.
func SimClientFromCmd(cmd *cobra.Command) SimClient {
	v := cmdutil.SimClientFromCmd(cmd)
	if v == nil {
		return nil
	}
	c, ok := v.(SimClient)
	if !ok {
		return nil
	}
	return c
}

// requireSimClient extracts the SimClient or returns an error.
func requireSimClient(cmd *cobra.Command) (SimClient, error) {
	c := SimClientFromCmd(cmd)
	if c == nil {
		return nil, fmt.Errorf("sim client not initialized")
	}
	return c, nil
}

// NewSvmCmd returns the `sim svm` parent command.
func NewSvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svm",
		Short: "Query SVM chain data (Solana, Eclipse) for balances and transactions",
		Long: "Access real-time, indexed SVM (Solana Virtual Machine) blockchain data.\n" +
			"Commands accept a Solana-style base58 wallet address as the primary argument.\n\n" +
			"Supported chains: Solana, Eclipse.\n\n" +
			"Available subcommands:\n" +
			"  balances     - SPL token balances with USD valuations and liquidity data\n" +
			"  transactions - Raw transaction history with block slot and signature\n\n" +
			"Note: SVM endpoints are currently in beta (served under /beta/svm/*).\n" +
			"All commands require a Sim API key.",
	}

	cmd.AddCommand(NewBalancesCmd())
	cmd.AddCommand(NewTransactionsCmd())

	return cmd
}
