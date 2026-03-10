package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmActivity_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN_ID")
	assert.Contains(t, out, "TYPE")
	assert.Contains(t, out, "ASSET_TYPE")
	assert.Contains(t, out, "SYMBOL")
	assert.Contains(t, out, "VALUE_USD")
	assert.Contains(t, out, "TX_HASH")
	assert.Contains(t, out, "BLOCK_TIME")
}

func TestEvmActivity_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "activity")
}

func TestEvmActivity_ActivityTypeFilter(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--activity-type", "receive", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp struct {
		Activity []struct {
			Type string `json:"type"`
		} `json:"activity"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	// All returned activities should be of the filtered type.
	for _, a := range resp.Activity {
		assert.Equal(t, "receive", a.Type)
	}
}

func TestEvmActivity_AssetTypeFilter(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--asset-type", "erc20", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp struct {
		Activity []struct {
			AssetType string `json:"asset_type"`
		} `json:"activity"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	for _, a := range resp.Activity {
		assert.Equal(t, "erc20", a.AssetType)
	}
}

func TestEvmActivity_Pagination(t *testing.T) {
	key := simAPIKey(t)

	// Fetch page 1 with a small limit to trigger pagination.
	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "activity")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "activity", evmTestAddress, "--chain-ids", "1", "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "activity")
	}
}
