package main

import "github.com/duneanalytics/cli/cli"

// Set by GoReleaser via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Execute(version, commit, date)
}
