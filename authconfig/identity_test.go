package authconfig_test

import (
	"testing"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadIdentity(t *testing.T) {
	setupTempDir(t)

	want := &authconfig.UserIdentity{
		CustomerID: "user_42",
		APIKeyHash: authconfig.HashAPIKey("test-api-key"),
	}
	require.NoError(t, authconfig.SaveIdentity(want))

	got, err := authconfig.LoadIdentity()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLoadIdentityNonExistent(t *testing.T) {
	setupTempDir(t)

	id, err := authconfig.LoadIdentity()
	assert.NoError(t, err)
	assert.Nil(t, id)
}

func TestHashAPIKeyDeterministic(t *testing.T) {
	h1 := authconfig.HashAPIKey("my-key")
	h2 := authconfig.HashAPIKey("my-key")
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, authconfig.HashAPIKey("other-key"))
}
