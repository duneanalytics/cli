package query_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMCPRoot(fake *dunemcptest.Fake) (*cobra.Command, *bytes.Buffer) {
	return dunemcptest.NewRoot(query.NewQueryCmd(), fake)
}

func TestCreateSavesQueryByDefault(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createDuneQuery": `{"query_id":4125432}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create", "--name", "Test", "--sql", "SELECT 1"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "Created query 4125432\n", buf.String())
	// The MCP tool makes queries temporary unless told otherwise; the CLI saves them.
	assert.JSONEq(t, `{"name":"Test","query":"SELECT 1","is_temp":false}`, fake.Last().Args)
}

func TestCreateAllFlags(t *testing.T) {
	params := filepath.Join(t.TempDir(), "params.json")
	require.NoError(t, os.WriteFile(params, []byte(`[{"key":"n","type":"number","value":"10"}]`), 0o600))
	fake := &dunemcptest.Fake{Results: map[string]string{"createDuneQuery": `{"query_id":1}`}}
	root, _ := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "SELECT {{n}}", "--description", "d",
		"--private", "--temp", "--folder-name", "Research", "--parameters", "@" + params})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"name":"T","query":"SELECT {{n}}","description":"d","is_private":true,"is_temp":true,
		"folder_name":"Research","parameters":[{"key":"n","type":"number","value":"10"}]}`, fake.Last().Args)
}

func TestCreateFolderFlagsConflict(t *testing.T) {
	root, _ := newMCPRoot(&dunemcptest.Fake{})
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "S", "--folder-id", "a", "--folder-name", "b"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestCreateJSONOutput(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createDuneQuery": `{"query_id":4125432}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create", "--name", "Test", "--sql", "SELECT 1", "-o", "json"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"query_id":4125432}`, buf.String())
}

func TestCreateToolError(t *testing.T) {
	fake := &dunemcptest.Fake{Errors: map[string]error{
		"createDuneQuery": errors.New("folder_ambiguous: pick a folder_id"),
	}}
	root, _ := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create", "--name", "T", "--sql", "S", "--folder-name", "x"})
	assert.EqualError(t, root.Execute(), "folder_ambiguous: pick a folder_id")
}

func TestUpdateSendsFalseValues(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDuneQuery": `{"queryId":7}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "update", "7", "--private=false", "--archived=false"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "Updated query 7\n", buf.String())
	assert.JSONEq(t, `{"queryId":7,"is_private":false,"is_archived":false}`, fake.Last().Args)
}

func TestUpdateEmptyDescriptionClearsIt(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDuneQuery": `{"queryId":7}`}}
	root, _ := newMCPRoot(fake)
	root.SetArgs([]string{"query", "update", "7", "--description", "", "--tags", "defi,uniswap"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"queryId":7,"description":null,"tags":["defi","uniswap"]}`, fake.Last().Args)
}

func TestUpdateNoFlags(t *testing.T) {
	root, _ := newMCPRoot(&dunemcptest.Fake{})
	root.SetArgs([]string{"query", "update", "1"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestUpdateNonIntegerID(t *testing.T) {
	root, _ := newMCPRoot(&dunemcptest.Fake{})
	root.SetArgs([]string{"query", "update", "abc", "--name", "X"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query ID")
}

func TestUpdateJSONOutput(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDuneQuery": `{"queryId":7}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "update", "7", "--name", "New", "-o", "json"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"query_id":7}`, buf.String())
}

const createAndRunResult = `{
	"query":{"query_id":11,"url":"https://dune.com/queries/11"},
	"execution":{"execution_id":"01EXEC","state":"COMPLETED","engine_used":"small"},
	"visualization":{"queryId":11,"visualizationId":22,"interactiveChartUrl":"https://dune.com/embeds/11/22",
		"chartJpegUrl":"https://img"},
	"result_preview":{"executionId":"01EXEC","state":"COMPLETED",
		"resultMetadata":{"columns":[{"name":"day","type":"date"},{"name":"n","type":"bigint"}],"totalRowCount":3,
			"executionCostCredits":"0.01"},
		"data":{"rows":[{"day":"2026-11-01","n":5},{"day":"2026-11-02","n":8}]}}}`

func TestCreateAndRun(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createAndExecuteQuery": createAndRunResult}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create-and-run", "--name", "Daily", "--sql", "SELECT 1", "--performance", "small",
		"--limit", "2", "--chart-type", "counter", "--config", `{"column":"n"}`, "--viz-name", "Count"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"name":"Daily","query":"SELECT 1","is_temp":false,"performance":"small","max_rows_returned":2,
		"visualization":{"chartType":"counter","visualizationName":"Count","config":{"chart_type":"counter","column":"n"}}}`,
		fake.Last().Args)
	out := buf.String()
	assert.Contains(t, out, "Created query 11: https://dune.com/queries/11")
	assert.Contains(t, out, "Visualization 22: https://dune.com/embeds/11/22")
	// Rows come out in the order of the result's columns.
	assert.Regexp(t, `day\s+n\n2026-11-01\s+5\n2026-11-02\s+8\n`, out)
	assert.Contains(t, out, "2 of 3 rows")
}

func TestCreateAndRunReportsFailuresAfterPrinting(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createAndExecuteQuery": `{
		"query":{"query_id":11,"url":"https://dune.com/queries/11"},
		"execution":{"execution_id":"01EXEC","state":"FAILED","engine_used":"small"},
		"result_preview":{"executionId":"01EXEC","state":"FAILED",
			"errorMessage":"line 1:8: Column 'x' cannot be resolved"}}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create-and-run", "--name", "Bad", "--sql", "SELECT x"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Column 'x' cannot be resolved")
	assert.Contains(t, buf.String(), "Created query 11")
}

func TestSearchOwnedOnly(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchDuneQueries": `{"results":[
		{"query_id":4700,"name":"DEX Volume","owner":{"type":"team","id":1,"handle":"remix","name":"remix"}}]}`}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "search", "--query", "dex", "--owned-only"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"search_string":"dex","scope":"owned_only"}`, fake.Last().Args)
	assert.Regexp(t, `4700\s+DEX Volume\s+remix`, buf.String())
}

func TestCreateAndRunJSONIsTheToolResult(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"createAndExecuteQuery": createAndRunResult}}
	root, buf := newMCPRoot(fake)
	root.SetArgs([]string{"query", "create-and-run", "--name", "Daily", "--sql", "SELECT 1", "-o", "json"})
	require.NoError(t, root.Execute())
	var got, want any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.NoError(t, json.Unmarshal([]byte(createAndRunResult), &want))
	assert.Equal(t, want, got)
}
