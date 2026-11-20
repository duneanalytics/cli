package dataset_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/duneanalytics/cli/cmd/dataset"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	dune.DuneClient
	searchDatasetsFn          func(models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error)
	searchByContractAddressFn func(models.SearchDatasetsByContractAddressRequest) (*models.SearchDatasetsResponse, error)
}

func (m *mockClient) SearchDatasets(req models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
	return m.searchDatasetsFn(req)
}

func (m *mockClient) SearchDatasetsByContractAddress(req models.SearchDatasetsByContractAddressRequest) (*models.SearchDatasetsResponse, error) {
	return m.searchByContractAddressFn(req)
}

func newTestRoot(mock dune.DuneClient) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(dataset.NewDatasetCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}

func runMCP(fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	root, buf := dunemcptest.NewRoot(dataset.NewDatasetCmd(), fake)
	root.SetArgs(append([]string{"dataset"}, args...))
	return buf, root.Execute()
}

const searchTablesResult = `{"total":144,"results":[{"full_name":"dex.trades","category":"spell",
	"dataset_type":"spell","blockchains":["ethereum","arbitrum"],"partition_columns":["blockchain","block_month"],
	"columns":[{"name":"blockchain","type":"varchar"}]}],
	"pagination":{"limit":1,"offset":0,"nextOffset":1,"hasMore":true}}`

func strPtr(s string) *string { return &s }

// runSearch runs `dune dataset search` against a REST mock that returns two
// tables and an MCP fake.
func runSearch(t *testing.T, fake *dunemcptest.Fake, args ...string) (models.SearchDatasetsRequest, string, string, error) {
	t.Helper()
	var got models.SearchDatasetsRequest
	next := int32(2)
	mock := &mockClient{searchDatasetsFn: func(req models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
		got = req
		return &models.SearchDatasetsResponse{
			Total: 144,
			Results: []models.SearchDatasetResult{
				{FullName: "dex.trades", Category: "spell", DatasetType: strPtr("spell"),
					Blockchains: []string{"ethereum", "arbitrum"}},
				{FullName: "my.private_table", Category: "community", Visibility: strPtr("private")},
			},
			Pagination: models.SearchDatasetsPagination{Limit: 2, NextOffset: &next, HasMore: true},
		}, nil
	}}
	root, buf := newTestRoot(mock)
	root.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		cmdutil.SetClient(cmd, mock)
		cmdutil.SetMCP(cmd, fake)
	}
	var stderr bytes.Buffer
	root.SetErr(&stderr)
	root.SetArgs(append([]string{"dataset", "search"}, args...))
	err := root.Execute()
	return got, buf.String(), stderr.String(), err
}

func TestSearchSendsEveryFlagToTheAPI(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchTables": searchTablesResult}}
	req, out, _, err := runSearch(t, fake, "--query", "dex trades", "--categories", "spell",
		"--dataset-types", "spell", "--include-schema", "--limit", "5", "--owner-scope", "me", "--include-private")
	require.NoError(t, err)

	assert.Equal(t, "dex trades", *req.Query)
	assert.Equal(t, []string{"spell"}, req.Categories)
	assert.Equal(t, []string{"spell"}, req.DatasetTypes)
	assert.Equal(t, "me", *req.OwnerScope)
	assert.True(t, *req.IncludePrivate)
	assert.True(t, *req.IncludeSchema)
	assert.Equal(t, int32(5), *req.Limit)
	// The MCP gets the same search without the flags it does not support.
	assert.JSONEq(t, `{"query":"dex trades","categories":["spell"],"datasetTypes":["spell"],"limit":5}`,
		fake.Last().Args)
	assert.Regexp(t, `dex.trades\s+spell\s+spell\s+blockchain, block_month\s+ethereum, arbitrum`, out)
	assert.Contains(t, out, "my.private_table")
	assert.Contains(t, out, "2 of 144 results")
}

func TestSearchJSONIsTheAPIShapePlusColumns(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchTables": searchTablesResult}}
	_, out, _, err := runSearch(t, fake, "--query", "dex", "-o", "json")
	require.NoError(t, err)

	assert.JSONEq(t, `{"total":144,"results":[{"full_name":"dex.trades","category":"spell","dataset_type":"spell",
		"blockchains":["ethereum","arbitrum"],"partition_columns":["blockchain","block_month"],
		"columns":[{"name":"blockchain","type":"varchar"}]},
		{"full_name":"my.private_table","category":"community","visibility":"private"}],
		"pagination":{"limit":2,"offset":0,"next_offset":2,"has_more":true}}`, out)
}

func TestSearchWithoutTheMCPKeepsTheAPIResults(t *testing.T) {
	fake := &dunemcptest.Fake{Errors: map[string]error{"searchTables": errors.New("mcp unavailable")}}
	_, out, stderr, err := runSearch(t, fake, "--query", "dex")
	require.NoError(t, err)
	assert.Contains(t, out, "dex.trades")
	assert.Contains(t, out, "2 of 144 results")
	assert.Equal(t, "warning: could not read partition columns: mcp unavailable\n", stderr)
}

// Scripts iterate over .results, so no hits must print [] rather than null.
func TestSearchWithNoHitsPrintsAnEmptyList(t *testing.T) {
	mock := &mockClient{searchDatasetsFn: func(models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
		return &models.SearchDatasetsResponse{Results: []models.SearchDatasetResult{}}, nil
	}}
	root, buf := newTestRoot(mock)
	root.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		cmdutil.SetClient(cmd, mock)
		cmdutil.SetMCP(cmd, &dunemcptest.Fake{Results: map[string]string{"searchTables": `{"results":[]}`}})
	}
	root.SetArgs([]string{"dataset", "search", "--query", "nothing", "-o", "json"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"total":0,"results":[],"pagination":{"limit":0,"offset":0,"has_more":false}}`, buf.String())
}

func TestSearchError(t *testing.T) {
	mock := &mockClient{searchDatasetsFn: func(models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
		return nil, errors.New("unauthorized")
	}}
	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"dataset", "search", "--query", "test"})
	assert.EqualError(t, root.Execute(), "unauthorized")
}

func TestBlockchains(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listBlockchains": `{"total":43,"results":[
		{"blockchain":"ethereum","table_count":3024}],"pagination":{"limit":1,"offset":0,"nextOffset":1,"hasMore":true}}`}}
	buf, err := runMCP(fake, "blockchains", "--categories", "decoded", "--limit", "1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"categories":["decoded"],"limit":1}`, fake.Last().Args)
	assert.Regexp(t, `ethereum\s+3024`, buf.String())
	assert.Contains(t, buf.String(), "1 of 43 blockchains")
}

func TestSize(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"getTableSize": `{"table_name":"dex.trades",
		"total_size_bytes":2147483648,"underlying_tables":[
		{"table_name":"delta_prod:dex.trades","size_bytes":2147483648},
		{"table_name":"delta_prod:dex.other","error":"not found"}],"warnings":["one table had no size"]}`}}
	buf, err := runMCP(fake, "size", "dex.trades")
	require.NoError(t, err)
	assert.JSONEq(t, `{"table_name":"dex.trades"}`, fake.Last().Args)
	out := buf.String()
	assert.Contains(t, out, "dex.trades: 2.0 GB")
	assert.Regexp(t, `delta_prod:dex.other\s+not found`, out)
	assert.Contains(t, out, "Warning: one table had no size")
}
