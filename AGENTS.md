# Agent Instructions for cli

> Keep this document lean. It is a navigation and working guide; usage belongs in `README.md`.

## Overview

`dune` is Dune's public command-line client. Commands call the REST API through
`duneapi-client-go`; only features with no REST endpoint call the Dune MCP server
through `dunemcp/`. Users and scripts depend on command names, flags, exit codes and
`-o json` output.

## Repository Map

| Path                     | Responsibility                                           |
| ------------------------ | -------------------------------------------------------- |
| `cmd/`                   | One package per command group; `cmd/main.go` is the entry. |
| `cli/`                   | Root command and wiring.                                 |
| `cmdutil/`, `output/`    | Shared flag parsing, clients and text/JSON output.       |
| `dunemcp/`               | Small MCP client and its test fake.                      |
| `authconfig/`            | API key and identity stored on the user's disk.          |
| `tracking/`              | Amplitude usage telemetry.                               |
| `install.sh`, `.goreleaser.yml`, `flake.nix` | Distribution.                  |

## Build & Test

```bash
make lint
make test
```

`flake.nix` pins a `vendorHash`; update it when `go.mod` or `go.sum` changes.

## Review risk boundaries

Claude auto-approve may approve a pull request only when every change is low risk:
bug fixes and features inside `cmd/`, `cli/`, `cmdutil/`, `output/` and `dunemcp/`,
tests, documentation, and patch or minor bumps of existing Go dependencies.

These changes need a human review:

- Everything under `.github/`, `install.sh`, `.goreleaser.yml`, and `AGENTS.md` or `CLAUDE.md`.
- `authconfig/`: how the API key is read, stored or written, and file permissions.
- `tracking/`: what telemetry is sent, when, or how users opt out.
- Sending the API key or user data to a new host, or changing which host a command calls.
- Breaking changes for users and scripts: removing or renaming a command or flag,
  changing a flag's default, changing the exit code of an existing case, or removing or
  renaming a key in `-o json` output.
- Newly added dependencies, major-version bumps of a dependency (a `0.x` minor counts as
  major), the Go version in `go.mod`, and dependencies from a source other than the Go
  module proxy (`replace` directives, forks).
- Deleting a test whose subject code stays.
