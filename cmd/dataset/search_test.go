package dataset_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/duneanalytics/cli/cmd/dataset"
	"github.com/duneanalytics/cli/cmdutil"
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

func TestSearchSuccess(t *testing.T) {
	dt := "spell"
	var gotReq models.SearchDatasetsRequest

	mock := &mockClient{
		searchDatasetsFn: func(req models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
			gotReq = req
			return &models.SearchDatasetsResponse{
				Total: 1,
				Results: []models.SearchDatasetResult{
					{
						FullName:    "dex.trades",
						Category:    "spell",
						DatasetType: &dt,
						Blockchains: []string{"ethereum", "arbitrum"},
					},
				},
				Pagination: models.SearchDatasetsPagination{
					Limit:   5,
					Offset:  0,
					HasMore: false,
				},
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{
		"dataset", "search",
		"--query", "dex trades",
		"--categories", "spell",
		"--blockchains", "ethereum",
		"--limit", "5",
	})

	require.NoError(t, root.Execute())

	// Verify flags mapped to request
	assert.Equal(t, "dex trades", *gotReq.Query)
	assert.Equal(t, []string{"spell"}, gotReq.Categories)
	assert.Equal(t, []string{"ethereum"}, gotReq.Blockchains)
	assert.Equal(t, int32(5), *gotReq.Limit)

	// Verify table output
	out := buf.String()
	assert.Contains(t, out, "dex.trades")
	assert.Contains(t, out, "spell")
	assert.Contains(t, out, "ethereum, arbitrum")
	assert.Contains(t, out, "1 of 1 results")
}

func TestSearchJSON(t *testing.T) {
	mock := &mockClient{
		searchDatasetsFn: func(req models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
			return &models.SearchDatasetsResponse{
				Total: 1,
				Results: []models.SearchDatasetResult{
					{
						FullName: "ethereum.transactions",
						Category: "canonical",
					},
				},
				Pagination: models.SearchDatasetsPagination{
					Limit:   20,
					Offset:  0,
					HasMore: false,
				},
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"dataset", "search", "--query", "ethereum", "-o", "json"})

	require.NoError(t, root.Execute())

	var resp models.SearchDatasetsResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Equal(t, int32(1), resp.Total)
	assert.Equal(t, "ethereum.transactions", resp.Results[0].FullName)
}

func TestSearchError(t *testing.T) {
	mock := &mockClient{
		searchDatasetsFn: func(req models.SearchDatasetsRequest) (*models.SearchDatasetsResponse, error) {
			return nil, fmt.Errorf("API error: unauthorized")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"dataset", "search", "--query", "test"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error: unauthorized")
}
