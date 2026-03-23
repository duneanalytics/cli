package evm

import (
	"github.com/spf13/cobra"
)

// NewStablecoinsCmd returns the `sim evm stablecoins` command.
func NewStablecoinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stablecoins <address>",
		Short: "Get stablecoin balances for a wallet address across multiple chains",
		Long: "Return only stablecoin balances (USDC, USDT, DAI, FRAX, etc.) for the given\n" +
			"wallet address across supported EVM chains. This is a convenience shorthand\n" +
			"for 'dune sim evm balances <address> --asset-class stablecoin'.\n\n" +
			"Each balance entry includes the token amount, current USD price, and total\n" +
			"USD value. The same response format and pagination as 'dune sim evm balances'\n" +
			"applies.\n\n" +
			"By default, queries all chains tagged 'default'. Use --chain-ids to restrict\n" +
			"to specific networks. Run 'dune sim evm supported-chains' for valid IDs.\n\n" +
			"For all token balances (not just stablecoins), use 'dune sim evm balances'.\n\n" +
			"Examples:\n" +
			"  dune sim evm stablecoins 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm stablecoins 0xd8da... --chain-ids 1,8453\n" +
			"  dune sim evm stablecoins 0xd8da... --exclude-spam -o json",
		Args: cobra.ExactArgs(1),
		RunE: runStablecoins,
	}

	addBalanceFlags(cmd)

	return cmd
}

func runStablecoins(cmd *cobra.Command, args []string) error {
	return runBalancesEndpoint(cmd, args, "/v1/evm/balances/", "/stablecoins")
}
