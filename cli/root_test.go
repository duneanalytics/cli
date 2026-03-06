package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRootTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	authconfig.SetDirFunc(func() (string, error) { return dir, nil })
	t.Cleanup(authconfig.ResetDirFunc)
	prevAPIKey, hadAPIKey := os.LookupEnv("DUNE_API_KEY")
	require.NoError(t, os.Unsetenv("DUNE_API_KEY"))
	t.Cleanup(func() {
		if hadAPIKey {
			_ = os.Setenv("DUNE_API_KEY", prevAPIKey)
			return
		}
		_ = os.Unsetenv("DUNE_API_KEY")
	})
	apiKeyFlag = ""
	return dir
}

func TestPersistentPreRunESkipAuth(t *testing.T) {
	setupRootTest(t)

	cmd := rootCmd
	cmd.Annotations = map[string]string{"skipAuth": "true"}
	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)
}

func TestPersistentPreRunEMalformedConfig(t *testing.T) {
	dir := setupRootTest(t)
	err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":\tbad\nyaml{["), 0o600)
	require.NoError(t, err)

	cmd := rootCmd
	cmd.Annotations = nil
	err = rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading auth config")
}

func TestPersistentPreRunEEmptyConfig(t *testing.T) {
	setupRootTest(t)
	require.NoError(t, authconfig.Save(&authconfig.Config{APIKey: "  "}))

	cmd := rootCmd
	cmd.Annotations = nil
	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty API key in config")
}

func TestPersistentPreRunEConfigFallback(t *testing.T) {
	setupRootTest(t)
	require.NoError(t, authconfig.Save(&authconfig.Config{APIKey: "saved_key"}))

	cmd := rootCmd
	cmd.Annotations = nil
	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)
}
