package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmStablecoins_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "stablecoins", evmTestAddress, "--chain-ids", "1"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "SYMBOL")
	assert.Contains(t, out, "VALUE_USD")
}

func TestEvmStablecoins_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "stablecoins", evmTestAddress, "--chain-ids", "1", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "wallet_address")
	assert.Contains(t, resp, "balances")
}
