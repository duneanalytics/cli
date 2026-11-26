package visualization

import (
	"errors"
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// ChartConfigHelp explains the --chart-type and --config flags, which describe a
// chart the way the Dune MCP's generateVisualization tool does.
const ChartConfigHelp = `CHART CONFIG (--chart-type and --config)
--chart-type is column, line, area, scatter, pie, counter, or table, and --config
describes the chart in terms of the query's result columns; config.chart_type must
equal --chart-type. Dune builds the full visualization options from it. Column names
must match the query output exactly; when the query has completed results, Dune
rejects a config that names other columns. Look at the columns first with
'dune execution results'. --config takes inline JSON, @path, or @-.

  column/line/area/scatter:
    {"chart_type":"line","x_axis":{"column":"day","title":"Day"},
     "y_axes":{"left":{"label_format":"0,0","series":[{"column":"volume","title":"Volume"}]},
               "right_enabled":false,"right_label_format":"","right_series":[]},
     "enable_legend":true,"stacked":false}
    Optional: x_axis.sort_x, y_axes.left.title/tick_format, group_by_column (one series
    per distinct value), when_duplicate_x (first, sum, all), normalize_to_percentage.
    Only line and area charts can use the right axis.
  pie:
    {"chart_type":"pie","x_axis_column":"token","y_axis_column":"volume",
     "y_axis_title":"Volume","y_axis_label_format":"$0.0a","show_labels":true}
  counter:
    {"chart_type":"counter","column":"total","row_num":1,"decimal_places":0,"prefix":"",
     "suffix":"","label":"Total","color_positive":false,"color_negative":false}
  table:
    {"chart_type":"table","columns":[{"name":"address","title":"Address",
     "column_type":"normal","data_type":"string","align":"left","is_hidden":false,
     "color_positive":false,"color_negative":false,"number_format":"0,0"}]}

Number formats (Numeral.js): 0, 0,0, 0,0.00, 0.0a, $0.0a, 0%.`

// AddChartFlags registers --chart-type and --config.
func AddChartFlags(cmd *cobra.Command) {
	cmd.Flags().String("chart-type", "", "chart type: column, line, area, scatter, pie, counter, or table")
	cmd.Flags().String("config", "", "chart config JSON, @path, or @- (see CHART CONFIG in --help)")
	cmd.MarkFlagsRequiredTogether("chart-type", "config")
}

// ChartArgs returns the chartType and config tool arguments, or nil when
// --chart-type and --config are not set.
func ChartArgs(cmd *cobra.Command) (map[string]any, error) {
	if !cmd.Flags().Changed("chart-type") {
		return nil, nil
	}
	chartType, _ := cmd.Flags().GetString("chart-type")
	var config map[string]any
	if err := cmdutil.JSONFlag(cmd, "config", &config); err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("--config must be a JSON object")
	}
	if ct, ok := config["chart_type"]; !ok {
		config["chart_type"] = chartType
	} else if ct != chartType {
		return nil, fmt.Errorf("the chart_type in --config, %q, does not match --chart-type %q", ct, chartType)
	}
	return map[string]any{"chartType": chartType, "config": config}, nil
}

// chartResult is what the Dune MCP's chart tools return about a visualization.
type chartResult struct {
	QueryID             int    `json:"queryId"`
	VisualizationID     int    `json:"visualizationId"`
	VisualizationName   string `json:"visualizationName"`
	InteractiveChartURL string `json:"interactiveChartUrl"`
	ChartJpegURL        string `json:"chartJpegUrl"`
}

// runChartTool calls generateVisualization or updateVisualization and prints
// "<verb> visualization <id> on query <id>" with the chart's URLs.
func runChartTool(cmd *cobra.Command, tool string, args map[string]any, verb string) error {
	var res chartResult
	raw, err := cmdutil.CallTool(cmd, tool, args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(cmd, raw)
	}
	fmt.Fprintf(w, "%s visualization %d on query %d\n", verb, res.VisualizationID, res.QueryID)
	fmt.Fprintf(w, "Chart: %s\n", res.InteractiveChartURL)
	if res.ChartJpegURL != "" {
		fmt.Fprintf(w, "Image: %s\n", res.ChartJpegURL)
	}
	return nil
}
