package evm

import (
	"context"
	"net/url"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/spf13/cobra"
)

// SimClient is the interface that evm commands use to talk to the Sim API.
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

// NewEvmCmd returns the `sim evm` parent command.
func NewEvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "Query EVM chain data (balances, activity, transactions, etc.)",
		Long: "Access real-time EVM blockchain data including token balances, activity feeds,\n" +
			"transaction history, NFT collectibles, token metadata, token holders,\n" +
			"and DeFi positions.",
	}

	cmd.AddCommand(NewSupportedChainsCmd())
	cmd.AddCommand(NewBalancesCmd())
	cmd.AddCommand(NewBalanceCmd())
	cmd.AddCommand(NewStablecoinsCmd())
	cmd.AddCommand(NewActivityCmd())
	cmd.AddCommand(NewTransactionsCmd())
	cmd.AddCommand(NewCollectiblesCmd())
	cmd.AddCommand(NewTokenInfoCmd())
	cmd.AddCommand(NewTokenHoldersCmd())
	cmd.AddCommand(NewDefiPositionsCmd())

	return cmd
}
