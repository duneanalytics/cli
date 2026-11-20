package query_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testQueryResponse = &models.GetQueryResponse{
	QueryID:     4125432,
	Name:        "Test Query",
	Description: "A test",
	QuerySQL:    "SELECT 1",
	Owner:       "user123",
	IsPrivate:   false,
	IsArchived:  false,
	IsUnsaved:   false,
	Version:     3,
	QueryEngine: "v2 DuneSQL",
	Tags:        []string{"defi", "test"},
	Parameters:  nil,
}

func TestGetSuccess(t *testing.T) {
	mock := &mockClient{
		getQueryFn: func(id int) (*models.GetQueryResponse, error) {
			assert.Equal(t, 4125432, id)
			return testQueryResponse, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "get", "4125432"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "ID:          4125432")
	assert.Contains(t, out, "Name:        Test Query")
	assert.Contains(t, out, "Description: A test")
	assert.Contains(t, out, "Owner:       user123")
	assert.Contains(t, out, "Engine:      v2 DuneSQL")
	assert.Contains(t, out, "Version:     3")
	assert.Contains(t, out, "Private:     false")
	assert.Contains(t, out, "Archived:    false")
	assert.Contains(t, out, "Tags:        defi, test")
	assert.Contains(t, out, "SQL:")
	assert.Contains(t, out, "  SELECT 1")
}

func TestGetJSONOutput(t *testing.T) {
	mock := &mockClient{
		getQueryFn: func(_ int) (*models.GetQueryResponse, error) {
			return testQueryResponse, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "get", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.GetQueryResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 4125432, got.QueryID)
	assert.Equal(t, "Test Query", got.Name)
	assert.Equal(t, "SELECT 1", got.QuerySQL)
}

func TestGetMissingArgument(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "get"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestGetNonIntegerID(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "get", "abc"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestGetAPIError(t *testing.T) {
	mock := &mockClient{
		getQueryFn: func(_ int) (*models.GetQueryResponse, error) {
			return nil, errors.New("api: not found")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "get", "999"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: not found")
}

func TestGetEmptyDescriptionAndTags(t *testing.T) {
	mock := &mockClient{
		getQueryFn: func(_ int) (*models.GetQueryResponse, error) {
			return &models.GetQueryResponse{
				QueryID:     1,
				Name:        "Minimal",
				QuerySQL:    "SELECT 1",
				Owner:       "user",
				QueryEngine: "v2 DuneSQL",
				Version:     1,
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "get", "1"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.NotContains(t, out, "Description:")
	assert.NotContains(t, out, "Tags:")
	assert.Contains(t, out, "ID:          1")
	assert.Contains(t, out, "Name:        Minimal")
}

func TestGetShowsTheLatestExecution(t *testing.T) {
	mock := &mockClient{getQueryFn: func(int) (*models.GetQueryResponse, error) { return testQueryResponse, nil }}
	fake := &dunemcptest.Fake{Results: map[string]string{"getDuneQuery": `{"query_id":4125432,
		"latest_execution_id":"01EXEC","latest_execution_state":"QUERY_STATE_COMPLETED"}`}}

	root, buf := newTestRootWithMCP(mock, fake)
	root.SetArgs([]string{"query", "get", "4125432"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"query_id":4125432}`, fake.Last().Args)
	assert.Contains(t, buf.String(), "Last run:    01EXEC (COMPLETED)")

	// JSON keeps every REST field and adds the latest execution.
	root, buf = newTestRootWithMCP(mock, fake)
	root.SetArgs([]string{"query", "get", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "SELECT 1", got["query_sql"])
	assert.Equal(t, "01EXEC", got["latest_execution_id"])
	assert.Equal(t, "QUERY_STATE_COMPLETED", got["latest_execution_state"])
}

func TestGetNeverRun(t *testing.T) {
	mock := &mockClient{getQueryFn: func(int) (*models.GetQueryResponse, error) { return testQueryResponse, nil }}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "get", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Contains(t, got, "latest_execution_id")
	assert.Nil(t, got["latest_execution_id"])
}

func TestGetPrintsTheQueryWhenTheLatestExecutionIsUnavailable(t *testing.T) {
	mock := &mockClient{getQueryFn: func(int) (*models.GetQueryResponse, error) { return testQueryResponse, nil }}
	fake := &dunemcptest.Fake{Errors: map[string]error{"getDuneQuery": errors.New("mcp unavailable")}}
	root, buf := newTestRootWithMCP(mock, fake)
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetArgs([]string{"query", "get", "4125432"})

	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "SELECT 1")
	assert.Contains(t, buf.String(), "Last run:    unknown")
	assert.Equal(t, "warning: could not read the latest execution: mcp unavailable\n", stderr.String())

	// In JSON, null means "never run", so an unknown latest execution leaves the keys out.
	root, buf = newTestRootWithMCP(mock, fake)
	root.SetErr(&stderr)
	root.SetArgs([]string{"query", "get", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())
	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "SELECT 1", got["query_sql"])
	assert.NotContains(t, got, "latest_execution_id")
	assert.NotContains(t, got, "latest_execution_state")
}
