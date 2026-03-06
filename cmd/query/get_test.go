package query_test

import (
	"encoding/json"
	"errors"
	"testing"

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
