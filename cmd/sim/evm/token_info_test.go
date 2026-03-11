package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmTokenInfo_Native_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "native", "--chain-ids", "1"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Chain:")
	assert.Contains(t, out, "Symbol:")
	assert.Contains(t, out, "Price USD:")
}

func TestEvmTokenInfo_Native_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "native", "--chain-ids", "1", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "contract_address")
	assert.Contains(t, resp, "tokens")

	tokens, ok := resp["tokens"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, tokens)

	token, ok := tokens[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, token, "chain")
	assert.Contains(t, token, "symbol")
	assert.Contains(t, token, "price_usd")
}

func TestEvmTokenInfo_ERC20(t *testing.T) {
	key := simAPIKey(t)

	// USDC on Base
	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913", "--chain-ids", "8453", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "tokens")

	tokens, ok := resp["tokens"].([]interface{})
	require.True(t, ok)
	if len(tokens) > 0 {
		token, ok := tokens[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, token, "symbol")
		assert.Contains(t, token, "decimals")
	}
}

func TestEvmTokenInfo_HistoricalPrices(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "native", "--chain-ids", "1", "--historical-prices", "168,24", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	tokens, ok := resp["tokens"].([]interface{})
	require.True(t, ok)
	require.NotEmpty(t, tokens)

	token, ok := tokens[0].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, token, "historical_prices", "historical_prices should be present when --historical-prices is set")
}

func TestEvmTokenInfo_HistoricalPrices_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "native", "--chain-ids", "1", "--historical-prices", "168"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Price 168h ago:")
}

func TestEvmTokenInfo_RequiresChainIds(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "token-info", "native"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain-ids")
}
