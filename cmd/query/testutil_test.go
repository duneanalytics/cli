package query_test

import (
	"bytes"
	"context"
	"time"

	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// mockExecution implements dune.Execution for testing.
type mockExecution struct {
	dune.Execution
	id               string
	waitGetResultsFn func(time.Duration, int) (*models.ResultsResponse, error)
}

func (m *mockExecution) WaitGetResults(poll time.Duration, maxRetries int) (*models.ResultsResponse, error) {
	return m.waitGetResultsFn(poll, maxRetries)
}

func (m *mockExecution) GetID() string { return m.id }

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	createQueryFn  func(models.CreateQueryRequest) (*models.CreateQueryResponse, error)
	getQueryFn     func(int) (*models.GetQueryResponse, error)
	updateQueryFn  func(int, models.UpdateQueryRequest) (*models.UpdateQueryResponse, error)
	archiveQueryFn func(int) (*models.UpdateQueryResponse, error)
	runQueryFn     func(models.ExecuteRequest) (dune.Execution, error)
	queryExecuteFn func(models.ExecuteRequest) (*models.ExecuteResponse, error)
	runSQLFn       func(models.ExecuteSQLRequest) (dune.Execution, error)
	sqlExecuteFn   func(models.ExecuteSQLRequest) (*models.ExecuteResponse, error)
}

func (m *mockClient) CreateQuery(req models.CreateQueryRequest) (*models.CreateQueryResponse, error) {
	return m.createQueryFn(req)
}

func (m *mockClient) GetQuery(queryID int) (*models.GetQueryResponse, error) {
	return m.getQueryFn(queryID)
}

func (m *mockClient) UpdateQuery(queryID int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error) {
	return m.updateQueryFn(queryID, req)
}

func (m *mockClient) ArchiveQuery(queryID int) (*models.UpdateQueryResponse, error) {
	return m.archiveQueryFn(queryID)
}

func (m *mockClient) RunQuery(req models.ExecuteRequest) (dune.Execution, error) {
	return m.runQueryFn(req)
}

func (m *mockClient) QueryExecute(req models.ExecuteRequest) (*models.ExecuteResponse, error) {
	return m.queryExecuteFn(req)
}

func (m *mockClient) RunSQL(req models.ExecuteSQLRequest) (dune.Execution, error) {
	return m.runSQLFn(req)
}

func (m *mockClient) SQLExecute(req models.ExecuteSQLRequest) (*models.ExecuteResponse, error) {
	return m.sqlExecuteFn(req)
}

// newTestRoot builds a root → query command tree with the mock injected.
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
