package visualization_test

import (
	"bytes"
	"testing"

	"github.com/duneanalytics/cli/cmd/visualization"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	root, buf := dunemcptest.NewRoot(visualization.NewVisualizationCmd(), fake)
	root.SetArgs(append([]string{"viz"}, args...))
	return buf, root.Execute()
}

const chartResult = `{"queryId":7,"visualizationId":9,"chartType":"counter","visualizationName":"Total",
	"interactiveChartUrl":"https://dune.com/embeds/7/9","chartJpegUrl":"https://img"}`

func TestCreateFromChartConfig(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"generateVisualization": chartResult}}
	buf, err := run(fake, "create", "--query-id", "7", "--name", "Total", "--chart-type", "counter",
		"--config", `{"column":"n","row_num":1}`)
	require.NoError(t, err)

	// config.chart_type is filled in from --chart-type.
	assert.JSONEq(t, `{"queryId":7,"visualizationName":"Total","chartType":"counter",
		"config":{"chart_type":"counter","column":"n","row_num":1}}`, fake.Last().Args)
	assert.Equal(t, "Created visualization 9 on query 7\nChart: https://dune.com/embeds/7/9\nImage: https://img\n",
		buf.String())
}

func TestChartTypeMismatch(t *testing.T) {
	_, err := run(&dunemcptest.Fake{}, "create", "--query-id", "7", "--name", "T", "--chart-type", "line",
		"--config", `{"chart_type":"pie"}`)
	assert.EqualError(t, err, `the chart_type in --config, "pie", does not match --chart-type "line"`)
}

func TestCreateNeedsOptionsOrConfig(t *testing.T) {
	_, err := run(&dunemcptest.Fake{}, "create", "--query-id", "7", "--name", "T")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[options config]")

	_, err = run(&dunemcptest.Fake{}, "create", "--query-id", "7", "--name", "T", "--config", "{}")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chart-type")
}

func TestUpdateFromChartConfigKeepsName(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateVisualization": chartResult}}
	_, err := run(fake, "update", "9", "--chart-type", "counter", "--config", `{"column":"n"}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"visualizationId":9,"chartType":"counter","config":{"chart_type":"counter","column":"n"}}`,
		fake.Last().Args)
}
