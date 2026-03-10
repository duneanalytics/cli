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

func TestCommandPathFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "simple subcommand",
			args: []string{"dune", "query", "list"},
			want: "query list",
		},
		{
			name: "root flag before subcommand",
			args: []string{"dune", "--api-key", "KEY", "query", "list"},
			want: "query list",
		},
		{
			name: "flag with equals syntax",
			args: []string{"dune", "--api-key=KEY", "query", "list"},
			want: "query list",
		},
		{
			name: "trailing flags after subcommand",
			args: []string{"dune", "query", "list", "--limit", "10"},
			want: "query list",
		},
		{
			name: "flags before and after subcommand",
			args: []string{"dune", "--api-key", "KEY", "query", "list", "--limit", "10"},
			want: "query list",
		},
		{
			name: "binary only",
			args: []string{"dune"},
			want: "unknown",
		},
		{
			name: "only flags",
			args: []string{"dune", "--help"},
			want: "unknown",
		},
		{
			name: "single subcommand",
			args: []string{"dune", "auth"},
			want: "auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, commandPathFromArgs(tt.args))
		})
	}
}
