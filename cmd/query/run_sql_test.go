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

func newSQLWaitMock(t *testing.T, resp *models.ResultsResponse, respErr error) *mockClient {
	t.Helper()
	return &mockClient{
		runSQLFn: func(req models.ExecuteSQLRequest) (dune.Execution, error) {
			return &mockExecution{
				id: "01ABC",
				waitGetResultsFn: func(_ time.Duration, _ int) (*models.ResultsResponse, error) {
					return resp, respErr
				},
			}, nil
		},
	}
}

func TestRunSQLSuccess(t *testing.T) {
	mock := newSQLWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1"})
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

func TestRunSQLJSONOutput(t *testing.T) {
	mock := newSQLWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.ResultsResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, int64(4125432), got.QueryID)
	assert.Equal(t, "QUERY_STATE_COMPLETED", got.State)
}

func TestRunSQLWithLimit(t *testing.T) {
	mock := newSQLWaitMock(t, testResultsResponse, nil)

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--limit", "1"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "Showing 1 of 2 rows")
	assert.NotContains(t, out, "200")
}

func TestRunSQLWithParams(t *testing.T) {
	var captured models.ExecuteSQLRequest
	mock := &mockClient{
		runSQLFn: func(req models.ExecuteSQLRequest) (dune.Execution, error) {
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
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--param", "wallet=0xabc", "--param", "days=30"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "0xabc", captured.QueryParameters["wallet"])
	assert.Equal(t, "30", captured.QueryParameters["days"])
}

func TestRunSQLWithPerformance(t *testing.T) {
	var captured models.ExecuteSQLRequest
	mock := &mockClient{
		runSQLFn: func(req models.ExecuteSQLRequest) (dune.Execution, error) {
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
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--performance", "large"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "large", captured.Performance)
}

func TestRunSQLExecutionFailed(t *testing.T) {
	failedResp := &models.ResultsResponse{
		QueryID: 4125432,
		State:   "QUERY_STATE_FAILED",
		Error: &models.ExecutionError{
			Type:    "EXECUTION_ERROR",
			Message: "syntax error at line 1",
		},
	}
	mock := newSQLWaitMock(t, failedResp, nil)

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT BAD"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "QUERY_STATE_FAILED")
	assert.Contains(t, err.Error(), "syntax error at line 1")
}

func TestRunSQLAPIError(t *testing.T) {
	mock := &mockClient{
		runSQLFn: func(_ models.ExecuteSQLRequest) (dune.Execution, error) {
			return nil, errors.New("api: connection refused")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: connection refused")
}

func TestRunSQLNoWait(t *testing.T) {
	mock := &mockClient{
		sqlExecuteFn: func(req models.ExecuteSQLRequest) (*models.ExecuteResponse, error) {
			return &models.ExecuteResponse{
				ExecutionID: "01ABCDEFGHIJKLMNOPQRSTUV",
				State:       "QUERY_STATE_PENDING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--no-wait"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Execution ID: 01ABCDEFGHIJKLMNOPQRSTUV")
	assert.Contains(t, out, "State:        QUERY_STATE_PENDING")
}

func TestRunSQLNoWaitJSON(t *testing.T) {
	mock := &mockClient{
		sqlExecuteFn: func(_ models.ExecuteSQLRequest) (*models.ExecuteResponse, error) {
			return &models.ExecuteResponse{
				ExecutionID: "01ABCDEFGHIJKLMNOPQRSTUV",
				State:       "QUERY_STATE_PENDING",
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--no-wait", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.ExecuteResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, "01ABCDEFGHIJKLMNOPQRSTUV", got.ExecutionID)
}

func TestRunSQLNoWaitAPIError(t *testing.T) {
	mock := &mockClient{
		sqlExecuteFn: func(_ models.ExecuteSQLRequest) (*models.ExecuteResponse, error) {
			return nil, errors.New("api: rate limited")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--no-wait"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: rate limited")
}

func TestRunSQLMissingSQL(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run-sql"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "sql" not set`)
}

func TestRunSQLInvalidPerformance(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--performance", "xlarge"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid performance tier")
}

func TestRunSQLInvalidParam(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"query", "run-sql", "--sql", "SELECT 1", "--param", "noequalssign"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected key=value")
}
