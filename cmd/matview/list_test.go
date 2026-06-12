package matview_test

import (
	"testing"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSuccess(t *testing.T) {
	var gotLimit, gotOffset int
	mock := &mockClient{
		listFn: func(limit, offset int) (*models.ListMaterializedViewsResponse, error) {
			gotLimit, gotOffset = limit, offset
			return &models.ListMaterializedViewsResponse{
				MaterializedViews: []*models.MaterializedViewListElement{
					{SQLID: "dune.t.result_a", QueryID: 1, IsPrivate: true, TableSizeBytes: 1024},
					{SQLID: "dune.t.result_b", QueryID: 2, IsPrivate: false, TableSizeBytes: 0},
				},
			}, nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "list", "--limit", "50", "--offset", "10"})
	require.NoError(t, root.Execute())

	assert.Equal(t, 50, gotLimit)
	assert.Equal(t, 10, gotOffset)
	out := buf.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "dune.t.result_a")
	assert.Contains(t, out, "dune.t.result_b")
	assert.Contains(t, out, "yes")
	assert.Contains(t, out, "no")
}

func TestListEmpty(t *testing.T) {
	mock := &mockClient{
		listFn: func(int, int) (*models.ListMaterializedViewsResponse, error) {
			return &models.ListMaterializedViewsResponse{}, nil
		},
	}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "list"})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "No materialized views found.")
}

func TestListAllPaginates(t *testing.T) {
	var calls []int // offsets requested
	mock := &mockClient{
		listFn: func(limit, offset int) (*models.ListMaterializedViewsResponse, error) {
			calls = append(calls, offset)
			switch offset {
			case 0:
				return &models.ListMaterializedViewsResponse{
					MaterializedViews: []*models.MaterializedViewListElement{{SQLID: "dune.t.result_a", QueryID: 1}},
					NextOffset:        1,
				}, nil
			default:
				return &models.ListMaterializedViewsResponse{
					MaterializedViews: []*models.MaterializedViewListElement{{SQLID: "dune.t.result_b", QueryID: 2}},
					NextOffset:        0,
				}, nil
			}
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "list", "--all"})
	require.NoError(t, root.Execute())

	assert.Equal(t, []int{0, 1}, calls, "should page until next_offset is 0")
	out := buf.String()
	assert.Contains(t, out, "dune.t.result_a")
	assert.Contains(t, out, "dune.t.result_b")
	assert.NotContains(t, out, "More results available", "no pagination hint when --all")
}

func TestListShowsNextOffsetHint(t *testing.T) {
	mock := &mockClient{
		listFn: func(int, int) (*models.ListMaterializedViewsResponse, error) {
			return &models.ListMaterializedViewsResponse{
				MaterializedViews: []*models.MaterializedViewListElement{{SQLID: "dune.t.result_a", QueryID: 1}},
				NextOffset:        100,
			}, nil
		},
	}
	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"matview", "list"})
	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "--offset 100")
}
