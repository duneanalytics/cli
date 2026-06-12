package matview_test

import (
	"bytes"
	"context"

	"github.com/duneanalytics/cli/cmd/matview"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	upsertFn  func(models.UpsertMaterializedViewRequest) (*models.UpsertMaterializedViewResponse, error)
	getFn     func(string) (*models.GetMaterializedViewResponse, error)
	listFn    func(int, int) (*models.ListMaterializedViewsResponse, error)
	refreshFn func(string, models.RefreshMaterializedViewRequest) (*models.RefreshMaterializedViewResponse, error)
	deleteFn  func(string) (*models.DeleteMaterializedViewResponse, error)
}

func (m *mockClient) UpsertMaterializedView(
	req models.UpsertMaterializedViewRequest,
) (*models.UpsertMaterializedViewResponse, error) {
	return m.upsertFn(req)
}

func (m *mockClient) GetMaterializedView(name string) (*models.GetMaterializedViewResponse, error) {
	return m.getFn(name)
}

func (m *mockClient) ListMaterializedViews(limit, offset int) (*models.ListMaterializedViewsResponse, error) {
	return m.listFn(limit, offset)
}

func (m *mockClient) RefreshMaterializedView(
	name string,
	req models.RefreshMaterializedViewRequest,
) (*models.RefreshMaterializedViewResponse, error) {
	return m.refreshFn(name, req)
}

func (m *mockClient) DeleteMaterializedView(name string) (*models.DeleteMaterializedViewResponse, error) {
	return m.deleteFn(name)
}

// newTestRoot builds a root → matview command tree with the mock injected.
func newTestRoot(mock dune.DuneClient) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(matview.NewMatviewCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}

func strptr(s string) *string { return &s }
