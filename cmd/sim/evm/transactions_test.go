package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmTransactions_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "HASH")
	assert.Contains(t, out, "FROM")
	assert.Contains(t, out, "TO")
	assert.Contains(t, out, "VALUE")
	assert.Contains(t, out, "BLOCK_TIME")
}

func TestEvmTransactions_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "transactions")
}

func TestEvmTransactions_DecodeJSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--decode", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "transactions")

	// When decode is enabled, transactions may contain decoded and logs fields.
	// We just verify the response is valid JSON with transactions.
	txs, ok := resp["transactions"].([]interface{})
	require.True(t, ok)
	if len(txs) > 0 {
		tx, ok := txs[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, tx, "hash")
		assert.Contains(t, tx, "chain")
	}
}

// TestEvmTransactions_DecodeText exercises --decode in text mode (the default).
// This is the code path that unmarshals decoded inputs into Go structs. If
// Value were typed as string rather than json.RawMessage, non-string ABI
// arguments (numbers, booleans, arrays) would cause an unmarshal error here.
func TestEvmTransactions_DecodeText(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--decode", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "HASH")

	// Text mode should print the stderr hint about --decode being JSON-only.
	assert.Contains(t, errBuf.String(), "--decode data is only visible in JSON output")
}

func TestEvmTransactions_Pagination(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "transactions")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "transactions", evmTestAddress, "--chain-ids", "1", "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "transactions")
	}
}
