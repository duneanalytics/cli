package svm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSvmTransactions_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "BLOCK_SLOT")
	assert.Contains(t, out, "BLOCK_TIME")
	assert.Contains(t, out, "TX_SIGNATURE")
}

func TestSvmTransactions_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "transactions")

	txns, ok := resp["transactions"].([]interface{})
	require.True(t, ok)
	if len(txns) > 0 {
		tx, ok := txns[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, tx, "address")
		assert.Contains(t, tx, "block_slot")
		assert.Contains(t, tx, "block_time")
		assert.Contains(t, tx, "chain")
		assert.Contains(t, tx, "raw_transaction")
	}
}

func TestSvmTransactions_Limit(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "3", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	txns, ok := resp["transactions"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(txns), 3)
}

func TestSvmTransactions_Pagination(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "transactions")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "transactions")
	}
}

func TestSvmTransactions_RawTransactionInJSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "svm", "transactions", svmTestAddress, "--limit", "1", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	txns, ok := resp["transactions"].([]interface{})
	require.True(t, ok)
	if len(txns) > 0 {
		tx, ok := txns[0].(map[string]interface{})
		require.True(t, ok)

		// raw_transaction should be a nested object with transaction data.
		rawTx, ok := tx["raw_transaction"].(map[string]interface{})
		if ok {
			// Should contain transaction with signatures.
			txData, ok := rawTx["transaction"].(map[string]interface{})
			if ok {
				assert.Contains(t, txData, "signatures")
			}
		}
	}
}
