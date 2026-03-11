package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test token: a token on Base with known holders.
const tokenHoldersChainID = "8453"
const tokenHoldersAddress = "0x63706e401c06ac8513145b7687A14804d17f814b"

func TestEvmTokenHolders_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress, "--chain-id", tokenHoldersChainID, "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "WALLET_ADDRESS")
	assert.Contains(t, out, "BALANCE")
	assert.Contains(t, out, "FIRST_ACQUIRED")
	assert.Contains(t, out, "HAS_TRANSFERRED")
}

func TestEvmTokenHolders_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress, "--chain-id", tokenHoldersChainID, "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "token_address")
	assert.Contains(t, resp, "chain_id")
	assert.Contains(t, resp, "holders")

	holders, ok := resp["holders"].([]interface{})
	require.True(t, ok)
	if len(holders) > 0 {
		h, ok := holders[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, h, "wallet_address")
		assert.Contains(t, h, "balance")
	}
}

func TestEvmTokenHolders_Pagination(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress, "--chain-id", tokenHoldersChainID, "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "holders")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress, "--chain-id", tokenHoldersChainID, "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "holders")
	}
}

func TestEvmTokenHolders_InvalidChainID(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress, "--chain-id", "notanumber"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--chain-id must be a numeric value")
}

func TestEvmTokenHolders_RequiresChainID(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-holders", tokenHoldersAddress})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain-id")
}
