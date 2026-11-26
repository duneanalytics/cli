package output

import (
	"bytes"
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

// AddFormatFlag registers the -o/--output and --compact flags on cmd, with the
// given default format.
func AddFormatFlag(cmd *cobra.Command, defaultFormat string) {
	cmd.Flags().StringP("output", "o", defaultFormat, `output format: "text" or "json"`)
	cmd.Flags().Bool("compact", false,
		"with -o json, print the JSON on one line without indentation (fewer tokens for agents and scripts)")
}

// FormatFromCmd reads the output flag value from cmd.
func FormatFromCmd(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	return f
}

// jsonIndent is the indentation PrintJSON and PrintRawJSON use: none with
// --compact, two spaces otherwise.
func jsonIndent(cmd *cobra.Command) string {
	if compact, _ := cmd.Flags().GetBool("compact"); compact {
		return ""
	}
	return "  "
}

// PrintJSON encodes v as JSON, indented unless --compact is set, and writes it
// to cmd's output.
func PrintJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", jsonIndent(cmd))
	// Keep URLs and SQL readable: write & < > as they are, not as \u0026 \u003c \u003e.
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// PrintRawJSON writes JSON that is already encoded, such as a Dune MCP tool
// result, formatted the same way as PrintJSON.
func PrintRawJSON(cmd *cobra.Command, raw json.RawMessage) error {
	var buf bytes.Buffer
	var err error
	if indent := jsonIndent(cmd); indent == "" {
		err = json.Compact(&buf, raw)
	} else {
		err = json.Indent(&buf, raw, "", indent)
	}
	if err != nil {
		return err
	}
	buf.WriteByte('\n')
	_, err = buf.WriteTo(cmd.OutOrStdout())
	return err
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

// FormatUSD formats a USD value for display with 2 decimal places.
func FormatUSD(v float64) string {
	if v == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", v)
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
