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

func TestSearchSendsMCPArguments(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchTables": searchTablesResult}}
	buf, err := runMCP(fake, "search", "--query", "dex trades", "--categories", "spell", "--dataset-types", "spell",
		"--include-schema", "--limit", "5", "--owner-scope", "me")
	require.NoError(t, err)

	// --owner-scope is deprecated and not sent.
	assert.JSONEq(t, `{"query":"dex trades","categories":["spell"],"datasetTypes":["spell"],"includeSchema":true,
		"limit":5}`, fake.Last().Args)
	assert.Regexp(t, `dex.trades\s+spell\s+spell\s+blockchain, block_month\s+ethereum, arbitrum`, buf.String())
	assert.Contains(t, buf.String(), "1 of 144 results")
}

func TestSearchJSONKeepsTheAPIShape(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchTables": searchTablesResult}}
	buf, err := runMCP(fake, "search", "--query", "dex", "-o", "json")
	require.NoError(t, err)

	assert.JSONEq(t, `{"total":144,"results":[{"full_name":"dex.trades","category":"spell","dataset_type":"spell",
		"blockchains":["ethereum","arbitrum"],"partition_columns":["blockchain","block_month"],
		"columns":[{"name":"blockchain","type":"varchar"}]}],
		"pagination":{"limit":1,"offset":0,"next_offset":1,"has_more":true}}`, buf.String())
}

func TestSearchError(t *testing.T) {
	fake := &dunemcptest.Fake{Errors: map[string]error{"searchTables": errors.New("unauthorized")}}
	_, err := runMCP(fake, "search", "--query", "test")
	assert.EqualError(t, err, "unauthorized")
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
