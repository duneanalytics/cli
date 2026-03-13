package evm_test

import (
	"context"
	"os"
	"testing"

	"github.com/duneanalytics/cli/cmd/sim"
	"github.com/duneanalytics/cli/cmd/sim/evm"
	"github.com/spf13/cobra"
)

// simAPIKey returns the DUNE_SIM_API_KEY env var or skips the test.
func simAPIKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("DUNE_SIM_API_KEY")
	if key == "" {
		t.Skip("DUNE_SIM_API_KEY not set, skipping e2e test")
	}
	return key
}

const evmTestAddress = "0xd8da6bf26964af9d7eed9e03e53415d37aa96045"

// newEvmTestRoot builds a minimal command tree: dune -> evm -> <subcommands>.
// No sim parent — used for public endpoints that don't require auth.
func newEvmTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())
	root.AddCommand(evm.NewEvmCmd())
	return root
}

// newSimTestRoot builds the full command tree: dune -> sim -> evm -> <subcommands>.
// Used for authenticated E2E tests. Pass the API key via --sim-api-key in SetArgs.
func newSimTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())
	root.AddCommand(sim.NewSimCmd())
	return root
}
