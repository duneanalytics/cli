package dataset

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// tableColumns returns a table's columns as "name type" strings. It reads the
// Dune MCP's columns list ([{"name","type",...}]) and falls back to the API's
// --include-schema field ({"fields":[{"name","type",...}]}).
func tableColumns(columns, schema json.RawMessage) []string {
	type column struct{ Name, Type string }
	var cols []column
	if len(columns) > 0 {
		_ = json.Unmarshal(columns, &cols)
	}
	if len(cols) == 0 && len(schema) > 0 {
		var s struct{ Fields []column }
		_ = json.Unmarshal(schema, &s)
		cols = s.Fields
	}
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name + " " + c.Type
	}
	return out
}

// printColumns lists the columns of each table that has them, one table per
// line, so text output carries what an agent needs to write SQL.
func printColumns(w io.Writer, names []string, columns [][]string) {
	header := false
	for i, cols := range columns {
		if len(cols) == 0 {
			continue
		}
		if !header {
			fmt.Fprintln(w, "\nCOLUMNS")
			header = true
		}
		fmt.Fprintf(w, "%s: %s\n", names[i], strings.Join(cols, ", "))
	}
}
