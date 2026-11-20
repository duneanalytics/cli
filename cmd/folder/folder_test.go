package folder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/duneanalytics/cli/cmd/folder"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	listFn        func(models.ListFoldersOptions) (*models.ListFoldersResponse, error)
	getFn         func(string, models.GetFolderOptions) (*models.GetFolderResponse, error)
	createFn      func(models.CreateFolderRequest) (*models.Folder, error)
	updateFn      func(string, models.UpdateFolderRequest) (*models.Folder, error)
	moveFn        func(string, models.MoveFolderRequest) (*models.Folder, error)
	deleteFn      func(string, bool) (*models.DeleteFolderResponse, error)
	moveContentFn func(models.MoveContentRequest) (*models.MoveContentResponse, error)
}

func (m *mockClient) ListFolders(o models.ListFoldersOptions) (*models.ListFoldersResponse, error) {
	return m.listFn(o)
}

func (m *mockClient) GetFolder(id string, o models.GetFolderOptions) (*models.GetFolderResponse, error) {
	return m.getFn(id, o)
}

func (m *mockClient) CreateFolder(r models.CreateFolderRequest) (*models.Folder, error) {
	return m.createFn(r)
}

func (m *mockClient) UpdateFolder(id string, r models.UpdateFolderRequest) (*models.Folder, error) {
	return m.updateFn(id, r)
}

func (m *mockClient) MoveFolder(id string, r models.MoveFolderRequest) (*models.Folder, error) {
	return m.moveFn(id, r)
}

func (m *mockClient) DeleteFolder(id string, recursive bool) (*models.DeleteFolderResponse, error) {
	return m.deleteFn(id, recursive)
}

func (m *mockClient) MoveContent(r models.MoveContentRequest) (*models.MoveContentResponse, error) {
	return m.moveContentFn(r)
}

var rwa = &models.Folder{FolderID: "f1", Name: "RWA", Path: "/Research/RWA", Icon: "folder", Color: "gray"}

func strPtr(s string) *string { return &s }
func intPtr(n int) *int       { return &n }

func run(t *testing.T, mock *mockClient, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(folder.NewFolderCmd())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"folder"}, args...))
	return &buf, root.Execute()
}

func apiError(status int, code string, details any) error {
	raw, _ := json.Marshal(details)
	return &dune.APIError{StatusCode: status, Message: "folder error", Code: code, Details: raw}
}

func TestListSendsOnlySetFlags(t *testing.T) {
	var got []models.ListFoldersOptions
	mock := &mockClient{listFn: func(o models.ListFoldersOptions) (*models.ListFoldersResponse, error) {
		got = append(got, o)
		return &models.ListFoldersResponse{
			Folders:    []models.FolderSummary{{FolderID: "f1", Name: "A", Path: "/A", ItemCount: 3}},
			NextOffset: intPtr(1),
		}, nil
	}}
	buf, err := run(t, mock, "list")
	require.NoError(t, err)
	assert.Equal(t, []models.ListFoldersOptions{{}}, got)
	assert.Regexp(t, `f1\s+/A\s+3`, buf.String())
	assert.Contains(t, buf.String(), "More folders: --offset 1")
}

