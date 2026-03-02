package execution_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testResultsResponse = &models.ResultsResponse{
	QueryID:            4125432,
	State:              "QUERY_STATE_COMPLETED",
	ExecutionEndedAt:   ptrTime(time.Now()),
	IsExecutionFinished: true,
	Result: models.Result{
		Metadata: models.ResultMetadata{
			ColumnNames: []string{"block_number", "tx_hash"},
			RowCount:    2,
		},
		Rows: []map[string]any{
			{"block_number": float64(100), "tx_hash": "0xabc"},
			{"block_number": float64(200), "tx_hash": "0xdef"},
		},
	},
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestResultsSuccess(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(id string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			assert.Equal(t, "01ABCDEFGHIJKLMNOPQRSTUV", id)
			return testResultsResponse, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABCDEFGHIJKLMNOPQRSTUV"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "block_number")
	assert.Contains(t, out, "tx_hash")
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "0xabc")
	assert.Contains(t, out, "2 rows")
}

func TestResultsJSONOutput(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return testResultsResponse, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.ResultsResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, int64(4125432), got.QueryID)
	assert.Equal(t, "QUERY_STATE_COMPLETED", got.State)
}

func TestResultsPending(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return &models.ResultsResponse{
				State: "QUERY_STATE_PENDING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Execution ID: 01ABC")
	assert.Contains(t, out, "State:        QUERY_STATE_PENDING")
}

func TestResultsExecuting(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return &models.ResultsResponse{
				State: "QUERY_STATE_EXECUTING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "State:        QUERY_STATE_EXECUTING")
}

func TestResultsFailed(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return &models.ResultsResponse{
				State: "QUERY_STATE_FAILED",
				Error: &models.ExecutionError{
					Type:    "EXECUTION_ERROR",
					Message: "syntax error at line 1",
				},
			}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error at line 1")
}

func TestResultsCancelled(t *testing.T) {
	now := time.Now()
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return &models.ResultsResponse{
				State:       "QUERY_STATE_CANCELLED",
				CancelledAt: &now,
			}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

func TestResultsWithLimitAndOffset(t *testing.T) {
	var capturedOpts models.ResultOptions
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, opts models.ResultOptions) (*models.ResultsResponse, error) {
			capturedOpts = opts
			return testResultsResponse, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC", "--limit", "10", "--offset", "5"})
	require.NoError(t, root.Execute())

	require.NotNil(t, capturedOpts.Page)
	assert.Equal(t, uint32(10), capturedOpts.Page.Limit)
	assert.Equal(t, uint64(5), capturedOpts.Page.Offset)
}

func TestResultsAPIError(t *testing.T) {
	mock := &mockClient{
		queryResultsV2Fn: func(_ string, _ models.ResultOptions) (*models.ResultsResponse, error) {
			return nil, errors.New("api: connection refused")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01ABC"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: connection refused")
}

func TestResultsMissingArgument(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"execution", "results"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}
