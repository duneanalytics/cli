package contract_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/cmd/contract"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	root, buf := dunemcptest.NewRoot(contract.NewContractCmd(), fake)
	root.SetArgs(append([]string{"contract"}, args...))
	return buf, root.Execute()
}

func TestSubmitOneContract(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"submitContractsForDecoding": `{"results":[
		{"index":0,"submission_id":"s1","status":"pending"}],"status_url":"https://dune.com/contracts/submissions"}`}}
	buf, err := run(fake, "submit", "--blockchain", "ethereum", "--address", "0x1", "--project", "uni",
		"--contract-name", "Token", "--abi", `[{"type":"event","name":"Transfer","inputs":[]}]`, "--proxy")
	require.NoError(t, err)
	// The ABI goes as JSON, not as a string holding JSON.
	assert.JSONEq(t, `{"submissions":[{"blockchain_name":"ethereum","address":"0x1","project_name":"uni",
		"contract_name":"Token","is_proxy":true,"abi":[{"type":"event","name":"Transfer","inputs":[]}]}]}`,
		fake.Last().Args)
	assert.Regexp(t, `0\s+s1\s+pending`, buf.String())
}

func TestSubmitBatchFailsWhenAnyItemFails(t *testing.T) {
	batch := filepath.Join(t.TempDir(), "batch.json")
	require.NoError(t, os.WriteFile(batch, []byte(`[{"blockchain_name":"base"},{"blockchain_name":"base"}]`), 0o600))
	fake := &dunemcptest.Fake{Results: map[string]string{"submitContractsForDecoding": `{"results":[
		{"index":0,"submission_id":"s1","status":"pending","replayed":true},
		{"index":1,"error":"abi must be valid JSON"}]}`}}
	buf, err := run(fake, "submit", "--batch", "@"+batch)

	assert.EqualError(t, err, "1 of 2 submissions failed")
	assert.JSONEq(t, `{"submissions":[{"blockchain_name":"base"},{"blockchain_name":"base"}]}`, fake.Last().Args)
	assert.Regexp(t, `0\s+s1\s+pending \(existing\)`, buf.String())
	assert.Contains(t, buf.String(), "abi must be valid JSON")
}

func TestSubmitRequiresContractFields(t *testing.T) {
	_, err := run(&dunemcptest.Fake{}, "submit", "--blockchain", "ethereum", "--address", "0x1",
		"--project", "uni", "--contract-name", "Token")
	assert.EqualError(t, err, "--abi is required unless you pass --batch")

	_, err = run(&dunemcptest.Fake{}, "submit", "--batch", "[]", "--address", "0x1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestSubmissionsFilters(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listContractSubmissions": `{"submissions":[
		{"id":"s1","blockchain_name":"base","address":"0x1","project_name":"uni","contract_name":"Token",
		 "status":"processed","submission_type":"new","created_at":"2026-11-01T00:00:00Z"}],"total":3,"next_cursor":"c2"}`}}
	buf, err := run(fake, "submissions", "--status", "processed", "--project", "uni", "--limit", "1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"processed","project_name":"uni","limit":1}`, fake.Last().Args)
	assert.Regexp(t, `s1\s+base\s+0x1\s+uni\.Token\s+new\s+processed`, buf.String())
	assert.Contains(t, buf.String(), "More submissions: --cursor c2")
}
