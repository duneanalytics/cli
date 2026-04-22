package query_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testResultsResponse = &models.ResultsResponse{
	QueryID: 4125432,
	State:   "QUERY_STATE_COMPLETED",
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

func newWaitMock(t *testing.T, resp *models.ResultsResponse, respErr error) *mockClient {
	t.Helper()
	return &mockClient{
		runQueryFn: func(req models.ExecuteRequest) (dune.Execution, error) {
			return &mockExecution{
				id: "01ABC",
				waitGetResultsFn: func(_ time.Duration, _ int) (*models.ResultsResponse, error) {
					return resp, respErr
				},
			}, nil
		},
	}
}

func TestRunSuccess(t *testing.T) {
	mock := newWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "block_number")
	assert.Contains(t, out, "tx_hash")
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "0xabc")
	assert.Contains(t, out, "200")
	assert.Contains(t, out, "0xdef")
	assert.Contains(t, out, "2 rows")
}

func TestRunJSONOutput(t *testing.T) {
	mock := newWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.ResultsResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, int64(4125432), got.QueryID)
	assert.Equal(t, "QUERY_STATE_COMPLETED", got.State)
}

func TestRunWithLimit(t *testing.T) {
	mock := newWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--limit", "1"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "Showing 1 of 2 rows")
	assert.NotContains(t, out, "200")
}

func TestRunWithParams(t *testing.T) {
	var captured models.ExecuteRequest
	mock := &mockClient{
		runQueryFn: func(req models.ExecuteRequest) (dune.Execution, error) {
			captured = req
			return &mockExecution{
				id: "01ABC",
				waitGetResultsFn: func(_ time.Duration, _ int) (*models.ResultsResponse, error) {
					return testResultsResponse, nil
				},
			}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--param", "wallet=0xabc", "--param", "days=30"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "0xabc", captured.QueryParameters["wallet"])
	assert.Equal(t, "30", captured.QueryParameters["days"])
}

func TestRunDefaultPerformance(t *testing.T) {
	var captured models.ExecuteRequest
	mock := &mockClient{
		runQueryFn: func(req models.ExecuteRequest) (dune.Execution, error) {
			captured = req
			return &mockExecution{
				id: "01ABC",
				waitGetResultsFn: func(_ time.Duration, _ int) (*models.ResultsResponse, error) {
					return testResultsResponse, nil
				},
			}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "", captured.Performance)
}

func TestRunWithPerformance(t *testing.T) {
	var captured models.ExecuteRequest
	mock := &mockClient{
		runQueryFn: func(req models.ExecuteRequest) (dune.Execution, error) {
			captured = req
			return &mockExecution{
				id: "01ABC",
				waitGetResultsFn: func(_ time.Duration, _ int) (*models.ResultsResponse, error) {
					return testResultsResponse, nil
				},
			}, nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--performance", "large"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "large", captured.Performance)
}

func TestRunExecutionFailed(t *testing.T) {
	failedResp := &models.ResultsResponse{
		QueryID: 4125432,
		State:   "QUERY_STATE_FAILED",
		Error: &models.ExecutionError{
			Type:    "EXECUTION_ERROR",
			Message: "syntax error at line 1",
		},
	}
	mock := newWaitMock(t, failedResp, nil)

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "QUERY_STATE_FAILED")
	assert.Contains(t, err.Error(), "syntax error at line 1")
}

func TestRunAPIError(t *testing.T) {
	mock := &mockClient{
		runQueryFn: func(_ models.ExecuteRequest) (dune.Execution, error) {
			return nil, errors.New("api: connection refused")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: connection refused")
}

func TestRunNoWait(t *testing.T) {
	mock := &mockClient{
		queryExecuteFn: func(req models.ExecuteRequest) (*models.ExecuteResponse, error) {
			return &models.ExecuteResponse{
				ExecutionID: "01ABCDEFGHIJKLMNOPQRSTUV",
				State:       "QUERY_STATE_PENDING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--no-wait"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Execution ID: 01ABCDEFGHIJKLMNOPQRSTUV")
	assert.Contains(t, out, "State:        QUERY_STATE_PENDING")
}

func TestRunNoWaitJSON(t *testing.T) {
	mock := &mockClient{
		queryExecuteFn: func(_ models.ExecuteRequest) (*models.ExecuteResponse, error) {
			return &models.ExecuteResponse{
				ExecutionID: "01ABCDEFGHIJKLMNOPQRSTUV",
				State:       "QUERY_STATE_PENDING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--no-wait", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.ExecuteResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "01ABCDEFGHIJKLMNOPQRSTUV", got.ExecutionID)
}

func TestRunNoWaitAPIError(t *testing.T) {
	mock := &mockClient{
		queryExecuteFn: func(_ models.ExecuteRequest) (*models.ExecuteResponse, error) {
			return nil, errors.New("api: rate limited")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run", "4125432", "--no-wait"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: rate limited")
}

func TestRunInvalidParam(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run", "4125432", "--param", "noequalssign"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected key=value")
}

func TestRunEmptyParamKey(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run", "4125432", "--param", "=value"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key cannot be empty")
}

func TestRunMissingArgument(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestRunNonIntegerID(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run", "abc"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestRunInvalidPerformance(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run", "4125432", "--performance", "xlarge"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid performance tier")
}
