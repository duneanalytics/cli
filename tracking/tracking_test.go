package tracking

import (
	"testing"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyClient records events sent via Track for test assertions.
type spyClient struct {
	events []amplitude.Event
}

func (s *spyClient) Track(event amplitude.Event)                                              { s.events = append(s.events, event) }
func (s *spyClient) Identify(amplitude.Identify, amplitude.EventOptions)                      {}
func (s *spyClient) GroupIdentify(string, string, amplitude.Identify, amplitude.EventOptions) {}
func (s *spyClient) SetGroup(string, []string, amplitude.EventOptions)                        {}
func (s *spyClient) Revenue(amplitude.Revenue, amplitude.EventOptions)                        {}
func (s *spyClient) Flush()                                                                   {}
func (s *spyClient) Shutdown()                                                                {}
func (s *spyClient) Add(amplitude.Plugin)                                                     {}
func (s *spyClient) Remove(string)                                                            {}
func (s *spyClient) Config() amplitude.Config                                                 { return amplitude.Config{} }

func newTestTracker(spy *spyClient) *Tracker {
	return &Tracker{client: spy, version: "test", userID: "cli", enabled: true}
}

func TestTracker_DisabledNoOp(t *testing.T) {
	tr := New(Config{Enabled: false})
	assert.False(t, tr.enabled)
	// Should not panic.
	tr.Track("test cmd", StatusSuccess, "", 100, false)
	tr.Track("sim evm balances", StatusSuccess, "", 100, true)
	tr.Shutdown()
}

func TestTracker_EmptyAmplitudeKeyDisabled(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: ""})
	assert.False(t, tr.enabled)
}

func TestTracker_DefaultUserIDCli(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	assert.Equal(t, "cli", tr.userID, "userID should default to 'cli'")
}

func TestTracker_SetUserID(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	assert.Equal(t, "cli", tr.userID)

	tr.SetUserID("user_123")
	assert.Equal(t, "123", tr.userID, "user_ prefix should be stripped")
}

func TestToAmplitudeUserID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user_123", "123"},
		{"user_0", "0"},
		{"team_456", "system_456"},
		{"team_1", "system_1"},
		{"anonymous", "anonymous"},
		{"something_else", "something_else"},
		{"", ""},
		{"user_", ""},
		{"team_", "system_"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, toAmplitudeUserID(tc.input))
		})
	}
}

func TestTracker_TrackWithoutSetUserID(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	// Should not panic — events are sent with "cli" UserID.
	tr.Track("test cmd", StatusSuccess, "", 100, false)
	tr.Shutdown()
}

func TestTrack_IsSim(t *testing.T) {
	spy := &spyClient{}
	tr := newTestTracker(spy)

	tr.Track("query list", StatusSuccess, "", 42, false)
	tr.Track("sim evm balances", StatusSuccess, "", 99, true)

	require.Len(t, spy.events, 2)

	// Non-sim event
	props0 := spy.events[0].EventProperties
	assert.Equal(t, "CLI Command Executed", spy.events[0].EventType)
	assert.Equal(t, "cli", spy.events[0].UserID)
	assert.Equal(t, "query list", props0["command_path"])
	assert.Equal(t, false, props0["is_sim"])

	// Sim event
	props1 := spy.events[1].EventProperties
	assert.Equal(t, "CLI Command Executed", spy.events[1].EventType)
	assert.Equal(t, "sim evm balances", props1["command_path"])
	assert.Equal(t, true, props1["is_sim"])
}

func TestTrack_SetUserIDReflectedInEvents(t *testing.T) {
	spy := &spyClient{}
	tr := newTestTracker(spy)

	tr.SetUserID("user_42")
	tr.Track("query run", StatusSuccess, "", 10, false)

	require.Len(t, spy.events, 1)
	assert.Equal(t, "42", spy.events[0].UserID)
}
