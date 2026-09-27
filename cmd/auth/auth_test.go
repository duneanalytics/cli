package auth_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/cmd/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	authconfig.SetDirFunc(func() (string, error) { return dir, nil })
	t.Cleanup(authconfig.ResetDirFunc)
	return dir
}

func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.PersistentFlags().String("api-key", "", "")
	root.SetContext(context.Background())
	root.AddCommand(auth.NewAuthCmd())
	return root
}

func TestAuthWithFlag(t *testing.T) {
	dir := setup(t)

	root := newRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth", "--api-key", "flag_key"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "flag_key")
	assert.Contains(t, buf.String(), "API key saved to")
}

func TestAuthWithEnvVar(t *testing.T) {
	dir := setup(t)

	t.Setenv("DUNE_API_KEY", "env_key")

	root := newRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"auth"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "env_key")
}

func TestAuthNonInteractiveStdinFails(t *testing.T) {
	setup(t)

	// Unset env var so it doesn't interfere
	t.Setenv("DUNE_API_KEY", "")

	root := newRoot()
	root.SetIn(strings.NewReader("prompt_key\n"))
	root.SetArgs([]string{"auth"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key provided")
}

func TestAuthEmptyInput(t *testing.T) {
	setup(t)

	t.Setenv("DUNE_API_KEY", "")

	root := newRoot()
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"auth"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key provided; pass --api-key, set DUNE_API_KEY, or run dune auth in an interactive terminal")
}
