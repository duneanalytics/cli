package matview_test

import (
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scheduledMatview is a matview with an existing refresh schedule, used to verify read-modify-write.
func scheduledMatview() *models.GetMaterializedViewResponse {
	return &models.GetMaterializedViewResponse{
		SQLID:     "dune.my_team.result_token_summary",
		QueryID:   12345,
		IsPrivate: true,
		Schedule: &models.MaterializedViewSchedule{
			CronExpression: "0 */6 * * *",
			Performance:    "medium",
			ExpiresAt:      strptr("2026-09-11T23:59:59Z"),
		},
	}
}

// Changing an unrelated setting must preserve the existing schedule.
func TestUpdatePreservesSchedule(t *testing.T) {
	var got models.UpsertMaterializedViewRequest
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) { return scheduledMatview(), nil },
		upsertFn: func(req models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			got = req
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.my_team.result_token_summary", ExecutionID: "e1"}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "update", "dune.my_team.result_token_summary", "--private=false"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "result_token_summary", got.Name, "upsert uses the bare name")
	assert.Equal(t, 12345, got.QueryID)
	assert.False(t, got.IsPrivate, "private override applied")
	require.NotNil(t, got.CronExpression, "existing schedule must be preserved")
	assert.Equal(t, "0 */6 * * *", *got.CronExpression)
	assert.Equal(t, "medium", got.Performance, "scheduled tier preserved")
	require.NotNil(t, got.ExpiresAt)
	assert.Equal(t, "2026-09-11T23:59:59Z", *got.ExpiresAt)
}

func TestUpdateReplacesCron(t *testing.T) {
	var got models.UpsertMaterializedViewRequest
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) { return scheduledMatview(), nil },
		upsertFn: func(req models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			got = req
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.my_team.result_token_summary", ExecutionID: "e1"}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "update", "dune.my_team.result_token_summary", "--cron", "0 0 * * *"})
	require.NoError(t, root.Execute())

	require.NotNil(t, got.CronExpression)
	assert.Equal(t, "0 0 * * *", *got.CronExpression)
}

func TestUpdateRemovesSchedule(t *testing.T) {
	var got models.UpsertMaterializedViewRequest
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) { return scheduledMatview(), nil },
		upsertFn: func(req models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error) {
			got = req
			return &models.UpsertMaterializedViewResponse{SQLID: "dune.my_team.result_token_summary", ExecutionID: "e1"}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "update", "dune.my_team.result_token_summary", "--no-schedule"})
	require.NoError(t, root.Execute())

	assert.Nil(t, got.CronExpression, "schedule must be removed")
	assert.Nil(t, got.ExpiresAt)
}

func TestUpdateCronAndNoScheduleMutuallyExclusive(t *testing.T) {
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) { return scheduledMatview(), nil },
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "update", "dune.t.result_x", "--cron", "0 0 * * *", "--no-schedule"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[cron no-schedule]")
}

func TestUpdateNotFound(t *testing.T) {
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) {
			return nil, errors.New("api: not found")
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "update", "dune.t.result_missing", "--private=false"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
