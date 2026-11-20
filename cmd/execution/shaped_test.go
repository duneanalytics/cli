package execution_test

import (
	"testing"
	"time"

	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var shapedResponse = &models.ResultsResponse{
	State:               "QUERY_STATE_COMPLETED",
	ExecutionEndedAt:    ptrTime(time.Now()),
	IsExecutionFinished: true,
	Result: models.Result{
		Metadata: models.ResultMetadata{ColumnNames: []string{"amount", "token"}, RowCount: 2, TotalRowCount: 10},
		Rows:     []map[string]any{{"token": "ETH", "amount": float64(9)}, {"token": "USDC", "amount": float64(4)}},
	},
}

// runShaped runs `dune execution results` and returns the options of every results call.
func runShaped(t *testing.T, mock *mockClient, args ...string) ([]models.ResultOptions, string, error) {
	t.Helper()
	var got []models.ResultOptions
	mock.queryResultsV2Fn = func(id string, opts models.ResultOptions) (*models.ResultsResponse, error) {
		assert.Equal(t, "01EXEC", id)
		got = append(got, opts)
		return shapedResponse, nil
	}
	root, buf := newTestRoot(mock)
	root.SetArgs(append([]string{"execution", "results", "01EXEC"}, args...))
	err := root.Execute()
	return got, buf.String(), err
}

func TestShapedResultsOptions(t *testing.T) {
	got, out, err := runShaped(t, &mockClient{}, "--no-wait", "--columns", "amount,token",
		"--filter", `"amount" > 1`, "--sort-by", `"amount" DESC`, "--sort-by", "token, block_time")
	require.NoError(t, err)

	require.Len(t, got, 1)
	assert.Equal(t, models.ResultOptions{
		Columns: []string{"amount", "token"},
		Filters: `"amount" > 1`,
		// Repeated --sort-by values stay separate, commas and all.
		SortBy: []string{`"amount" DESC`, "token, block_time"},
	}, got[0])
	assert.Regexp(t, `amount\s+token\n9\s+ETH\n4\s+USDC\n`, out)
	assert.Contains(t, out, "2 rows")
}

func TestShapedResultsWithPage(t *testing.T) {
	got, _, err := runShaped(t, &mockClient{}, "--no-wait", "--filter", "x", "--limit", "10", "--offset", "5")
	require.NoError(t, err)
	assert.Equal(t, models.ResultOptions{Filters: "x", Page: &models.ResultPageOption{Offset: 5, Limit: 10}}, got[0])
}

func TestSampleSendsNoPagination(t *testing.T) {
	got, _, err := runShaped(t, &mockClient{}, "--no-wait", "--sample", "3")
	require.NoError(t, err)
	assert.Equal(t, models.ResultOptions{SampleCount: 3}, got[0])

	_, _, err = runShaped(t, &mockClient{}, "--no-wait", "--sample", "3", "--offset", "5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")

	_, _, err = runShaped(t, &mockClient{}, "--no-wait", "--sample", "0")
	assert.EqualError(t, err, "sample must be positive, got 0")
}

// Waiting with shaping polls the status rather than downloading every row.
func TestShapedResultsWaitPollsStatus(t *testing.T) {
	execution.PollInterval = 0
	t.Cleanup(func() { execution.PollInterval = 2 * time.Second })

	states := []string{"QUERY_STATE_PENDING", "QUERY_STATE_EXECUTING", "QUERY_STATE_COMPLETED"}
	statusCalls := 0
	mock := &mockClient{queryStatusFn: func(string) (*models.StatusResponse, error) {
		statusCalls++
		return &models.StatusResponse{State: states[statusCalls-1]}, nil
	}}
	got, _, err := runShaped(t, mock, "--sample", "5", "--timeout", "10")
	require.NoError(t, err)
	assert.Equal(t, 3, statusCalls)
	require.Len(t, got, 1)
	assert.Equal(t, models.ResultOptions{SampleCount: 5}, got[0])
}

// A timeout must fail, so a script does not take a still-running execution for a result.
func TestShapedResultsTimeoutFails(t *testing.T) {
	mock := &mockClient{queryStatusFn: func(string) (*models.StatusResponse, error) {
		return &models.StatusResponse{State: "QUERY_STATE_EXECUTING"}, nil
	}}
	got, _, err := runShaped(t, mock, "--filter", `"amount" > 1`, "--timeout", "0")
	require.EqualError(t, err,
		"execution 01EXEC is still EXECUTING after 0s; rerun with a longer --timeout or --no-wait")
	assert.Empty(t, got, "no results are fetched")
}

func TestShapedResultsFailedExecution(t *testing.T) {
	mock := &mockClient{queryResultsV2Fn: func(string, models.ResultOptions) (*models.ResultsResponse, error) {
		return &models.ResultsResponse{State: "QUERY_STATE_FAILED",
			Error: &models.ExecutionError{Message: "Column 'x' cannot be resolved"}}, nil
	}}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"execution", "results", "01EXEC", "--no-wait", "--filter", `"x" > 1`})
	assert.EqualError(t, root.Execute(), "Column 'x' cannot be resolved")
}
