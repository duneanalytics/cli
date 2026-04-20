package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmSupportedProtocols_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "supported-protocols"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "FAMILY")
	assert.Contains(t, out, "CHAINS")
	assert.Contains(t, out, "SUB_PROTOCOLS")
}

func TestEvmSupportedProtocols_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "supported-protocols", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "protocol_families")

	families, ok := resp["protocol_families"].([]interface{})
	require.True(t, ok, "protocol_families should be an array")
	require.NotEmpty(t, families, "should have at least one protocol family")

	first, ok := families[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, first, "family")
	assert.Contains(t, first, "chains")
	assert.Contains(t, first, "sub_protocols")

	chains, ok := first["chains"].([]interface{})
	require.True(t, ok, "chains should be an array")
	if len(chains) > 0 {
		c, ok := chains[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, c, "chain_id")
		assert.Contains(t, c, "chain_name")
		assert.Contains(t, c, "status")
	}
}

func TestEvmSupportedProtocols_IncludePreviewChainsFlag(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{
		"sim", "--sim-api-key", key, "evm", "supported-protocols",
		"--include-preview-chains", "-o", "json",
	})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "protocol_families")
}
