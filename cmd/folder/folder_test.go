package folder_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/duneanalytics/cli/cmd/folder"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const folderResult = `{"folder":{"folder_id":"f1","name":"RWA","path":"/Research/RWA","parent_folder_id":null,
	"description":null,"icon":"folder","color":"gray"}}`

func run(t *testing.T, fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	t.Helper()
	root, buf := dunemcptest.NewRoot(folder.NewFolderCmd(), fake)
	root.SetArgs(append([]string{"folder"}, args...))
	return buf, root.Execute()
}

func TestListSendsOnlySetFlags(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listDuneFolders": `{"folders":[
		{"folder_id":"f1","name":"A","path":"/A","item_count":3}],"next_offset":1}`}}
	buf, err := run(t, fake, "list")
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, fake.Last().Args)
	assert.Regexp(t, `f1\s+/A\s+3`, buf.String())
	assert.Contains(t, buf.String(), "More folders: --offset 1")
}

func TestCreate(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createDuneFolder": folderResult}}
	buf, err := run(t, fake, "create", "--name", "RWA", "--parent-id", "p1", "--color", "#E5477D")
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"RWA","parent_folder_id":"p1","color":"#E5477D"}`, fake.Last().Args)
	assert.Equal(t, "Created folder /Research/RWA (f1)\n", buf.String())
}

func TestMoveToRootSendsNull(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"moveDuneFolder": folderResult}}
	_, err := run(t, fake, "move", "f1", "--to-root")
	require.NoError(t, err)
	assert.JSONEq(t, `{"folder_id":"f1","parent_folder_id":null}`, fake.Last().Args)
}

func TestMoveNeedsOneDestination(t *testing.T) {
	_, err := run(t, &dunemcptest.Fake{}, "move", "f1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[to to-root]")

	_, err = run(t, &dunemcptest.Fake{}, "move", "f1", "--to", "f2", "--to-root")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestUpdateEmptyDescriptionClearsIt(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDuneFolder": folderResult}}
	_, err := run(t, fake, "update", "f1", "--description", "")
	require.NoError(t, err)
	assert.JSONEq(t, `{"folder_id":"f1","description":null}`, fake.Last().Args)

	_, err = run(t, fake, "update", "f1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestMoveContent(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"moveContentToFolder": `{"moved_query_ids":[1,2],
		"moved_dashboard_ids":[],"target_folder":{"folder_id":"f1","name":"RWA","path":"/Research/RWA"}}`}}
	buf, err := run(t, fake, "move-content", "--query-ids", "1,2", "--to", "f1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_ids":[1,2],"target_folder_id":"f1"}`, fake.Last().Args)
	assert.Equal(t, "Moved 2 queries and 0 dashboards to /Research/RWA\n", buf.String())

	_, err = run(t, fake, "move-content", "--to-root")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--query-ids, --dashboard-ids")
}

func TestDeleteRecursive(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"deleteDuneFolder": `{"deleted_folder_id":"f1",
		"path":"/A","deleted_folder_count":2,"archived_query_count":3,"archived_dashboard_count":1}`}}
	buf, err := run(t, fake, "delete", "f1", "--recursive")
	require.NoError(t, err)
	assert.JSONEq(t, `{"folder_id":"f1","recursive":true}`, fake.Last().Args)
	assert.Equal(t, "Deleted folder /A: 2 folders deleted, 3 queries and 1 dashboards archived\n", buf.String())
}

// The server's refusal does not name the CLI flag, so the CLI adds it.
func TestDeleteNotEmptyNamesTheFlag(t *testing.T) {
	fake := &dunemcptest.Fake{Errors: map[string]error{"deleteDuneFolder": errors.New(
		"Folder /A is not empty: it has 1 subfolder(s), 2 query(ies) and 0 dashboard(s).")}}
	_, err := run(t, fake, "delete", "f1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Folder /A is not empty")
	assert.Contains(t, err.Error(), "pass --recursive")
	assert.JSONEq(t, `{"folder_id":"f1"}`, fake.Last().Args)
}
