package schedule_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/duneanalytics/cli/cmd/schedule"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient embeds the interface so unimplemented methods panic.
type mockClient struct {
	dune.DuneClient
	listFn            func(models.ListSchedulesOptions) (*models.ListSchedulesResponse, error)
	setQueryFn        func(int, models.SetScheduleRequest) (*models.SetScheduleResponse, error)
	setDashboardFn    func(int, models.SetScheduleRequest) (*models.SetScheduleResponse, error)
	deleteQueryFn     func(int) (*models.DeleteScheduleResponse, error)
	deleteDashboardFn func(int) (*models.DeleteScheduleResponse, error)
}

func (m *mockClient) ListSchedules(opts models.ListSchedulesOptions) (*models.ListSchedulesResponse, error) {
	return m.listFn(opts)
}

func (m *mockClient) SetQuerySchedule(id int, req models.SetScheduleRequest) (*models.SetScheduleResponse, error) {
	return m.setQueryFn(id, req)
}

func (m *mockClient) SetDashboardSchedule(
	id int, req models.SetScheduleRequest,
) (*models.SetScheduleResponse, error) {
	return m.setDashboardFn(id, req)
}

func (m *mockClient) DeleteQuerySchedule(id int) (*models.DeleteScheduleResponse, error) {
	return m.deleteQueryFn(id)
}

func (m *mockClient) DeleteDashboardSchedule(id int) (*models.DeleteScheduleResponse, error) {
	return m.deleteDashboardFn(id)
}

func run(mock dune.DuneClient, args ...string) (*bytes.Buffer, error) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			cmdutil.SetClient(cmd, mock)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(schedule.NewScheduleCmd())
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"schedule"}, args...))
	return &buf, root.Execute()
}

func ptr[T any](v T) *T { return &v }

func TestSetQuery(t *testing.T) {
	var gotID int
	var gotReq models.SetScheduleRequest
	next := time.Date(2026, 11, 20, 6, 0, 0, 0, time.UTC)
	mock := &mockClient{setQueryFn: func(id int, req models.SetScheduleRequest) (*models.SetScheduleResponse, error) {
		gotID, gotReq = id, req
		return &models.SetScheduleResponse{Created: true, Schedule: models.Schedule{
			Type: "query", QueryID: ptr(int64(5)), CronExpression: "0 6 * * *", Performance: "small",
			NextExecutionTime: &next,
		}}, nil
	}}
	buf, err := run(mock, "set", "--query-id", "5", "--cron", "0 6 * * *", "--performance", "small")
	require.NoError(t, err)
	assert.Equal(t, 5, gotID)
	assert.Equal(t, models.SetScheduleRequest{CronExpression: "0 6 * * *", Performance: "small"}, gotReq)
	assert.Nil(t, gotReq.ExpiresAt)
	assert.Contains(t, buf.String(), `Created schedule for query 5: "0 6 * * *" on small`)
	assert.Contains(t, buf.String(), "Next run:   2026-11-20T06:00:00Z")
	assert.Contains(t, buf.String(), "Expires at: -")
}

func TestSetDashboardWithExpiry(t *testing.T) {
	var gotID int
	var gotReq models.SetScheduleRequest
	mock := &mockClient{setDashboardFn: func(id int, req models.SetScheduleRequest) (*models.SetScheduleResponse, error) {
		gotID, gotReq = id, req
		return &models.SetScheduleResponse{Schedule: models.Schedule{
			Type: "dashboard", DashboardID: ptr(int64(678)), CronExpression: req.CronExpression,
			Performance: req.Performance, ExpiresAt: req.ExpiresAt,
		}}, nil
	}}
	buf, err := run(mock, "set", "--dashboard-id", "678", "--cron", "0 */6 * * *", "--performance", "medium",
		"--expires-at", "2027-01-01T02:00:00+02:00")
	require.NoError(t, err)
	assert.Equal(t, 678, gotID)
	require.NotNil(t, gotReq.ExpiresAt)
	assert.True(t, gotReq.ExpiresAt.Equal(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)), gotReq.ExpiresAt)
	assert.Contains(t, buf.String(), `Replaced schedule for dashboard 678: "0 */6 * * *" on medium`)
	assert.Contains(t, buf.String(), "Expires at: 2027-01-01T02:00:00+02:00")
}

func TestSetRejectsBadExpiry(t *testing.T) {
	_, err := run(&mockClient{}, "set", "--query-id", "5", "--cron", "0 6 * * *", "--performance", "small",
		"--expires-at", "tomorrow")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --expires-at")
}

func TestSetNeedsExactlyOneTarget(t *testing.T) {
	_, err := run(&mockClient{}, "set", "--cron", "0 6 * * *", "--performance", "small")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[query-id dashboard-id]")

	_, err = run(&mockClient{},
		"set", "--query-id", "1", "--dashboard-id", "2", "--cron", "x", "--performance", "small")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestSetReportsAPIError(t *testing.T) {
	apiErr := &dune.APIError{StatusCode: 400, Message: "unsupported cron"}
	mock := &mockClient{setQueryFn: func(int, models.SetScheduleRequest) (*models.SetScheduleResponse, error) {
		return nil, apiErr
	}}
	_, err := run(mock, "set", "--query-id", "5", "--cron", "1 2 3 4 5", "--performance", "small")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported cron")
}

func TestDeleteWithoutSchedule(t *testing.T) {
	var gotID int
	mock := &mockClient{deleteDashboardFn: func(id int) (*models.DeleteScheduleResponse, error) {
		gotID = id
		return &models.DeleteScheduleResponse{Deleted: false}, nil
	}}
	buf, err := run(mock, "delete", "--dashboard-id", "9")
	require.NoError(t, err)
	assert.Equal(t, 9, gotID)
	assert.Equal(t, "There was no schedule to delete\n", buf.String())
}

func TestDeleteQueryJSON(t *testing.T) {
	var gotID int
	mock := &mockClient{deleteQueryFn: func(id int) (*models.DeleteScheduleResponse, error) {
		gotID = id
		return &models.DeleteScheduleResponse{Deleted: true}, nil
	}}
	buf, err := run(mock, "delete", "--query-id", "12", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, 12, gotID)
	assert.JSONEq(t, `{"deleted":true}`, buf.String())
}

func TestList(t *testing.T) {
	var gotOpts models.ListSchedulesOptions
	mock := &mockClient{listFn: func(opts models.ListSchedulesOptions) (*models.ListSchedulesResponse, error) {
		gotOpts = opts
		return &models.ListSchedulesResponse{Total: 40, Schedules: []models.Schedule{{
			Type: "dashboard", DashboardID: ptr(int64(9)), CronExpression: "0 */6 * * *", Performance: "medium",
		}}}, nil
	}}
	buf, err := run(mock, "list", "--limit", "1")
	require.NoError(t, err)
	assert.Equal(t, models.ListSchedulesOptions{Limit: 1}, gotOpts)
	assert.Regexp(t, `dashboard\s+dashboard 9\s+0 \*/6 \* \* \*\s+medium\s+-\s+-`, buf.String())
	assert.Contains(t, buf.String(), "1 of 40 schedules")
}
