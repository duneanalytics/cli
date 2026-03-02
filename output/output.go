package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	FormatText = "text"
	FormatJSON = "json"
	FormatCSV  = "csv"
)

// AddFormatFlag registers the -o/--output flag on cmd with the given default.
func AddFormatFlag(cmd *cobra.Command, defaultFormat string) {
	cmd.Flags().StringP("output", "o", defaultFormat, `output format: "text", "json", or "csv"`)
}

// FormatFromCmd reads the output flag value from cmd.
func FormatFromCmd(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	return f
}

// PrintJSON encodes v as indented JSON and writes it to w.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintTable writes columns and rows as aligned text using tabwriter.
func PrintTable(w io.Writer, columns []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, col := range columns {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, col)
	}
	fmt.Fprintln(tw)

	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

// PrintCSV writes columns and rows as CSV to w.
func PrintCSV(w io.Writer, columns []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(columns); err != nil {
		return err
	}
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
