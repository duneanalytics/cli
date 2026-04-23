# Dune CLI

A command-line interface for interacting with the [Dune](https://dune.com/) API — query data, manage visualizations, and build dashboards from your terminal.

## Installation

### Install Script
```bash
curl -sSfL https://github.com/duneanalytics/cli/raw/main/install.sh | bash
```

### [AUR](https://aur.archlinux.org/packages/dune-cli)
```bash
yay -S dune-cli # -bin | -git
```

## Authentication

```bash
# Save your API key to ~/.config/dune/config.yaml
dune auth --api-key <key>

# Or run interactively (prompts for key)
dune auth

# Or set via environment variable
export DUNE_API_KEY=<key>
```

The `--api-key` flag is available on all commands to override the stored key.

## Commands

### `dune query`

Manage and execute Dune queries.

| Command | Description |
|---------|-------------|
| `query create --name <name> --sql <sql> [--description] [--private] [--temp]` | Create a new saved query |
| `query get <query-id>` | Get a saved query's details and SQL |
| `query update <query-id> [--name] [--sql] [--description] [--private] [--tags]` | Update an existing query |
| `query archive <query-id>` | Archive a saved query |
| `query run <query-id> [--param key=value] [--performance free\|small\|medium\|large] [--limit] [--timeout] [--no-wait]` | Execute a saved query and display results |
| `query run-sql --sql <sql> [--param key=value] [--performance free\|small\|medium\|large] [--limit] [--timeout] [--no-wait]` | Execute raw SQL directly |

### `dune execution`

Manage query executions.

| Command | Description |
|---------|-------------|
| `execution results <execution-id> [--limit] [--offset] [--timeout] [--no-wait]` | Fetch results of a query execution |

### `dune dataset`

Search the Dune dataset catalog.

| Command | Description |
|---------|-------------|
| `dataset search [--query] [--categories] [--blockchains] [--schemas] [--dataset-types] [--owner-scope] [--include-private] [--include-schema] [--include-metadata] [--limit] [--offset]` | Search for datasets |
| `dataset search-by-contract --contract-address <address> [--blockchains] [--include-schema] [--limit] [--offset]` | Search for decoded tables by contract address |

Categories: `canonical`, `decoded`, `spell`, `community`

### `dune visualization` (alias: `viz`)

Create and manage visualizations on saved queries.

| Command | Description |
|---------|-------------|
| `viz create --query-id <id> --name <name> --options <json> [--type chart\|table\|counter\|...]` | Create a visualization on a saved query |
| `viz get <visualization-id>` | Get a visualization's details and options |
| `viz update <visualization-id> [--name] [--type] [--description] [--options]` | Update an existing visualization |
| `viz delete <visualization-id>` | Permanently delete a visualization |
| `viz list --query-id <id> [--limit] [--offset]` | List all visualizations for a query |

Supported types: `chart`, `table`, `counter`, `pivot`, `cohort`, `funnel`, `choropleth`, `sankey`, `sunburst_sequence`, `word_cloud`.

### `dune dashboard` (alias: `dash`)

Create and manage dashboards.

| Command | Description |
|---------|-------------|
| `dashboard create --name <name> [--visualization-ids 1,2,3] [--text-widgets <json>] [--columns-per-row 1\|2\|3] [--private]` | Create a new dashboard |
| `dashboard get <dashboard-id>` or `dashboard get --owner <handle> --slug <slug>` | Get a dashboard's details and widgets |
| `dashboard update <dashboard-id> [--name] [--slug] [--private] [--tags] [--visualization-widgets <json>] [--text-widgets <json>]` | Update an existing dashboard |
| `dashboard archive <dashboard-id>` | Archive a dashboard |

### `dune docs`

Search and browse Dune documentation. No authentication required.

| Command | Description |
|---------|-------------|
| `docs search --query <text> [--api-reference-only] [--code-only]` | Search the Dune documentation |

### `dune usage`

Show credit and resource usage for your account.

```bash
dune usage [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
```

## Output Format

All commands (except `auth`) support `-o, --output <format>` with `text` (default) or `json`.
