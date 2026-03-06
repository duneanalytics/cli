package usage_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/duneanalytics/cli/cmd/usage"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	dune.DuneClient
	getUsageFn         func() (*models.UsageResponse, error)
	getUsageForDatesFn func(string, string) (*models.UsageResponse, error)
}

func (m *mockClient) GetUsage() (*models.UsageResponse, error) {
	return m.getUsageFn()
}

func (m *mockClient) GetUsageForDates(start, end string) (*models.UsageResponse, error) {
	return m.getUsageForDatesFn(start, end)
}

func newTestRoot(mock dune.DuneClient) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(usage.NewUsageCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)

	return root, &buf
}

func sampleUsageResponse() *models.UsageResponse {
	return &models.UsageResponse{
		PrivateQueries:    5,
		PrivateDashboards: 2,
		BytesUsed:         1288490188, // ~1.2 GB
		BytesAllowed:      10737418240,
		BillingPeriods: []models.BillingPeriod{
			{
				StartDate:       "2025-03-01",
				EndDate:         "2025-04-01",
				CreditsUsed:     450,
				CreditsIncluded: 1000,
			},
		},
	}
}

func TestUsageTextOutput(t *testing.T) {
	mock := &mockClient{
		getUsageFn: func() (*models.UsageResponse, error) {
			return sampleUsageResponse(), nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"usage"})
	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Private Queries:      5")
	assert.Contains(t, out, "Private Dashboards:   2")
	assert.Contains(t, out, "1.2 GB")
	assert.Contains(t, out, "10.0 GB")
	assert.Contains(t, out, "2025-03-01")
	assert.Contains(t, out, "450.00")
	assert.Contains(t, out, "1000")
}

func TestUsageJSONOutput(t *testing.T) {
	mock := &mockClient{
		getUsageFn: func() (*models.UsageResponse, error) {
			return sampleUsageResponse(), nil
		},
	}

	root, buf := newTestRoot(mock)
	root.SetArgs([]string{"usage", "-o", "json"})
	require.NoError(t, root.Execute())

	var got models.UsageResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 5, got.PrivateQueries)
	assert.Equal(t, 2, got.PrivateDashboards)
	assert.Len(t, got.BillingPeriods, 1)
	assert.Equal(t, float64(450), got.BillingPeriods[0].CreditsUsed)
}

func TestUsageAPIError(t *testing.T) {
	mock := &mockClient{
		getUsageFn: func() (*models.UsageResponse, error) {
			return nil, errors.New("api: unauthorized")
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"usage"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api: unauthorized")
}

func TestUsageWithDateFlags(t *testing.T) {
	var gotStart, gotEnd string
	mock := &mockClient{
		getUsageForDatesFn: func(start, end string) (*models.UsageResponse, error) {
			gotStart = start
			gotEnd = end
			return sampleUsageResponse(), nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"usage", "--start-date", "2025-01-01", "--end-date", "2025-02-01"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "2025-01-01", gotStart)
	assert.Equal(t, "2025-02-01", gotEnd)
}

func TestUsageWithOnlyStartDate(t *testing.T) {
	var called bool
	mock := &mockClient{
		getUsageForDatesFn: func(start, end string) (*models.UsageResponse, error) {
			called = true
			assert.Equal(t, "2025-01-01", start)
			assert.Equal(t, "", end)
			return sampleUsageResponse(), nil
		},
	}

	root, _ := newTestRoot(mock)
	root.SetArgs([]string{"usage", "--start-date", "2025-01-01"})
	require.NoError(t, root.Execute())
	assert.True(t, called)
}

func TestUsageInvalidDateFormat(t *testing.T) {
	root, _ := newTestRoot(&mockClient{})
	root.SetArgs([]string{"usage", "--start-date", "not-a-date"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --start-date")
}
