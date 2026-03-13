package svm_test

import (
	"context"
	"os"
	"testing"

	"github.com/duneanalytics/cli/cmd/sim"
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

const svmTestAddress = "86xCnPeV69n6t3DnyGvkKobf9FdN2H9oiVDdaMpo2MMY"

// newSimTestRoot builds the full command tree: dune -> sim -> svm -> <subcommands>.
// Used for authenticated E2E tests. Pass the API key via --sim-api-key in SetArgs.
func newSimTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())
	root.AddCommand(sim.NewSimCmd())
	return root
}
