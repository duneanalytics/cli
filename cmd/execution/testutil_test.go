package execution_test

import (
	"bytes"
	"context"

	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

type mockClient struct {
	dune.DuneClient
	queryResultsV2Fn func(string, models.ResultOptions) (*models.ResultsResponse, error)
}

func (m *mockClient) QueryResultsV2(executionID string, options models.ResultOptions) (*models.ResultsResponse, error) {
	return m.queryResultsV2Fn(executionID, options)
}

// newTestRoot builds a root → execution command tree with the mock injected.
func newTestRoot(mock dune.DuneClient) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(execution.NewExecutionCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}
