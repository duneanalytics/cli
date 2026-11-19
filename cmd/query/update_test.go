package query_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runUpdate runs `dune query update` against a recording SDK mock and an MCP fake.
func runUpdate(t *testing.T, args ...string) (*mockClient, *dunemcptest.Fake, string, error) {
	t.Helper()
	calls := []string{}
	mock := &mockClient{calls: &calls}
	mock.archiveQueryFn = func(id int) (*models.UpdateQueryResponse, error) { return mock.record("ArchiveQuery", id) }
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDuneQuery": `{"queryId":7}`}}
	root, buf := newTestRootWithMCP(mock, fake)
	root.SetArgs(append([]string{"query", "update"}, args...))
	err := root.Execute()
	return mock, fake, buf.String(), err
}

func TestUpdatePublicDoesNotPatch(t *testing.T) {
	mock, fake, out, err := runUpdate(t, "7", "--private=false")
	require.NoError(t, err)
	assert.Equal(t, "Updated query 7\n", out)
	assert.Equal(t, []string{"MakeQueryPublic"}, *mock.calls)
	assert.Empty(t, mock.updateReqs)
	assert.Empty(t, fake.Calls)
}

func TestUpdatePrivateAndName(t *testing.T) {
	mock, fake, _, err := runUpdate(t, "7", "--private", "--name", "New")
	require.NoError(t, err)
	assert.Equal(t, []string{"UpdateQuery", "MakeQueryPrivate"}, *mock.calls)
	require.Len(t, mock.updateReqs, 1)
	req := mock.updateReqs[0]
	require.NotNil(t, req.Name)
	assert.Equal(t, "New", *req.Name)
	assert.Nil(t, req.QuerySQL)
	assert.Nil(t, req.Description)
	assert.Nil(t, req.IsPrivate)
	assert.Nil(t, req.IsArchived)
	assert.Empty(t, fake.Calls)
}

func TestUpdateArchived(t *testing.T) {
	mock, _, _, err := runUpdate(t, "7", "--archived", "--sql", "SELECT 2")
	require.NoError(t, err)
	// Edit first, then archive.
	assert.Equal(t, []string{"UpdateQuery", "ArchiveQuery"}, *mock.calls)
	assert.Nil(t, mock.updateReqs[0].IsArchived)

	mock, _, _, err = runUpdate(t, "7", "--archived=false", "--sql", "SELECT 2")
	require.NoError(t, err)
	// Restore first, then edit.
	assert.Equal(t, []string{"UnarchiveQuery", "UpdateQuery"}, *mock.calls)
}

func TestUpdateTempGoesToMCPAlone(t *testing.T) {
	mock, fake, out, err := runUpdate(t, "7", "--temp=false")
	require.NoError(t, err)
	assert.Equal(t, "Updated query 7\n", out)
	assert.Empty(t, *mock.calls)
	require.Len(t, fake.Calls, 1)
	assert.Equal(t, "updateDuneQuery", fake.Last().Tool)
	assert.JSONEq(t, `{"queryId":7,"is_temp":false}`, fake.Last().Args)

	// Other flags never reach the MCP tool.
	mock, fake, _, err = runUpdate(t, "7", "--temp", "--name", "N", "--private")
	require.NoError(t, err)
	assert.Equal(t, []string{"UpdateQuery", "MakeQueryPrivate"}, *mock.calls)
	assert.JSONEq(t, `{"queryId":7,"is_temp":true}`, fake.Last().Args)
}

func TestUpdateEmptyDescriptionClearsIt(t *testing.T) {
	mock, _, _, err := runUpdate(t, "7", "--description", "", "--tags", "defi,uniswap")
	require.NoError(t, err)
	require.Len(t, mock.updateReqs, 1)
	req := mock.updateReqs[0]
	require.NotNil(t, req.Description)
	assert.Equal(t, "", *req.Description)
	assert.Equal(t, []string{"defi", "uniswap"}, req.Tags)
	assert.Nil(t, req.Name)
}

func TestUpdateParameters(t *testing.T) {
	params := filepath.Join(t.TempDir(), "params.json")
	require.NoError(t, os.WriteFile(params, []byte(`[{"key":"n","type":"number","value":"10"}]`), 0o600))
	mock, _, _, err := runUpdate(t, "7", "--parameters", "@"+params)
	require.NoError(t, err)
	require.Len(t, mock.updateReqs, 1)
	assert.Equal(t, []models.QueryParameter{{Key: "n", Type: "number", Value: "10"}}, mock.updateReqs[0].Parameters)
}

// An empty --tags or --parameters must reach the API as an empty list, which clears them.
func TestUpdateClearsTagsAndParameters(t *testing.T) {
	mock, _, _, err := runUpdate(t, "7", "--tags", "", "--parameters", "[]")
	require.NoError(t, err)
	require.Len(t, mock.updateReqs, 1)
	body, err := json.Marshal(mock.updateReqs[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{"tags":[],"parameters":[]}`, string(body))
}

func TestUpdateNoFlags(t *testing.T) {
	mock, fake, _, err := runUpdate(t, "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
	assert.Empty(t, *mock.calls)
	assert.Empty(t, fake.Calls)
}

func TestUpdateNonIntegerID(t *testing.T) {
	_, _, _, err := runUpdate(t, "abc", "--name", "X")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestUpdateJSONOutput(t *testing.T) {
	_, _, out, err := runUpdate(t, "7", "--name", "New", "-o", "json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_id":7}`, out)
}
