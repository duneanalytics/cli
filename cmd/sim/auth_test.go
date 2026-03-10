package sim_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/cmd/sim"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func setupAuthTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	authconfig.SetDirFunc(func() (string, error) { return dir, nil })
	t.Cleanup(authconfig.ResetDirFunc)
	// Clear env var so it doesn't interfere with tests.
	t.Setenv("DUNE_SIM_API_KEY", "")
	return dir
}

func newSimAuthRoot() *cobra.Command {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())

	simCmd := sim.NewSimCmd()
	root.AddCommand(simCmd)

	return root
}

func TestSimAuth_WithFlag(t *testing.T) {
	dir := setupAuthTest(t)

	root := newSimAuthRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "auth", "--api-key", "sk_sim_flag_key"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "sk_sim_flag_key")
	assert.Contains(t, buf.String(), "Sim API key saved to")
}

func TestSimAuth_WithEnvVar(t *testing.T) {
	dir := setupAuthTest(t)

	t.Setenv("DUNE_SIM_API_KEY", "sk_sim_env_key")

	root := newSimAuthRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "auth"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "sk_sim_env_key")
}

func TestSimAuth_WithPrompt(t *testing.T) {
	dir := setupAuthTest(t)

	root := newSimAuthRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetIn(strings.NewReader("sk_sim_prompt_key\n"))
	root.SetArgs([]string{"sim", "auth"})
	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "sk_sim_prompt_key")
}

func TestSimAuth_EmptyInput(t *testing.T) {
	setupAuthTest(t)

	root := newSimAuthRoot()
	root.SetIn(strings.NewReader("\n"))
	root.SetArgs([]string{"sim", "auth"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no API key provided")
}

func TestSimAuth_PreservesExistingConfig(t *testing.T) {
	dir := setupAuthTest(t)

	// Pre-populate config with existing fields.
	existing := &authconfig.Config{
		APIKey: "existing_dune_key",
	}
	telemetryTrue := true
	existing.Telemetry = &telemetryTrue
	require.NoError(t, authconfig.Save(existing))

	root := newSimAuthRoot()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"sim", "auth", "--api-key", "sk_sim_new"})
	require.NoError(t, root.Execute())

	// Verify all fields are preserved.
	data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)

	var cfg authconfig.Config
	require.NoError(t, yaml.Unmarshal(data, &cfg))

	assert.Equal(t, "existing_dune_key", cfg.APIKey, "existing api_key should be preserved")
	assert.Equal(t, "sk_sim_new", cfg.SimAPIKey, "sim_api_key should be set")
	require.NotNil(t, cfg.Telemetry, "telemetry should be preserved")
	assert.True(t, *cfg.Telemetry, "telemetry value should be preserved")
}
