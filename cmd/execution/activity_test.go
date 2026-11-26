package execution_test

import (
	"os"
	"strings"
	"testing"

	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const activeRuns = `{
  "executions": [
    {"execution_id": "01KAQ5RUN", "state": "running", "source": "schedule", "query_id": 42, "query_name": "Daily DEX volume",
     "started_by": {"user_id": 7, "handle": "alice"}, "performance": "medium", "uses_concurrency_slot": true, "can_cancel": true},
    {"execution_id": "01KAQ5QUEUED", "state": "queued", "source": "api", "api_key_name": "prod ingest",
     "performance": "large", "uses_concurrency_slot": false, "can_cancel": false}
  ],
  "total_count": 3,
  "concurrency_slot_limit": 1
}`

func TestListActiveRuns(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listTeamExecutions": activeRuns}}
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "list"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"status":"active"}`, fake.Last().Args)
	out := buf.String()
	assert.Regexp(t, `01KAQ5RUN\s+running\s+Daily DEX volume\s+alice\s+schedule\s+medium\s+yes\s+yes`, out)
	assert.Regexp(t, `01KAQ5QUEUED\s+queued\s+ad-hoc SQL\s+API key prod ingest\s+api\s+large\s+no\s+no`, out)
	assert.Contains(t, out, "2 of 3 queued and running runs, 1 concurrency slots")
}

func TestListFinishedRuns(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listTeamExecutions": `{"executions": [
		{"execution_id": "01KAQ5DONE", "state": "timed_out", "source": "mcp", "query_id": 42,
		 "ended_at": "2026-11-24T10:00:00.000Z", "duration_seconds": 1800}]}`}}
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "list", "--status", "finished"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"status":"finished"}`, fake.Last().Args)
	assert.Regexp(t, `01KAQ5DONE\s+timed_out\s+42\s+-\s+mcp\s+1800.0s\s+2026-11-24T10:00:00.000Z`, buf.String())
}

func TestListRejectsUnknownStatus(t *testing.T) {
	fake := &dunemcptest.Fake{}
	root, _ := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "list", "--status", "running"})
	require.Error(t, root.Execute())
	assert.Empty(t, fake.Calls)
}

func TestListJSONPassesTheResultThrough(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listTeamExecutions": activeRuns}}
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "list", "-o", "json"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, activeRuns, buf.String())
}

func cancelFake(result string) *dunemcptest.Fake {
	return &dunemcptest.Fake{Results: map[string]string{"cancelTeamExecution": result}}
}

func TestCancelAsksFirstAndCancelsOnYes(t *testing.T) {
	fake := cancelFake(`{"execution_id": "01KAQ5RUN", "cancelled": true}`)
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetIn(strings.NewReader("y\n"))
	root.SetArgs([]string{"execution", "cancel", "01KAQ5RUN"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"execution_id":"01KAQ5RUN"}`, fake.Last().Args)
	assert.Contains(t, buf.String(), "Cancelled run 01KAQ5RUN")
}

func TestCancelDoesNothingWhenTheAnswerIsNo(t *testing.T) {
	fake := cancelFake(`{"execution_id": "01KAQ5RUN", "cancelled": true}`)
	root, _ := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"execution", "cancel", "01KAQ5RUN"})
	require.NoError(t, root.Execute())
	assert.Empty(t, fake.Calls)
}

func TestCancelRefusesWithoutYesWhenNoOneCanAnswer(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer devNull.Close()
	fake := cancelFake(`{"execution_id": "01KAQ5RUN", "cancelled": true}`)
	root, _ := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetIn(devNull)
	root.SetArgs([]string{"execution", "cancel", "01KAQ5RUN"})

	err = root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--yes")
	assert.Empty(t, fake.Calls)
}

func TestCancelWithYesSkipsTheQuestion(t *testing.T) {
	fake := cancelFake(`{"execution_id": "01KAQ5RUN", "cancelled": true}`)
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "cancel", "01KAQ5RUN", "--yes", "-o", "json"})
	require.NoError(t, root.Execute())
	assert.JSONEq(t, `{"execution_id": "01KAQ5RUN", "cancelled": true}`, buf.String())
}

func TestCancelFailsWhenTheRunWasNotCancelled(t *testing.T) {
	fake := cancelFake(`{"execution_id": "01KAQ5RUN", "cancelled": false}`)
	root, _ := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs([]string{"execution", "cancel", "01KAQ5RUN", "-y"})
	require.Error(t, root.Execute())
}
