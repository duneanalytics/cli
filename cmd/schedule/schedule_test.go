package schedule_test

import (
	"bytes"
	"testing"

	"github.com/duneanalytics/cli/cmd/schedule"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	root, buf := dunemcptest.NewRoot(schedule.NewScheduleCmd(), fake)
	root.SetArgs(append([]string{"schedule"}, args...))
	return buf, root.Execute()
}

func TestSet(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"setSchedule": `{"query_id":5,"cron_expression":"0 6 * * *",
		"performance":"small","next_execution_time":"2026-11-20T06:00:00Z","created":true}`}}
	buf, err := run(fake, "set", "--query-id", "5", "--cron", "0 6 * * *", "--performance", "small")
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_id":5,"cron_expression":"0 6 * * *","performance":"small"}`, fake.Last().Args)
	assert.Contains(t, buf.String(), `Created schedule for query 5: "0 6 * * *" on small`)
	assert.Contains(t, buf.String(), "Expires at: -")
}

func TestSetNeedsExactlyOneTarget(t *testing.T) {
	_, err := run(&dunemcptest.Fake{}, "set", "--cron", "0 6 * * *", "--performance", "small")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[query-id dashboard-id]")

	_, err = run(&dunemcptest.Fake{},
		"set", "--query-id", "1", "--dashboard-id", "2", "--cron", "x", "--performance", "small")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestDeleteWithoutSchedule(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"deleteSchedule": `{"deleted":false}`}}
	buf, err := run(fake, "delete", "--dashboard-id", "9")
	require.NoError(t, err)
	assert.JSONEq(t, `{"dashboard_id":9}`, fake.Last().Args)
	assert.Equal(t, "There was no schedule to delete\n", buf.String())
}

func TestList(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"listSchedules": `{"schedules":[
		{"type":"dashboard","dashboard_id":9,"cron_expression":"0 */6 * * *","performance":"medium"}],"total":40}`}}
	buf, err := run(fake, "list", "--limit", "1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"limit":1}`, fake.Last().Args)
	assert.Regexp(t, `dashboard\s+dashboard 9\s+0 \*/6 \* \* \*\s+medium\s+-\s+-`, buf.String())
	assert.Contains(t, buf.String(), "1 of 40 schedules")
}
