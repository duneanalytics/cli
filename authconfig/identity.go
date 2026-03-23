package authconfig

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UserIdentity holds the cached identity resolved from the API key.
type UserIdentity struct {
	CustomerID string `yaml:"customer_id"`  // e.g. "user_123" or "team_456"
	APIKeyHash string `yaml:"api_key_hash"` // SHA-256 of the API key for cache invalidation
}

// identityFileName is the file storing the cached user identity.
const identityFileName = "user_identity.yaml"

// LoadIdentity reads the cached user identity. Returns nil, nil if the file does not exist.
func LoadIdentity() (*UserIdentity, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, identityFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var id UserIdentity
	if err := yaml.Unmarshal(data, &id); err != nil {
		return nil, err
	}
	return &id, nil
}

// SaveIdentity writes the user identity cache to disk.
func SaveIdentity(id *UserIdentity) error {
	dir, err := Dir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := yaml.Marshal(id)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, identityFileName), data, 0o600)
}

// HashAPIKey returns the hex-encoded SHA-256 hash of an API key.
func HashAPIKey(apiKey string) string {
	h := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", h)
}
