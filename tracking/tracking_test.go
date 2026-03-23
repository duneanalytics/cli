package tracking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracker_DisabledNoOp(t *testing.T) {
	tr := New(Config{Enabled: false})
	assert.False(t, tr.enabled)
	// Should not panic.
	tr.Track("test cmd", StatusSuccess, "", 100)
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
	// Should not panic — events are sent with "anonymous" UserID.
	tr.Track("test cmd", StatusSuccess, "", 100)
	tr.Shutdown()
}
