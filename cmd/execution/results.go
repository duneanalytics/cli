package execution

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

// PollInterval controls the polling interval when waiting for execution results.
var PollInterval = 2 * time.Second

func newResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <execution-id>",
		Short: "Retrieve execution results for a query execution by execution ID",
		Long: "Retrieve the results of a query execution. By default, waits for the execution\n" +
			"to complete (up to the timeout) before returning results.\n\n" +
			"Behavior:\n" +
			"  1. Checks the current execution status\n" +
			"  2. If still running: polls every 2 seconds until complete or timeout is reached\n" +
			"  3. If completed: returns the result data\n" +
			"  4. If failed/cancelled: returns the error details\n\n" +
			"Use --no-wait to return the current state immediately without polling.\n" +
			"Use --limit and --offset to paginate through large result sets.\n\n" +
			"Shape the rows on the server instead of downloading them all:\n" +
			"  --columns   return only these columns, in this order\n" +
			"  --filter    keep rows matching a WHERE-like expression: <column> <operator> <literal>,\n" +
			"              combined with AND/OR and parentheses; operators =, <, <=, >, >=, LIKE,\n" +
			"              IN (...), IS NOT NULL; dates as strings, e.g. block_time > '2026-03-01'\n" +
			"  --sort-by   ORDER BY fragments in priority order, e.g. '\"amount\" DESC' (repeatable)\n" +
			"  --sample    a random sample of about N rows; the count varies from call to call, and\n" +
			"              --sample cannot be combined with --limit or --offset\n" +
			"Use this after submitting a query with --no-wait, or to re-fetch results from\n" +
			"a completed execution.\n\n" +
			"Examples:\n" +
			"  dune execution results 01JXYZ...\n" +
			"  dune execution results 01JXYZ... --limit 100 --offset 200\n" +
			"  dune execution results 01JXYZ... --no-wait\n" +
			"  dune execution results 01JXYZ... --output json\n" +
			"  dune execution results 01JXYZ... --columns block_time,amount --filter '\"amount\" > 0' \\\n" +
			"    --sort-by '\"amount\" DESC' --limit 10\n" +
			"  dune execution results 01JXYZ... --sample 1000 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runResults,
	}

	cmd.Flags().Int("limit", 0, "maximum number of result rows to return (0 = all); use with --offset to paginate large result sets")
	cmd.Flags().Int("offset", 0, "number of rows to skip before starting to return results; use with --limit for pagination through large result sets")
	cmd.Flags().Bool("no-wait", false, "return the current execution state immediately without waiting for completion; useful for checking status of long-running queries")
	cmd.Flags().Int("timeout", 300, "maximum seconds to wait for the execution to complete before timing out (default 300s = 5 minutes)")
	cmd.Flags().StringSlice("columns", nil, "return only these columns, in this order (comma-separated)")
	cmd.Flags().String("filter", "", `keep only rows matching this WHERE-like expression, e.g. '"amount" > 0'`)
	cmd.Flags().StringArray("sort-by", nil, `sort by this ORDER BY fragment, e.g. '"block_time" DESC' (repeatable)`)
	cmd.Flags().Int("sample", 0, "return a random sample of about this many rows; the count varies")
	cmd.MarkFlagsMutuallyExclusive("sample", "limit")
	cmd.MarkFlagsMutuallyExclusive("sample", "offset")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runResults(cmd *cobra.Command, args []string) error {
	executionID := args[0]

	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	noWait, _ := cmd.Flags().GetBool("no-wait")

	if limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative, got %d", offset)
	}

	opts := models.ResultOptions{}
	if limit > 0 || offset > 0 {
		opts.Page = &models.ResultPageOption{
			Offset: uint64(offset),
			Limit:  uint32(limit),
		}
	}
	fs := cmd.Flags()
	if fs.Changed("columns") {
		opts.Columns, _ = fs.GetStringSlice("columns")
	}
	opts.Filters, _ = fs.GetString("filter")
	if fs.Changed("sort-by") {
		opts.SortBy, _ = fs.GetStringArray("sort-by")
	}
	if fs.Changed("sample") {
		sample, _ := fs.GetInt("sample")
		if sample <= 0 {
			return fmt.Errorf("sample must be positive, got %d", sample)
		}
		opts.SampleCount = uint32(sample)
	}

	client := cmdutil.ClientFromCmd(cmd)

	if !noWait {
		timeout, _ := cmd.Flags().GetInt("timeout")
		if err := WaitForExecution(func() (*models.StatusResponse, error) { return client.QueryStatus(executionID) },
			cmd.ErrOrStderr(), executionID, timeout); err != nil {
			return err
		}
	}

	// Fetch final results with any paging and shaping options.
	resp, err := client.QueryResultsV2(executionID, opts)
	if err != nil {
		return err
	}
	return handleResultsResponse(cmd, executionID, resp)
}

func handleResultsResponse(cmd *cobra.Command, executionID string, resp *models.ResultsResponse) error {
	switch resp.State {
	case "QUERY_STATE_COMPLETED":
		return output.DisplayResults(cmd, resp)
	case "QUERY_STATE_PENDING", "QUERY_STATE_EXECUTING":
		w := cmd.OutOrStdout()
		switch output.FormatFromCmd(cmd) {
		case output.FormatJSON:
			return output.PrintJSON(cmd, resp)
		default:
			fmt.Fprintf(w, "Execution ID: %s\n", executionID)
			fmt.Fprintf(w, "State:        %s\n", resp.State)
			return nil
		}
	case "QUERY_STATE_FAILED":
		msg := "execution failed"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return fmt.Errorf("%s", msg)
	case "QUERY_STATE_CANCELLED":
		return fmt.Errorf("execution was cancelled")
	default:
		return fmt.Errorf("unexpected execution state: %s", resp.State)
	}
}

// WaitForExecution polls an execution's status with status, without downloading
// rows, until it stops running, and fails when it is still running at the timeout.
// It retries a failed status call that may succeed later, with a note on stderr,
// until the timeout; a client error such as an unknown execution ID or a bad API
// key fails at once.
func WaitForExecution(
	status func() (*models.StatusResponse, error), stderr io.Writer, executionID string, timeout int,
) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		st, err := status()
		if err == nil && st.State != "QUERY_STATE_PENDING" && st.State != "QUERY_STATE_EXECUTING" {
			return nil
		}
		if err != nil && !retryable(err) {
			return err
		}
		if !time.Now().Add(PollInterval).Before(deadline) {
			if err != nil {
				return fmt.Errorf("%w: %w", dune.ErrorRetriesExhausted, err)
			}
			return fmt.Errorf("execution %s is still %s after %ds; rerun with a longer --timeout or --no-wait",
				executionID, strings.TrimPrefix(st.State, "QUERY_STATE_"), timeout)
		}
		if err != nil {
			fmt.Fprintf(stderr, "checking the execution status failed, retrying: %v\n", err)
		}
		time.Sleep(PollInterval)
	}
}

// retryable reports whether a failed API call may succeed if repeated: every
// failure except a 4xx response other than 429 Too Many Requests.
func retryable(err error) bool {
	var apiErr *dune.APIError
	if !errors.As(err, &apiErr) {
		return true
	}
	return apiErr.StatusCode < 400 || apiErr.StatusCode >= 500 || apiErr.StatusCode == http.StatusTooManyRequests
}
