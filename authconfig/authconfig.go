package authconfig

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config holds the persisted CLI configuration.
type Config struct {
	APIKey string `yaml:"api_key"`
}

// configDirFunc allows tests to override the config directory.
var (
	configDirFunc = defaultDir
	configMu      sync.RWMutex
)

func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "dune"), nil
}

// SetDirFunc overrides the config directory function (for testing).
func SetDirFunc(fn func() (string, error)) {
	configMu.Lock()
	defer configMu.Unlock()
	configDirFunc = fn
}

// ResetDirFunc restores the default config directory function.
func ResetDirFunc() {
	configMu.Lock()
	defer configMu.Unlock()
	configDirFunc = defaultDir
}

// Dir returns the config directory path ($HOME/.config/dune).
func Dir() (string, error) {
	configMu.RLock()
	defer configMu.RUnlock()
	return configDirFunc()
}

// Path returns the full path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads and parses the config file. Returns nil, nil if the file does not exist.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory (0700) and file (0600) as needed.
func Save(cfg *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o600)
}
