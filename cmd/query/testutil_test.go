package query_test

import (
	"bytes"
	"context"

	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	createQueryFn func(models.CreateQueryRequest) (*models.CreateQueryResponse, error)
	getQueryFn    func(int) (*models.GetQueryResponse, error)
	updateQueryFn func(int, models.UpdateQueryRequest) (*models.UpdateQueryResponse, error)
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
