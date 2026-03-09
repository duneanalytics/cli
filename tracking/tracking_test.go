package tracking

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLoadOrCreateAnonID_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	id := loadOrCreateAnonID(dir)

	_, err := uuid.Parse(id)
	assert.NoError(t, err, "should return a valid UUID")
	assert.NotEqual(t, anonFallback, id)

	data, err := os.ReadFile(filepath.Join(dir, anonIDFile))
	require.NoError(t, err)
	assert.Equal(t, id, string(data))
}

func TestLoadOrCreateAnonID_ReusesExisting(t *testing.T) {
	dir := t.TempDir()
	knownID := "550e8400-e29b-41d4-a716-446655440000"
	require.NoError(t, os.WriteFile(filepath.Join(dir, anonIDFile), []byte(knownID), 0o644))

	id := loadOrCreateAnonID(dir)
	assert.Equal(t, knownID, id)
}

func TestLoadOrCreateAnonID_InvalidDir(t *testing.T) {
	id := loadOrCreateAnonID("/nonexistent/path/that/should/not/exist")
	assert.Equal(t, anonFallback, id)
}
