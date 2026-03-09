package tracking

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/amplitude/analytics-go/amplitude"
	"github.com/google/uuid"
)

const (
	StatusSuccess = "success"
	StatusError   = "error"
)

type Tracker struct {
	client  amplitude.Client
	version string
	userID  string
	enabled bool
}

type Config struct {
	AmplitudeKey string
	CLIVersion   string
	ConfigDir    string
	Enabled      bool
}

func New(cfg Config) *Tracker {
	if !cfg.Enabled || cfg.AmplitudeKey == "" {
		return &Tracker{enabled: false}
	}

	ampConfig := amplitude.NewConfig(cfg.AmplitudeKey)
	ampConfig.ServerZone = "EU"
	ampConfig.FlushQueueSize = 1
	ampConfig.FlushInterval = 1 * time.Second
	if !isDevVersion(cfg.CLIVersion) {
		ampConfig.Logger = silentLogger{}
	}

	return &Tracker{
		client:  amplitude.NewClient(ampConfig),
		version: cfg.CLIVersion,
		userID:  loadOrCreateAnonID(cfg.ConfigDir),
		enabled: true,
	}
}

func (t *Tracker) Track(commandPath, status, errMsg string, durationMs int64) {
	if !t.enabled || t.client == nil {
		return
	}
	if len(errMsg) > 256 {
		errMsg = errMsg[:256]
	}
	t.client.Track(amplitude.Event{
		UserID:    t.userID,
		EventType: "CLI Command Executed",
		EventProperties: map[string]interface{}{
			"command_path":  commandPath,
			"status":        status,
			"duration_ms":   durationMs,
			"error_message": errMsg,
			"cli_version":   t.version,
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
		},
	})
}

func (t *Tracker) Shutdown() {
	if t.enabled && t.client != nil {
		t.client.Shutdown()
	}
}

// isDevVersion returns true for local / non-release builds.
func isDevVersion(v string) bool {
	return v == "" || v == "dev"
}

// silentLogger suppresses all amplitude SDK log output.
type silentLogger struct{}

func (silentLogger) Debugf(string, ...interface{}) {}
func (silentLogger) Infof(string, ...interface{})  {}
func (silentLogger) Warnf(string, ...interface{})  {}
func (silentLogger) Errorf(string, ...interface{}) {}

const (
	anonIDFile   = "anonymous_id"
	anonFallback = "anonymous"
)

func loadOrCreateAnonID(configDir string) string {
	if configDir == "" {
		return anonFallback
	}

	path := filepath.Join(configDir, anonIDFile)

	data, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if _, err := uuid.Parse(id); err == nil {
			return id
		}
	}

	id := uuid.New().String()

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return anonFallback
	}
	if err := os.WriteFile(path, []byte(id), 0o644); err != nil {
		return anonFallback
	}

	return id
}
