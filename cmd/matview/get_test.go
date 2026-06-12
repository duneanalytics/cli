package matview_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWithSchedule(t *testing.T) {
	mock := &mockClient{
		getFn: func(name string) (*models.GetMaterializedViewResponse, error) {
			assert.Equal(t, "dune.my_team.result_token_summary", name)
			return &models.GetMaterializedViewResponse{
				SQLID:            "dune.my_team.result_token_summary",
				QueryID:          12345,
				IsPrivate:        true,
				TableSizeBytes:   2048,
				LastExecutionIDs: []string{"exec_1", "exec_2"},
				Schedule: &models.MaterializedViewSchedule{
					CronExpression:    "0 */6 * * *",
					Performance:       "medium",
					NextExecutionTime: strptr("2026-06-12T00:00:00Z"),
					ExpiresAt:         strptr("2026-09-11T23:59:59Z"),
				},
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "get", "dune.my_team.result_token_summary"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Name:        dune.my_team.result_token_summary")
	assert.Contains(t, out, "Query ID:    12345")
	assert.Contains(t, out, "Private:     true")
	assert.Contains(t, out, "Table size:  2.0 KB")
	assert.Contains(t, out, "Executions:  exec_1, exec_2")
	assert.Contains(t, out, "Refresh schedule:")
	assert.Contains(t, out, "Cron:        0 */6 * * *")
	assert.Contains(t, out, "Performance: medium")
	assert.Contains(t, out, "Next run:    2026-06-12T00:00:00Z")
	assert.Contains(t, out, "Expires at:  2026-09-11T23:59:59Z")
}

func TestGetWithoutSchedule(t *testing.T) {
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) {
			return &models.GetMaterializedViewResponse{
				SQLID:   "dune.my_team.result_x",
				QueryID: 7,
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "get", "dune.my_team.result_x"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Refresh schedule: none")
	assert.NotContains(t, out, "Cron:")
}

func TestGetJSONOutput(t *testing.T) {
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) {
			return &models.GetMaterializedViewResponse{SQLID: "dune.t.result_x", QueryID: 7, IsPrivate: true}, nil
		},
	}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "get", "dune.t.result_x", "-o", "json"})
	require.NoError(t, root.Execute())

	var resp models.GetMaterializedViewResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Equal(t, "dune.t.result_x", resp.SQLID)
	assert.Equal(t, 7, resp.QueryID)
}

func TestGetMissingArgument(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"matview", "get"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestGetAPIError(t *testing.T) {
	mock := &mockClient{
		getFn: func(string) (*models.GetMaterializedViewResponse, error) {
			return nil, errors.New("api: not found")
		},
	}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"matview", "get", "dune.t.result_missing"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
