package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmBalance_NativeText(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balance", evmTestAddress, "--token", "native", "--chain-ids", "1"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Chain:")
	assert.Contains(t, out, "Symbol:")
	assert.Contains(t, out, "ETH")
	assert.Contains(t, out, "Amount:")
	assert.Contains(t, out, "Price USD:")
	assert.Contains(t, out, "Value USD:")
}

func TestEvmBalance_NativeJSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balance", evmTestAddress, "--token", "native", "--chain-ids", "1", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "wallet_address")
	assert.Contains(t, resp, "balances")

	balances, ok := resp["balances"].([]interface{})
	require.True(t, ok)
	require.Len(t, balances, 1)

	first, ok := balances[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "native", first["address"])
}

func TestEvmBalance_MissingRequiredFlags(t *testing.T) {
	key := simAPIKey(t)

	// Missing --token
	root := newSimTestRoot()
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balance", evmTestAddress, "--chain-ids", "1"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")

	// Missing --chain-ids
	root2 := newSimTestRoot()
	root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "balance", evmTestAddress, "--token", "native"})
	err2 := root2.Execute()
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "chain-ids")
}
