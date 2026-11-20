package contract_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/duneanalytics/cli/cmd/contract"
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
	submitFn func(models.SubmitContractsRequest) (*models.SubmitContractsResponse, error)
	listFn   func(models.ListContractSubmissionsOptions) (*models.ListContractSubmissionsResponse, error)
}

func (m *mockClient) SubmitContracts(req models.SubmitContractsRequest) (*models.SubmitContractsResponse, error) {
	return m.submitFn(req)
}

func (m *mockClient) ListContractSubmissions(
	opts models.ListContractSubmissionsOptions,
) (*models.ListContractSubmissionsResponse, error) {
	return m.listFn(opts)
}

func run(mock dune.DuneClient, args ...string) (*bytes.Buffer, error) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(contract.NewContractCmd())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"contract"}, args...))
	return &buf, root.Execute()
}

func ptr[T any](v T) *T { return &v }

func TestSubmitOneContract(t *testing.T) {
	var got models.SubmitContractsRequest
	mock := &mockClient{submitFn: func(req models.SubmitContractsRequest) (*models.SubmitContractsResponse, error) {
		got = req
		return &models.SubmitContractsResponse{Results: []models.ContractSubmissionResult{
			{Index: 0, SubmissionID: ptr("s1"), Status: ptr("pending")},
		}}, nil
	}}
	buf, err := run(mock, "submit", "--blockchain", "ethereum", "--address", "0x1", "--project", "uni",
		"--contract-name", "Token", "--abi", `[{"type":"event","name":"Transfer","inputs":[]}]`, "--proxy")
	require.NoError(t, err)
	require.Len(t, got.Submissions, 1)
	item := got.Submissions[0]
	// The ABI goes as JSON, not as a string holding JSON.
	assert.JSONEq(t, `[{"type":"event","name":"Transfer","inputs":[]}]`, string(item.ABI))
	item.ABI = nil
	// Unset optional flags stay nil or false.
	assert.Equal(t, models.ContractSubmissionInput{
		BlockchainName: "ethereum", Address: "0x1", ProjectName: "uni", ContractName: "Token", IsProxy: true,
	}, item)
	assert.Regexp(t, `0\s+s1\s+pending`, buf.String())
}

func TestSubmitMapsEveryFlag(t *testing.T) {
	var got models.SubmitContractsRequest
	mock := &mockClient{submitFn: func(req models.SubmitContractsRequest) (*models.SubmitContractsResponse, error) {
		got = req
		return &models.SubmitContractsResponse{Results: []models.ContractSubmissionResult{
			{Index: 0, SubmissionID: ptr("s1"), Status: ptr("needs_manual_review")},
		}}, nil
	}}
	_, err := run(mock, "submit", "--blockchain", "base", "--address", "0x2", "--project", "new",
		"--contract-name", "Pool", "--abi", `"[]"`, "--multiple-instances", "--factory", "--manual-abi",
		"--type", "rename", "--reason", "typo", "--old-project", "old", "--old-contract-name", "OldPool",
		"--idempotency-key", "k1")
	require.NoError(t, err)
	require.Len(t, got.Submissions, 1)
	assert.Equal(t, models.ContractSubmissionInput{
		BlockchainName: "base", Address: "0x2", ProjectName: "new", ContractName: "Pool", ABI: []byte(`"[]"`),
		HasMultipleInstances: true, IsCreatedByFactory: true, IsManualABI: true,
		SubmissionType: ptr("rename"), ResubmissionReason: ptr("typo"), OldProjectName: ptr("old"),
		OldContractName: ptr("OldPool"), IdempotencyKey: ptr("k1"),
	}, got.Submissions[0])
}

func TestSubmitBatchFailsWhenAnyItemFails(t *testing.T) {
	batch := filepath.Join(t.TempDir(), "batch.json")
	require.NoError(t, os.WriteFile(batch, []byte(`[{"blockchain_name":"base","submission_type":"delete"},
		{"blockchain_name":"base","abi":[]}]`), 0o600))
	var got models.SubmitContractsRequest
	mock := &mockClient{submitFn: func(req models.SubmitContractsRequest) (*models.SubmitContractsResponse, error) {
		got = req
		return &models.SubmitContractsResponse{Results: []models.ContractSubmissionResult{
			{Index: 0, SubmissionID: ptr("s1"), Status: ptr("pending"), Replayed: true},
			{Index: 1, Error: ptr("abi must be valid JSON")},
		}}, nil
	}}
	buf, err := run(mock, "submit", "--batch", "@"+batch)

	assert.EqualError(t, err, "1 of 2 submissions failed")
	require.Len(t, got.Submissions, 2)
	assert.Equal(t, "base", got.Submissions[0].BlockchainName)
	assert.Equal(t, ptr("delete"), got.Submissions[0].SubmissionType)
	assert.JSONEq(t, `[]`, string(got.Submissions[1].ABI))
	assert.Regexp(t, `0\s+s1\s+pending \(existing\)`, buf.String())
	assert.Regexp(t, `1\s+abi must be valid JSON`, buf.String())
}

func TestSubmitRequiresContractFields(t *testing.T) {
	_, err := run(&mockClient{}, "submit", "--blockchain", "ethereum", "--address", "0x1",
		"--project", "uni", "--contract-name", "Token")
	assert.EqualError(t, err, "--abi is required unless you pass --batch")

	_, err = run(&mockClient{}, "submit", "--batch", "[]", "--address", "0x1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestSubmitReportsAPIError(t *testing.T) {
	mock := &mockClient{submitFn: func(models.SubmitContractsRequest) (*models.SubmitContractsResponse, error) {
		return nil, &dune.APIError{StatusCode: 400, Message: "too many submissions"}
	}}
	_, err := run(mock, "submit", "--batch", "[{}]")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too many submissions")
}

func TestSubmissionsFilters(t *testing.T) {
	var got models.ListContractSubmissionsOptions
	mock := &mockClient{listFn: func(
		opts models.ListContractSubmissionsOptions,
	) (*models.ListContractSubmissionsResponse, error) {
		got = opts
		return &models.ListContractSubmissionsResponse{Total: 3, NextCursor: ptr("c2"),
			Submissions: []models.ContractSubmission{{
				ID: "s1", BlockchainName: "base", Address: "0x1", ProjectName: "uni", ContractName: "Token",
				Status: "processed", SubmissionType: "new", CreatedAt: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			}}}, nil
	}}
	buf, err := run(mock, "submissions", "--status", "processed", "--project", "uni", "--limit", "1")
	require.NoError(t, err)
	assert.Equal(t, models.ListContractSubmissionsOptions{Status: "processed", ProjectName: "uni", Limit: 1}, got)
	assert.Regexp(t, `s1\s+base\s+0x1\s+uni\.Token\s+new\s+processed\s+2026-11-01T00:00:00Z`, buf.String())
	assert.Contains(t, buf.String(), "1 of 3 submissions")
	assert.Contains(t, buf.String(), "More submissions: --cursor c2")
}

func TestSubmissionsJSON(t *testing.T) {
	mock := &mockClient{listFn: func(
		models.ListContractSubmissionsOptions,
	) (*models.ListContractSubmissionsResponse, error) {
		return &models.ListContractSubmissionsResponse{Submissions: []models.ContractSubmission{}}, nil
	}}
	buf, err := run(mock, "submissions", "-o", "json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"submissions":[],"total":0}`, buf.String())
}
