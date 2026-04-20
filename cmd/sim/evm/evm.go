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
		Short: "Query EVM chain data (balances, activity, transactions, tokens, NFTs, DeFi)",
		Long: "Access real-time, indexed EVM blockchain data. All commands accept an\n" +
			"Ethereum-style address (0x...) as the primary argument and return data\n" +
			"across multiple EVM chains simultaneously.\n\n" +
			"Available subcommands:\n" +
			"  supported-chains - List all supported EVM chains and endpoint availability (public, no auth)\n" +
			"  balances         - Native + ERC20 token balances with USD valuations\n" +
			"  balance          - Single-token balance lookup on one chain\n" +
			"  stablecoins      - Stablecoin-only balances (USDC, USDT, DAI, etc.)\n" +
			"  activity         - Chronological feed of transfers, swaps, mints, burns, approvals\n" +
			"  transactions     - Raw transaction history with optional ABI decoding\n" +
			"  collectibles     - ERC721 and ERC1155 NFT holdings with spam filtering\n" +
			"  token-info       - Token metadata, pricing, supply, and market cap\n" +
			"  token-holders    - Top holders of an ERC20 token ranked by balance\n" +
			"  defi-positions   - DeFi positions across lending, AMM, and vault protocols (beta)\n" +
			"  supported-protocols - DeFi protocol families and chains covered by defi-positions\n\n" +
			"Most commands support --chain-ids to restrict results to specific networks.\n" +
			"Run 'dune sim evm supported-chains' to discover valid chain IDs, tags, and\n" +
			"which endpoints are available per chain.\n\n" +
			"All commands except 'supported-chains' require a Sim API key.",
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
	cmd.AddCommand(NewSupportedProtocolsCmd())

	return cmd
}
