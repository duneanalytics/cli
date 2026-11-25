package query_test

import (
	"testing"

	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const alertResult = `{"alert_id":"a-1","query_id":42,"channel":"slack","cron_expression":"0 * * * *",
	"next_execution_time":"2026-11-26T10:00:00Z"}`

func runAlert(fake *dunemcptest.Fake, args ...string) (string, error) {
	root, buf := newMCPRoot(fake)
	root.SetArgs(append([]string{"query", "alert"}, args...))
	err := root.Execute()
	return buf.String(), err
}

func TestAlertSendsEmailChannel(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"setQueryAlert": alertResult}}
	_, err := runAlert(fake, "42", "--cron", "0 * * * *", "--performance", "small",
		"--email", "a@x.com", "--email", "b@x.com")
	require.NoError(t, err)
	assert.Equal(t, "setQueryAlert", fake.Last().Tool)
	assert.JSONEq(t, `{"query_id":42,"cron_expression":"0 * * * *","performance":"small",
		"channel":{"type":"email","emails":["a@x.com","b@x.com"]}}`, fake.Last().Args)
}

func TestAlertWebhookChannelsAndOutputHideTheURL(t *testing.T) {
	for flag, typ := range map[string]string{"--slack-webhook": "slack", "--discord-webhook": "discord", "--webhook": "webhook"} {
		fake := &dunemcptest.Fake{Results: map[string]string{"setQueryAlert": alertResult}}
		out, err := runAlert(fake, "42", "--cron", "0 * * * *", "--performance", "small", flag, "https://hooks.example/SECRET")
		require.NoError(t, err)
		assert.JSONEq(t, `{"query_id":42,"cron_expression":"0 * * * *","performance":"small",
			"channel":{"type":"`+typ+`","url":"https://hooks.example/SECRET"}}`, fake.Last().Args)
		assert.NotContains(t, out, "SECRET")
	}
}

func TestAlertPrintsAlertAndSchedule(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"setQueryAlert": alertResult}}
	out, err := runAlert(fake, "42", "--cron", "0 * * * *", "--performance", "small", "--slack-webhook", "https://h")
	require.NoError(t, err)
	assert.Equal(t, "Set slack alert a-1 on query 42\nSchedule: 0 * * * * (UTC), next run 2026-11-26T10:00:00Z\n", out)
}

func TestAlertRejectsBadInputBeforeCalling(t *testing.T) {
	for name, args := range map[string][]string{
		"no channel":    {"42", "--cron", "0 * * * *", "--performance", "small"},
		"two channels":  {"42", "--cron", "0 * * * *", "--performance", "small", "--email", "a@x.com", "--webhook", "https://h"},
		"empty webhook": {"42", "--cron", "0 * * * *", "--performance", "small", "--webhook", ""},
		"no cron":       {"42", "--performance", "small", "--email", "a@x.com"},
		"bad query id":  {"0", "--cron", "0 * * * *", "--performance", "small", "--email", "a@x.com"},
	} {
		t.Run(name, func(t *testing.T) {
			fake := &dunemcptest.Fake{}
			_, err := runAlert(fake, args...)
			require.Error(t, err)
			assert.Empty(t, fake.Calls)
		})
	}
}