func TestListJSON(t *testing.T) {
	mock := &mockClient{listFn: func(models.ListFoldersOptions) (*models.ListFoldersResponse, error) {
		return &models.ListFoldersResponse{
			Folders: []models.FolderSummary{{FolderID: "f1", Name: "A", Path: "/A", ItemCount: 3}},
		}, nil
	}}
	buf, err := run(t, mock, "list", "-o", "json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"folders":[{"folder_id":"f1","name":"A","path":"/A","item_count":3}]}`, buf.String())
}

// --parent-path is resolved one segment at a time, matching names exactly: "Research Archive"
// contains "Research" but must not be picked for it.
func TestListResolvesNestedPath(t *testing.T) {
	var calls []models.ListFoldersOptions
	mock := &mockClient{listFn: func(o models.ListFoldersOptions) (*models.ListFoldersResponse, error) {
		calls = append(calls, o)
		switch {
		case o.ParentFolderID == "" && o.NameContains == "Research" && o.Offset == 0:
			return &models.ListFoldersResponse{
				Folders:    []models.FolderSummary{{FolderID: "wrong", Name: "Research Archive"}},
				NextOffset: intPtr(1),
			}, nil
		case o.ParentFolderID == "" && o.NameContains == "Research" && o.Offset == 1:
			return &models.ListFoldersResponse{Folders: []models.FolderSummary{{FolderID: "r1", Name: "Research"}}}, nil
		case o.ParentFolderID == "r1" && o.NameContains == "RWA":
			return &models.ListFoldersResponse{Folders: []models.FolderSummary{
				{FolderID: "wrong2", Name: "RWA old"}, {FolderID: "f1", Name: "RWA"},
			}}, nil
		case o.ParentFolderID == "f1":
			return &models.ListFoldersResponse{Folders: []models.FolderSummary{
				{FolderID: "c1", Name: "Tokens", Path: "/Research/RWA/Tokens", ItemCount: 2},
			}}, nil
		}
		t.Fatalf("unexpected call %+v", o)
		return nil, nil
	}}
	buf, err := run(t, mock, "list", "--parent-path", "/Research/RWA/", "--name-contains", "tok",
		"--limit", "10", "--offset", "5")
	require.NoError(t, err)
	require.Len(t, calls, 4)
	assert.Equal(t, models.ListFoldersOptions{ParentFolderID: "f1", NameContains: "tok", Limit: 10, Offset: 5},
		calls[3])
	assert.Regexp(t, `c1\s+/Research/RWA/Tokens\s+2`, buf.String())
}

func TestListMissingPathSegment(t *testing.T) {
	mock := &mockClient{listFn: func(o models.ListFoldersOptions) (*models.ListFoldersResponse, error) {
		if o.ParentFolderID == "" {
			return &models.ListFoldersResponse{Folders: []models.FolderSummary{{FolderID: "r1", Name: "Research"}}}, nil
		}
		return &models.ListFoldersResponse{Folders: []models.FolderSummary{{FolderID: "x", Name: "Nope2"}}}, nil
	}}
	_, err := run(t, mock, "list", "--parent-path", "/Research/Nope")
	require.Error(t, err)
	assert.Equal(t, "no folder at /Research/Nope", err.Error())
}

func TestGet(t *testing.T) {
	var gotID string
	var gotOpts models.GetFolderOptions
	mock := &mockClient{getFn: func(id string, o models.GetFolderOptions) (*models.GetFolderResponse, error) {
		gotID, gotOpts = id, o
		return &models.GetFolderResponse{
			Folder:     *rwa,
			Items:      []models.FolderItem{{Type: "query", ID: 7, Name: "Q"}},
			TotalItems: 1,
		}, nil
	}}
	buf, err := run(t, mock, "get", "f1", "--limit", "100")
	require.NoError(t, err)
	assert.Equal(t, "f1", gotID)
	assert.Equal(t, models.GetFolderOptions{Limit: 100}, gotOpts)
	assert.Contains(t, buf.String(), "Path:        /Research/RWA\n")
	assert.Regexp(t, `query\s+7\s+Q\s+false`, buf.String())
}

func TestCreate(t *testing.T) {
	var got models.CreateFolderRequest
	mock := &mockClient{createFn: func(r models.CreateFolderRequest) (*models.Folder, error) {
		got = r
		return rwa, nil
	}}
	buf, err := run(t, mock, "create", "--name", "RWA", "--parent-id", "p1", "--color", "#E5477D")
	require.NoError(t, err)
	assert.Equal(t, models.CreateFolderRequest{
		Name: "RWA", ParentFolderID: strPtr("p1"), Color: strPtr("#E5477D"),
	}, got)
	assert.Equal(t, "Created folder /Research/RWA (f1)\n", buf.String())
}

func TestCreateConflictNamesExistingFolder(t *testing.T) {
	mock := &mockClient{createFn: func(models.CreateFolderRequest) (*models.Folder, error) {
		return nil, apiError(409, "folder_already_exists", models.FolderErrorDetails{ExistingFolderID: strPtr("ex1")})
	}}
	_, err := run(t, mock, "create", "--name", "RWA")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ex1")
	var apiErr *dune.APIError
	assert.ErrorAs(t, err, &apiErr)
}

func TestMoveToRootSendsNil(t *testing.T) {
	var gotID string
	got := models.MoveFolderRequest{ParentFolderID: strPtr("unset")}
	mock := &mockClient{moveFn: func(id string, r models.MoveFolderRequest) (*models.Folder, error) {
		gotID, got = id, r
		return rwa, nil
	}}
	buf, err := run(t, mock, "move", "f1", "--to-root")
	require.NoError(t, err)
	assert.Equal(t, "f1", gotID)
	assert.Nil(t, got.ParentFolderID)
	assert.Equal(t, "Moved folder /Research/RWA (f1)\n", buf.String())

	_, err = run(t, mock, "move", "f1", "--to", "f2")
	require.NoError(t, err)
	assert.Equal(t, strPtr("f2"), got.ParentFolderID)
}

func TestMoveNeedsOneDestination(t *testing.T) {
	_, err := run(t, &mockClient{}, "move", "f1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[to to-root]")

	_, err = run(t, &mockClient{}, "move", "f1", "--to", "f2", "--to-root")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestUpdateEmptyDescriptionClearsIt(t *testing.T) {
	var got models.UpdateFolderRequest
	mock := &mockClient{updateFn: func(_ string, r models.UpdateFolderRequest) (*models.Folder, error) {
		got = r
		return rwa, nil
	}}
	buf, err := run(t, mock, "update", "f1", "--description", "")
	require.NoError(t, err)
	assert.Equal(t, models.UpdateFolderRequest{Description: strPtr("")}, got)
	assert.Equal(t, "Updated folder /Research/RWA (f1)\n", buf.String())

	_, err = run(t, mock, "update", "f1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestMoveContent(t *testing.T) {
	var got models.MoveContentRequest
	mock := &mockClient{moveContentFn: func(r models.MoveContentRequest) (*models.MoveContentResponse, error) {
		got = r
		return &models.MoveContentResponse{MovedQueryIDs: []int64{1, 2}, TargetFolder: rwa}, nil
	}}
	buf, err := run(t, mock, "move-content", "--query-ids", "1,2", "--to", "f1")
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, got.QueryIDs)
	assert.Empty(t, got.DashboardIDs)
	assert.Equal(t, strPtr("f1"), got.TargetFolderID)
	assert.Equal(t, "Moved 2 queries and 0 dashboards to /Research/RWA\n", buf.String())

	_, err = run(t, mock, "move-content", "--to-root")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--query-ids, --dashboard-ids")
}

func TestDeleteRecursive(t *testing.T) {
	var gotID string
	var gotRecursive bool
	mock := &mockClient{deleteFn: func(id string, recursive bool) (*models.DeleteFolderResponse, error) {
		gotID, gotRecursive = id, recursive
		return &models.DeleteFolderResponse{FolderID: "f1", Path: "/A", DeletedFolderCount: 2,
			ArchivedQueryCount: 3, ArchivedDashboardCount: 1}, nil
	}}
	buf, err := run(t, mock, "delete", "f1", "--recursive")
	require.NoError(t, err)
	assert.Equal(t, "f1", gotID)
	assert.True(t, gotRecursive)
	assert.Equal(t, "Deleted folder /A: 2 folders deleted, 3 queries and 1 dashboards archived\n", buf.String())
}

// The server's refusal does not name the CLI flag, so the CLI adds it with the counts.
func TestDeleteNotEmptyNamesTheFlag(t *testing.T) {
	var gotRecursive = true
	mock := &mockClient{deleteFn: func(_ string, recursive bool) (*models.DeleteFolderResponse, error) {
		gotRecursive = recursive
		return nil, apiError(409, "folder_not_empty", models.FolderErrorDetails{
			SubfolderCount: intPtr(1), QueryCount: intPtr(2), DashboardCount: intPtr(0),
		})
	}}
	_, err := run(t, mock, "delete", "f1")
	require.Error(t, err)
	assert.False(t, gotRecursive)
	assert.Contains(t, err.Error(), "1 subfolders, 2 queries and 0 dashboards")
	assert.Contains(t, err.Error(), "pass --recursive")
}

// Other errors, matched on code rather than text, get no hint.
func TestDeleteOtherErrorHasNoHint(t *testing.T) {
	mock := &mockClient{deleteFn: func(string, bool) (*models.DeleteFolderResponse, error) {
		return nil, apiError(404, "folder_not_found", nil)
	}}
	_, err := run(t, mock, "delete", "f1")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "--recursive")
}
