package evm

import (
	"github.com/spf13/cobra"
)

// NewStablecoinsCmd returns the `sim evm stablecoins` command.
func NewStablecoinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stablecoins <address>",
		Short: "Get stablecoin balances for a wallet address",
		Long: "Return stablecoin balances for the given wallet address across supported\n" +
			"EVM chains, including USD valuations.\n\n" +
			"Examples:\n" +
			"  dune sim evm stablecoins 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm stablecoins 0xd8da... --chain-ids 1,8453\n" +
			"  dune sim evm stablecoins 0xd8da... -o json",
		Args: cobra.ExactArgs(1),
		RunE: runStablecoins,
	}

	addBalanceFlags(cmd)

	return cmd
}

func runStablecoins(cmd *cobra.Command, args []string) error {
	return runBalancesEndpoint(cmd, args, "/v1/evm/balances/", "/stablecoins")
}
