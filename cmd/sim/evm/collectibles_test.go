package evm_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmCollectibles_Text(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "SYMBOL")
	assert.Contains(t, out, "TOKEN_ID")
	assert.Contains(t, out, "STANDARD")
	assert.Contains(t, out, "BALANCE")
}

func TestEvmCollectibles_JSON(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "entries")
	assert.Contains(t, resp, "address")
}

func TestEvmCollectibles_FilterSpamDisabled(t *testing.T) {
	key := simAPIKey(t)

	// Fetch with spam filtered (default) and without, compare counts.
	rootFiltered := newSimTestRoot()
	var bufFiltered bytes.Buffer
	rootFiltered.SetOut(&bufFiltered)
	rootFiltered.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--limit", "250", "-o", "json"})
	require.NoError(t, rootFiltered.Execute())

	var respFiltered map[string]interface{}
	require.NoError(t, json.Unmarshal(bufFiltered.Bytes(), &respFiltered))
	filteredEntries, ok := respFiltered["entries"].([]interface{})
	require.True(t, ok)

	rootUnfiltered := newSimTestRoot()
	var bufUnfiltered bytes.Buffer
	rootUnfiltered.SetOut(&bufUnfiltered)
	rootUnfiltered.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--filter-spam=false", "--limit", "250", "-o", "json"})
	require.NoError(t, rootUnfiltered.Execute())

	var respUnfiltered map[string]interface{}
	require.NoError(t, json.Unmarshal(bufUnfiltered.Bytes(), &respUnfiltered))
	unfilteredEntries, ok := respUnfiltered["entries"].([]interface{})
	require.True(t, ok)

	// With spam filtering disabled we should get at least as many entries.
	assert.GreaterOrEqual(t, len(unfilteredEntries), len(filteredEntries),
		"disabling spam filter should return >= entries than with filter enabled")
}

func TestEvmCollectibles_SpamScores(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--filter-spam=false", "--show-spam-scores", "--limit", "5", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "entries")

	// When show_spam_scores is enabled, entries should contain spam_score.
	entries, ok := resp["entries"].([]interface{})
	require.True(t, ok)
	if len(entries) > 0 {
		entry, ok := entries[0].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, entry, "spam_score", "spam_score should be present when --show-spam-scores is set")
		assert.Contains(t, entry, "is_spam")
	}
}

// TestEvmCollectibles_SpamScoresText verifies that --show-spam-scores
// adds SPAM and SPAM_SCORE columns in text mode.
func TestEvmCollectibles_SpamScoresText(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--show-spam-scores", "--limit", "5"})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "CHAIN")
	assert.Contains(t, out, "SPAM")
	assert.Contains(t, out, "SPAM_SCORE")
}

func TestEvmCollectibles_Pagination(t *testing.T) {
	key := simAPIKey(t)

	root := newSimTestRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--limit", "2", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Contains(t, resp, "entries")

	// If next_offset is present, fetch page 2.
	if offset, ok := resp["next_offset"].(string); ok && offset != "" {
		root2 := newSimTestRoot()
		var buf2 bytes.Buffer
		root2.SetOut(&buf2)
		root2.SetArgs([]string{"sim", "--sim-api-key", key, "evm", "collectibles", evmTestAddress, "--chain-ids", "1", "--limit", "2", "--offset", offset, "-o", "json"})

		require.NoError(t, root2.Execute())

		var resp2 map[string]interface{}
		require.NoError(t, json.Unmarshal(buf2.Bytes(), &resp2))
		assert.Contains(t, resp2, "entries")
	}
}
