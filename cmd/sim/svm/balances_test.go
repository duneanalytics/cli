package svm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSvmBalances_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "SYMBOL")
	assert.Contains(t, out, "BALANCE")
	assert.Contains(t, out, "PRICE_USD")
	assert.Contains(t, out, "VALUE_USD")
}

func TestSvmBalances_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress, "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "wallet_address")
	assert.Contains(t, resp, "balances")

	balances, ok := resp["balances"].([]interface{})
	require.True(t, ok)
	if len(balances) > 0 {
		b, ok := balances[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, b, "chain")
		assert.Contains(t, b, "address")
		assert.Contains(t, b, "amount")
	}
}

func TestSvmBalances_WithChains(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress, "--chains", "solana", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "balances")

	// All balances should be on solana chain.
	balances, ok := resp["balances"].([]interface{})
	require.True(t, ok)
	for _, bal := range balances {
		b, ok := bal.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "solana", b["chain"])
	}
}

func TestSvmBalances_Limit(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress, "--limit", "3", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	balances, ok := resp["balances"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(balances), 3)
}

func TestSvmBalances_Pagination(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress, "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "balances")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "balances", svmTestAddress, "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "balances")
	}
}
