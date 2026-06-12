package matview_test

import (
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSuccess(t *testing.T) {
	var gotName string
	mock := &mockClient{
		deleteFn: func(name string) (*models.DeleteMaterializedViewResponse, error) {
			gotName = name
			return &models.DeleteMaterializedViewResponse{Message: "ok"}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "delete", "dune.my_team.result_token_summary"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "dune.my_team.result_token_summary", gotName)
	assert.Contains(t, buf.String(), "Deleted materialized view dune.my_team.result_token_summary")
}

func TestDeleteViaAlias(t *testing.T) {
	mock := &mockClient{
		deleteFn: func(string) (*models.DeleteMaterializedViewResponse, error) {
			return &models.DeleteMaterializedViewResponse{Message: "ok"}, nil
		},
	}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"mv", "delete", "dune.t.result_x"})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "Deleted materialized view dune.t.result_x")
}

func TestDeleteAPIError(t *testing.T) {
	mock := &mockClient{
		deleteFn: func(string) (*models.DeleteMaterializedViewResponse, error) {
			return nil, errors.New("api: Materialized view not found")
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "delete", "dune.t.result_missing"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Materialized view not found")
}
