package authconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	authconfig.SetDirFunc(func() (string, error) { return dir, nil })
	t.Cleanup(authconfig.ResetDirFunc)
	return dir
}

func TestSaveAndLoad(t *testing.T) {
	setupTempDir(t)

	want := &authconfig.Config{APIKey: "dune_test_key_123"}
	require.NoError(t, authconfig.Save(want))

	got, err := authconfig.Load()
	require.NoError(t, err)
	assert.Equal(t, want.APIKey, got.APIKey)
}

func TestLoadNonExistent(t *testing.T) {
	setupTempDir(t)

	cfg, err := authconfig.Load()
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestLoadMalformedYAML(t *testing.T) {
	dir := setupTempDir(t)

	err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":\tbad\nyaml{["), 0o600)
	require.NoError(t, err)

	_, err = authconfig.Load()
	assert.Error(t, err)
}

func TestSaveCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "sub", "dir")
	authconfig.SetDirFunc(func() (string, error) { return nested, nil })
	t.Cleanup(authconfig.ResetDirFunc)

	require.NoError(t, authconfig.Save(&authconfig.Config{APIKey: "key"}))

	info, err := os.Stat(nested)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFilePermissions(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "newdir")
	authconfig.SetDirFunc(func() (string, error) { return dir, nil })
	t.Cleanup(authconfig.ResetDirFunc)

	require.NoError(t, authconfig.Save(&authconfig.Config{APIKey: "key"}))

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}
