package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmBalances_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "SYMBOL")
	assert.Contains(t, out, "AMOUNT")
	assert.Contains(t, out, "PRICE_USD")
	assert.Contains(t, out, "VALUE_USD")
}

func TestEvmBalances_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "wallet_address")
	assert.Contains(t, resp, "balances")
}

func TestEvmBalances_WithFilters(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--filters", "native", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "ETH")
}

func TestEvmBalances_ExcludeSpam(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--exclude-spam", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
}

func TestEvmBalances_Pagination(t *testing.T) {
	key := simAPIKey(t)

	// Fetch page 1 with a small limit to trigger pagination.
	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "balances")

	// If next_offset is present, pagination is working.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		// Fetch page 2 using the offset.
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balances", evmTestAddress, "--chain-ids", "1", "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "balances")
	}
}
