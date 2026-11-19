package matview_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSuccess(t *testing.T) {
	var got models.UpsertMaterializedViewRequest
	mock := &mockClient{
		upsertFn: func(req models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			got = req
			return &models.UpsertMaterializedViewResponse{
				SQLID:       "dune.my_team.result_token_summary",
				ExecutionID: "01HZ065",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{
		"matview", "create",
		"--name", "result_token_summary",
		"--query-id", "12345",
		"--performance", "medium",
	})
	require.NoError(t, root.Execute())

	assert.Equal(t, "result_token_summary", got.Name)
	assert.Equal(t, 12345, got.QueryID)
	assert.Equal(t, "medium", got.Performance)
	assert.Nil(t, got.CronExpression)

	out := buf.String()
	assert.Contains(t, out, "Created materialized view dune.my_team.result_token_summary")
	assert.Contains(t, out, "Execution ID: 01HZ065")
}

func TestCreateWithCron(t *testing.T) {
	var got models.UpsertMaterializedViewRequest
	mock := &mockClient{
		upsertFn: func(req models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			got = req
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.t.result_x", ExecutionID: "e1"}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{
		"matview", "create",
		"--name", "result_scheduled",
		"--query-id", "7",
		"--cron", "0 */6 * * *",
	})
	require.NoError(t, root.Execute())

	require.NotNil(t, got.CronExpression)
	assert.Equal(t, "0 */6 * * *", *got.CronExpression)
}

// The server gives a matview its source query's privacy and ignores is_private, so --private
// on a public query must fail instead of quietly creating a public matview.
func TestCreatePrivateOnPublicQueryFails(t *testing.T) {
	mock := &mockClient{
		private: false,
		upsertFn: func(models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			t.Fatal("the matview must not be created")
			return nil, nil
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "create", "--name", "result_secret", "--query-id", "7", "--private"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query 7 is public")
	assert.Contains(t, err.Error(), "dune query update 7 --private=true")
}

func TestCreatePrivateOnPrivateQuerySucceeds(t *testing.T) {
	created := false
	mock := &mockClient{
		private: true,
		upsertFn: func(models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			created = true
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.t.result_secret", ExecutionID: "e1"}, nil
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "create", "--name", "result_secret", "--query-id", "7", "--private"})
	require.NoError(t, root.Execute())
	assert.True(t, created)
}

func TestCreateInvalidName(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"matview", "create", "--name", "not_result", "--query-id", "1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid matview name")
}

func TestCreateExpiresAtRequiresCron(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{
		"matview", "create",
		"--name", "result_token_summary",
		"--query-id", "1",
		"--expires-at", "2026-09-11T23:59:59Z",
	})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--expires-at requires --cron")
}

func TestCreateInvalidPerformance(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{
		"matview", "create",
		"--name", "result_token_summary",
		"--query-id", "1",
		"--performance", "huge",
	})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid performance tier")
}

func TestCreateJSONOutput(t *testing.T) {
	mock := &mockClient{
		upsertFn: func(models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.t.result_x", ExecutionID: "e1"}, nil
		},
	}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "create", "--name", "result_x_long", "--query-id", "1", "-o", "json"})
	require.NoError(t, root.Execute())

	var resp models.UpsertMaterializedViewResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Equal(t, "dune.t.result_x", resp.SQLID)
}

func TestCreateAPIError(t *testing.T) {
	mock := &mockClient{
		upsertFn: func(models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			return nil, errors.New("api: plan does not allow private matviews")
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "create", "--name", "result_token_summary", "--query-id", "1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "plan does not allow private matviews")
}
