package dashboard

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// positionArgs maps the position flags to the fields of the tool's position.
var positionArgs = map[string]string{"row": "row", "col": "col", "size-x": "sizeX", "size-y": "sizeY"}

func newUpdateWidgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-widget <dashboard_id>",
		Short: "Move or resize one widget, or edit a text widget's markdown",
		Long: `Change one widget on a dashboard. Every other widget keeps its content and
position, so unlike 'dune dashboard update' there is no need to send them all.

Get the widget's ID and type from 'dune dashboard get <id> -o json':
  visualization  an entry in visualization_widgets (numeric widget_id)
  text           an entry in text_widgets (numeric widget_id)
  param          an entry in param_widgets (UUID widget_id)

Give a new position (all of --row, --col, --size-x and --size-y), new --text
for a text widget, or both. The grid has 6 columns; one row is 32 px and charts
are 8 rows high by default.

To add or remove widgets, use 'dune dashboard update'. To change a chart's
title, description, or counter label, use 'dune viz update'.

Examples:
  dune dashboard update-widget 12345 --widget-type visualization --widget-id 678 \
    --row 0 --col 0 --size-x 6 --size-y 8
  dune dashboard update-widget 12345 --widget-type text --widget-id 679 --text @intro.md`,
		Args: cobra.ExactArgs(1),
		RunE: runUpdateWidget,
	}

	cmd.Flags().String("widget-type", "", "visualization, text, or param (required)")
	cmd.Flags().String("widget-id", "", "the widget_id from 'dune dashboard get' (required)")
	cmd.Flags().Int("row", 0, "top edge, in grid rows from the top (0 is the top)")
	cmd.Flags().Int("col", 0, "left edge, in grid columns from the left (0 to 5)")
	cmd.Flags().Int("size-x", 0, "width in grid columns (1 to 6)")
	cmd.Flags().Int("size-y", 0, "height in grid rows")
	cmd.Flags().String("text", "", "new markdown for a text widget: inline, @path, or @-")
	_ = cmd.MarkFlagRequired("widget-type")
	_ = cmd.MarkFlagRequired("widget-id")
	cmd.MarkFlagsRequiredTogether("row", "col", "size-x", "size-y")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runUpdateWidget(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid dashboard ID %q: must be an integer", args[0])
	}
	widgetType, _ := cmd.Flags().GetString("widget-type")
	idFlag, _ := cmd.Flags().GetString("widget-id")

	var widgetID any = idFlag
	switch widgetType {
	case "visualization", "text":
		if widgetID, err = strconv.Atoi(idFlag); err != nil {
			return fmt.Errorf("invalid --widget-id %q: %s widgets have integer IDs", idFlag, widgetType)
		}
	case "param":
	default:
		return fmt.Errorf("invalid --widget-type %q: must be visualization, text, or param", widgetType)
	}

	toolArgs := map[string]any{"dashboardId": dashboardID, "widgetType": widgetType, "widgetId": widgetID}
	if cmd.Flags().Changed("row") {
		position := map[string]any{}
		for flag, field := range positionArgs {
			position[field], _ = cmd.Flags().GetInt(flag)
		}
		toolArgs["position"] = position
	}
	if cmd.Flags().Changed("text") {
		text, err := cmdutil.FileFlag(cmd, "text")
		if err != nil {
			return err
		}
		toolArgs["text"] = string(text)
	}
	if toolArgs["position"] == nil && toolArgs["text"] == nil {
		return errors.New("provide a new position (--row, --col, --size-x and --size-y), --text, or both")
	}

	var res struct {
		DashboardURL string `json:"dashboardUrl"`
	}
	raw, err := cmdutil.CallTool(cmd, "updateDashboardWidget", toolArgs, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	fmt.Fprintf(w, "Updated %s widget %s on dashboard %d\n%s\n", widgetType, idFlag, dashboardID, res.DashboardURL)
	return nil
}
