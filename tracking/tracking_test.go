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

func TestTracker_DefaultUserIDAnonymous(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	assert.Equal(t, "anonymous", tr.userID, "userID should default to 'anonymous'")
}

func TestTracker_SetUserID(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	assert.Equal(t, "anonymous", tr.userID)

	tr.SetUserID("user_123")
	assert.Equal(t, "user_123", tr.userID)
}

func TestTracker_TrackWithoutSetUserID(t *testing.T) {
	tr := New(Config{Enabled: true, AmplitudeKey: "test-key"})
	// Should not panic — events are sent with "anonymous" UserID.
	tr.Track("test cmd", StatusSuccess, "", 100)
	tr.Shutdown()
}
