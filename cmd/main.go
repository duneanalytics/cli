package main

import (
	"runtime/debug"
	"strings"

	"github.com/duneanalytics/cli/cli"
)

// Set by GoReleaser or Makefile via ldflags.
var (
	version      = ""
	commit       = ""
	date         = ""
	amplitudeKey = ""
)

func main() {
	resolveVersion()
	cli.Execute(version, commit, date, amplitudeKey)
}

// resolveVersion fills in version/commit/date from Go build info
// when they haven't been set via ldflags (i.e. plain go build / go install).
func resolveVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		setDefaults()
		return
	}

	// For go install ...@vX.Y.Z, the module version is clean (e.g. "v0.0.2").
	// For local builds it is "(devel)" — not useful.
	if version == "" {
		if v := info.Main.Version; v != "" && v != "(devel)" && !strings.Contains(v, "-") {
			version = strings.TrimPrefix(v, "v")
		}
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if commit == "" && s.Value != "" {
				commit = s.Value
				if len(commit) > 12 {
					commit = commit[:12]
				}
			}
		case "vcs.time":
			if date == "" && s.Value != "" {
				date = s.Value
			}
		}
	}

	setDefaults()
}

func setDefaults() {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
	if date == "" {
		date = "unknown"
	}
}
