package visualization_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/duneanalytics/cli/cmd/visualization"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restClient embeds the interface so unimplemented methods panic.
type restClient struct {
	dune.DuneClient
	viz *models.GetVisualizationResponse
}

func (c *restClient) GetVisualization(int) (*models.GetVisualizationResponse, error) {
	return c.viz, nil
}

const vizResult = `{"id":9,"queryId":7,"type":"chart","chartType":"line","name":"Daily","options":{},
	"createdAt":"2026-01-01T00:00:00.000Z","updatedAt":"2026-01-02T00:00:00.000Z","isPrivate":true,
	"interactiveChartUrl":"https://dune.com/embeds/7/9/KEY",
	"chartJpegUrl":"https://core-api.dune.com/public/screenshot?url=x&cache=0"}`

func runGet(args ...string) (*bytes.Buffer, *dunemcptest.Fake, error) {
	fake := &dunemcptest.Fake{Results: map[string]string{"getVisualization": vizResult}}
	rest := &restClient{viz: &models.GetVisualizationResponse{
		ID: 9, QueryID: 7, Name: "Daily", Type: "chart", Options: map[string]any{"globalSeriesType": "line"},
		CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-02T00:00:00Z",
	}}
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(c *cobra.Command, _ []string) {
			cmdutil.SetClient(c, rest)
			cmdutil.SetMCP(c, fake)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(visualization.NewVisualizationCmd())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs(append([]string{"viz", "get"}, args...))
	return &buf, fake, root.Execute()
}

func TestGetShowsChartTypeAndLinks(t *testing.T) {
	buf, fake, err := runGet("9")
	require.NoError(t, err)

	assert.JSONEq(t, `{"visualizationId":9}`, fake.Last().Args)
	out := buf.String()
	assert.Regexp(t, `Chart Type\s+line`, out)
	assert.Regexp(t, `Private\s+true`, out)
	assert.Regexp(t, `Chart URL\s+https://dune.com/embeds/7/9/KEY`, out)
	assert.Regexp(t, `Image URL\s+https://core-api.dune.com/public/screenshot`, out)
}

func TestGetJSONKeepsTheRESTFields(t *testing.T) {
	buf, _, err := runGet("9", "-o", "json")
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	// The REST shape, including its raw options, is unchanged.
	assert.Equal(t, float64(7), got["query_id"])
	assert.Equal(t, "2026-01-01T00:00:00Z", got["created_at"])
	assert.Equal(t, map[string]any{"globalSeriesType": "line"}, got["options"])
	// The MCP fields are added in the same snake_case style.
	assert.Equal(t, "line", got["chart_type"])
	assert.Equal(t, true, got["is_private"])
	assert.Equal(t, "https://dune.com/embeds/7/9/KEY", got["interactive_chart_url"])
}
