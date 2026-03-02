package query_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	createQueryFn func(models.CreateQueryRequest) (*models.CreateQueryResponse, error)
}

func (m *mockClient) CreateQuery(req models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
	return m.createQueryFn(req)
}

// newTestRoot builds a root → query → create command tree with the mock injected.
func newTestRoot(mock dune.DuneClient) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())

	root.AddCommand(query.NewQueryCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}

func TestCreateSuccess(t *testing.T) {
	mock := &mockClient{
		createQueryFn: func(req models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
			assert.Equal(t, "Test", req.Name)
			assert.Equal(t, "SELECT 1", req.QuerySQL)
			return &models.CreateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "create", "--name", "Test", "--sql", "SELECT 1"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "Created query 4125432\n", buf.String())
}

func TestCreateJSONOutput(t *testing.T) {
	mock := &mockClient{
		createQueryFn: func(_ models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
			return &models.CreateQueryResponse{QueryID: 4125432}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "create", "--name", "Test", "--sql", "SELECT 1", "-o", "json"})
	require.NoError(t, root.Execute())

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 4125432, got["query_id"])
}

func TestCreatePrivateFlag(t *testing.T) {
	mock := &mockClient{
		createQueryFn: func(req models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
			assert.True(t, req.IsPrivate)
			return &models.CreateQueryResponse{QueryID: 1}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "S", "--private"})
	require.NoError(t, root.Execute())
}

func TestCreateDescriptionFlag(t *testing.T) {
	mock := &mockClient{
		createQueryFn: func(req models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
			assert.Equal(t, "my desc", req.Description)
			return &models.CreateQueryResponse{QueryID: 1}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "S", "--description", "my desc"})
	require.NoError(t, root.Execute())
}

func TestCreateMissingName(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "create", "--sql", "SELECT 1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "name" not set`)
}

func TestCreateMissingSQL(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "create", "--name", "Test"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "sql" not set`)
}

func TestCreateAPIError(t *testing.T) {
	mock := &mockClient{
		createQueryFn: func(_ models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
			return nil, errors.New("api: unauthorized")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "S"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: unauthorized")
}
