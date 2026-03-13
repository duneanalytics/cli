package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmDefiPositions_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress})

	require.NoError(t, root.Execute())

	out := buf.String()
	// Should contain table headers.
	assert.Contains(t, out, "TYPE")
	assert.Contains(t, out, "CHAIN_ID")
	assert.Contains(t, out, "USD_VALUE")
	assert.Contains(t, out, "DETAILS")
}

func TestEvmDefiPositions_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress, "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "positions")

	positions, ok := resp["positions"].([]interface{})
	require.True(t, ok)
	if len(positions) > 0 {
		p, ok := positions[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, p, "type")
		assert.Contains(t, p, "chain_id")
		assert.Contains(t, p, "usd_value")
	}
}

func TestEvmDefiPositions_WithChainIDs(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress, "--chain-ids", "1", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "positions")

	// All positions should be on chain 1.
	positions, ok := resp["positions"].([]interface{})
	require.True(t, ok)
	for _, pos := range positions {
		p, ok := pos.(map[string]interface{})
		require.True(t, ok)
		chainID, ok := p["chain_id"].(float64)
		if ok {
			assert.Equal(t, float64(1), chainID)
		}
	}
}

func TestEvmDefiPositions_Aggregations(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress, "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	// Check aggregations object.
	agg, ok := resp["aggregations"].(map[string]interface{})
	if ok {
		assert.Contains(t, agg, "total_usd_value")
	}
}

func TestEvmDefiPositions_TextAggregationSummary(t *testing.T) {
	key := simAPIKey(t)

	// First check via JSON whether aggregations are present for this address.
	jsonRoot := newSimTestRoot()
	var jsonBuf bytes.Buffer
	jsonRoot.SetOut(&jsonBuf)
	jsonRoot.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress, "-o", "json"})
	require.NoError(t, jsonRoot.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(jsonBuf.Bytes(), &resp))
	if _, ok := resp["aggregations"]; !ok {
		t.Skip("API did not return aggregations for this address, skipping text aggregation test")
	}

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "defi-positions", evmTestAddress})

	require.NoError(t, root.Execute())

	out := buf.String()
	// When aggregations are present, the summary should appear in text output.
	assert.Contains(t, out, "Total USD Value:")
}
