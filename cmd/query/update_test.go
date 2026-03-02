package query_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSingleFlag(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(id int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			assert.Equal(t, 4125432, id)
			require.NotNil(t, req.Name)
			assert.Equal(t, "New", *req.Name)
			assert.Nil(t, req.QuerySQL)
			assert.Nil(t, req.Description)
			assert.Nil(t, req.IsPrivate)
			assert.Nil(t, req.Tags)
			return &models.UpdateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "4125432", "--name", "New"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "Updated query 4125432\n", buf.String())
}

func TestUpdateMultipleFlags(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(_ int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			require.NotNil(t, req.Name)
			assert.Equal(t, "New", *req.Name)
			require.NotNil(t, req.QuerySQL)
			assert.Equal(t, "SELECT 2", *req.QuerySQL)
			assert.Equal(t, []string{"defi", "uniswap"}, req.Tags)
			assert.Nil(t, req.Description)
			assert.Nil(t, req.IsPrivate)
			return &models.UpdateQueryResponse{QueryID: 1}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "1", "--name", "New", "--sql", "SELECT 2", "--tags", "defi,uniswap"})
	require.NoError(t, root.Execute())
}

func TestUpdatePrivateFlag(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(_ int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			require.NotNil(t, req.IsPrivate)
			assert.True(t, *req.IsPrivate)
			return &models.UpdateQueryResponse{QueryID: 1}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "1", "--private"})
	require.NoError(t, root.Execute())
}

func TestUpdatePrivateFalse(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(_ int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			require.NotNil(t, req.IsPrivate)
			assert.False(t, *req.IsPrivate)
			return &models.UpdateQueryResponse{QueryID: 1}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "1", "--private=false"})
	require.NoError(t, root.Execute())
}

func TestUpdateNoFlags(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "update", "1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestUpdateNonIntegerID(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "update", "abc", "--name", "X"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestUpdateAPIError(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(_ int, _ models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			return nil, errors.New("api: unauthorized")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "1", "--name", "X"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: unauthorized")
}

func TestUpdateJSONOutput(t *testing.T) {
	mock := &mockClient{
		updateQueryFn: func(_ int, _ models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
			return &models.UpdateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "update", "4125432", "--name", "New", "-o", "json"})
	require.NoError(t, root.Execute())

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 4125432, got["query_id"])
}
