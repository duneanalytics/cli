package dataset_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchByContractSuccess(t *testing.T) {
	dt := "decoded"
	var gotReq models.SearchDatasetsByContractAddressRequest

	mock := &mockClient{
		searchByContractAddressFn: func(req models.SearchDatasetsByContractAddressRequest) (*models.SearchDatasetsResponse, error) {
			gotReq = req
			return &models.SearchDatasetsResponse{
				Total: 2,
				Results: []models.SearchDatasetResult{
					{
						FullName:    "uniswap_v3_ethereum.UniswapV3Factory_evt_PoolCreated",
						Category:    "decoded",
						DatasetType: &dt,
						Blockchains: []string{"ethereum"},
					},
					{
						FullName:    "uniswap_v3_ethereum.UniswapV3Factory_call_createPool",
						Category:    "decoded",
						DatasetType: &dt,
						Blockchains: []string{"ethereum"},
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
	root.SetArgs([]string{
		"dataset", "search-by-contract",
		"--contract-address", "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
		"--blockchains", "ethereum",
		"--limit", "10",
	})

	require.NoError(t, root.Execute())

	// Verify flags mapped to request
	assert.Equal(t, "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", gotReq.ContractAddress)
	assert.Equal(t, []string{"ethereum"}, gotReq.Blockchains)
	assert.Equal(t, int32(10), *gotReq.Limit)

	// Verify table output
	out := buf.String()
	assert.Contains(t, out, "uniswap_v3_ethereum.UniswapV3Factory_evt_PoolCreated")
	assert.Contains(t, out, "uniswap_v3_ethereum.UniswapV3Factory_call_createPool")
	assert.Contains(t, out, "decoded")
	assert.Contains(t, out, "2 of 2 results")
}

func TestSearchByContractJSON(t *testing.T) {
	mock := &mockClient{
		searchByContractAddressFn: func(req models.SearchDatasetsByContractAddressRequest) (*models.SearchDatasetsResponse, error) {
			return &models.SearchDatasetsResponse{
				Total: 1,
				Results: []models.SearchDatasetResult{
					{
						FullName: "uniswap_v3_ethereum.UniswapV3Factory_evt_PoolCreated",
						Category: "decoded",
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
	root.SetArgs([]string{
		"dataset", "search-by-contract",
		"--contract-address", "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
		"-o", "json",
	})

	require.NoError(t, root.Execute())

	var resp models.SearchDatasetsResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Equal(t, int32(1), resp.Total)
	assert.Equal(t, "uniswap_v3_ethereum.UniswapV3Factory_evt_PoolCreated", resp.Results[0].FullName)
}

func TestSearchByContractError(t *testing.T) {
	mock := &mockClient{
		searchByContractAddressFn: func(req models.SearchDatasetsByContractAddressRequest) (*models.SearchDatasetsResponse, error) {
			return nil, fmt.Errorf("API error: unauthorized")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{
		"dataset", "search-by-contract",
		"--contract-address", "0xdeadbeef",
	})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error: unauthorized")
}

func TestSearchByContractRequiresAddress(t *testing.T) {
	mock := &mockClient{}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"dataset", "search-by-contract"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contract-address")
}
