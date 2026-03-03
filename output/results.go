package output

import (
	"fmt"

	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// DisplayResults renders a ResultsResponse as JSON or table based on cmd flags.
func DisplayResults(cmd *cobra.Command, resp *models.ResultsResponse) error {
	w := cmd.OutOrStdout()

	if FormatFromCmd(cmd) == FormatJSON {
		return PrintJSON(w, resp)
	}

	limit, _ := cmd.Flags().GetInt("limit")
	columns := resp.Result.Metadata.ColumnNames
	sourceRows := resp.Result.Rows
	totalRows := len(sourceRows)

	if limit > 0 && limit < totalRows {
		sourceRows = sourceRows[:limit]
	}
	rows := ResultRowsToStrings(sourceRows, columns)

	PrintTable(w, columns, rows)
	if limit > 0 && limit < totalRows {
		fmt.Fprintf(w, "\nShowing %d of %d rows\n", limit, totalRows)
	} else {
		fmt.Fprintf(w, "\n%d rows\n", totalRows)
	}
	return nil
}

// ResultRowsToStrings converts result rows to a string grid using column ordering.
func ResultRowsToStrings(rows []map[string]any, columns []string) [][]string {
	out := make([][]string, len(rows))
	for i, row := range rows {
		cells := make([]string, len(columns))
		for j, col := range columns {
			cells[j] = fmt.Sprintf("%v", row[col])
		}
		out[i] = cells
	}
	return out
}
