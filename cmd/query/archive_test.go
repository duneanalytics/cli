package query_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiveSuccess(t *testing.T) {
	mock := &mockClient{
		archiveQueryFn: func(id int) (*models.UpdateQueryResponse, error) {
			assert.Equal(t, 4125432, id)
			return &models.UpdateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "archive", "4125432"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "Archived query 4125432\n", buf.String())
}

func TestArchiveJSONOutput(t *testing.T) {
	mock := &mockClient{
		archiveQueryFn: func(_ int) (*models.UpdateQueryResponse, error) {
			return &models.UpdateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "archive", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 4125432, got["query_id"])
}

func TestArchiveMissingArgument(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "archive"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestArchiveNonIntegerID(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "archive", "abc"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestArchiveAPIError(t *testing.T) {
	mock := &mockClient{
		archiveQueryFn: func(_ int) (*models.UpdateQueryResponse, error) {
			return nil, errors.New("api: not found")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "archive", "999"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: not found")
}
