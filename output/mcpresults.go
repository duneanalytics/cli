package output

import (
	"fmt"
	"io"
	"strings"
)

// ExecutionResult is a query execution result as the Dune MCP's
// getExecutionResults tool returns it.
type ExecutionResult struct {
	ExecutionID    string `json:"executionId"`
	State          string `json:"state"`
	ErrorMessage   string `json:"errorMessage"`
	ResultMetadata *struct {
		Columns []struct {
			Name string `json:"name"`
		} `json:"columns"`
		TotalRowCount        int    `json:"totalRowCount"`
		ExecutionCostCredits string `json:"executionCostCredits"`
	} `json:"resultMetadata"`
	Data *struct {
		Rows []map[string]any `json:"rows"`
	} `json:"data"`
	Notice string `json:"notice"`
}

// Err returns an error when the execution failed, was cancelled, or expired.
func (r ExecutionResult) Err() error {
	switch r.State {
	case "FAILED", "CANCELLED", "EXPIRED":
		msg := r.ErrorMessage
		if msg == "" {
			msg = "no details"
		}
		return fmt.Errorf("execution %s %s: %s", r.ExecutionID, strings.ToLower(r.State), msg)
	}
	return nil
}

// PrintExecutionResult writes the rows of a completed execution as a table, or
// the execution's state while it is still running.
func PrintExecutionResult(w io.Writer, r ExecutionResult) {
	if r.State != "COMPLETED" {
		fmt.Fprintf(w, "Execution ID: %s\n", r.ExecutionID)
		fmt.Fprintf(w, "State:        %s\n", r.State)
		return
	}
	var columns []string
	total := 0
	credits := ""
	if m := r.ResultMetadata; m != nil {
		for _, c := range m.Columns {
			columns = append(columns, c.Name)
		}
		total = m.TotalRowCount
		credits = m.ExecutionCostCredits
	}
	var rows []map[string]any
	if r.Data != nil {
		rows = r.Data.Rows
	}
	PrintTable(w, columns, ResultRowsToStrings(rows, columns))
	fmt.Fprintf(w, "\n%d of %d rows\n", len(rows), total)
	if credits != "" {
		fmt.Fprintf(w, "Cost: %s credits\n", credits)
	}
	if r.Notice != "" {
		fmt.Fprintf(w, "\n%s\n", r.Notice)
	}
}
