package matview_test

import (
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshSuccess(t *testing.T) {
	var gotName string
	var gotReq models.RefreshMaterializedViewRequest
	mock := &mockClient{
		refreshFn: func(
			name string, req models.RefreshMaterializedViewRequest,
		) (*models.RefreshMaterializedViewResponse, error) {
			gotName, gotReq = name, req
			return &models.RefreshMaterializedViewResponse{
				SQLID:       "dune.my_team.result_token_summary",
				ExecutionID: "01HZ999",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "refresh", "dune.my_team.result_token_summary", "--performance", "large"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "dune.my_team.result_token_summary", gotName)
	assert.Equal(t, "large", gotReq.Performance)
	out := buf.String()
	assert.Contains(t, out, "Refreshing materialized view dune.my_team.result_token_summary")
	assert.Contains(t, out, "Execution ID: 01HZ999")
}

func TestRefreshDefaultPerformance(t *testing.T) {
	var gotReq models.RefreshMaterializedViewRequest
	mock := &mockClient{
		refreshFn: func(
			_ string, req models.RefreshMaterializedViewRequest,
		) (*models.RefreshMaterializedViewResponse, error) {
			gotReq = req
			return &models.RefreshMaterializedViewResponse{SQLID: "dune.t.result_x", ExecutionID: "e1"}, nil
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "refresh", "dune.t.result_x"})
	require.NoError(t, root.Execute())
	assert.Empty(t, gotReq.Performance, "no --performance sends empty (server default)")
}

func TestRefreshInvalidPerformance(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"matview", "refresh", "dune.t.result_x", "--performance", "huge"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid performance tier")
}

func TestRefreshAPIError(t *testing.T) {
	mock := &mockClient{
		refreshFn: func(string, models.RefreshMaterializedViewRequest) (*models.RefreshMaterializedViewResponse, error) {
			return nil, errors.New("api: not enough credits")
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "refresh", "dune.t.result_x"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enough credits")
}
