package evm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/duneanalytics/cli/cmd/sim"
	"github.com/duneanalytics/cli/cmd/sim/evm"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newEvmTestRoot builds a minimal command tree: dune -> evm -> <subcommands>.
// A bare (unauthenticated) SimClient is stored in context so public-endpoint
// commands can use the shared HTTP infrastructure.
func newEvmTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())

	evmCmd := evm.NewEvmCmd()
	root.AddCommand(evmCmd)

	// Inject a bare client like simPreRun does for skipSimAuth commands.
	cmdutil.SetSimClient(root, sim.NewBareSimClient())

	return root
}

// supported-chains is a public endpoint — no API key required.

func TestSupportedChains_Text(t *testing.T) {
	root := newEvmTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"evm", "supported-chains"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "CHAIN_ID")
	assert.Contains(t, out, "BALANCES")
	assert.Contains(t, out, "ethereum")
}

func TestSupportedChains_JSON(t *testing.T) {
	root := newEvmTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"evm", "supported-chains", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "chains")

	chains, ok := resp["chains"].([]interface{})
	require.True(t, ok, "chains should be an array")
	require.NotEmpty(t, chains, "should have at least one chain")

	first, ok := chains[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, first, "name")
	assert.Contains(t, first, "chain_id")
}
