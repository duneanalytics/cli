package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

// AddFormatFlag registers the -o/--output flag on cmd with the given default.
func AddFormatFlag(cmd *cobra.Command, defaultFormat string) {
	cmd.Flags().StringP("output", "o", defaultFormat, `output format: "text" or "json"`)
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

// FormatBytes returns a human-readable byte size string (e.g. "1.2 GB").
func FormatBytes(b int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
		tb = gb * 1024
	)
	switch {
	case b >= tb:
		return fmt.Sprintf("%.1f TB", float64(b)/float64(tb))
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
