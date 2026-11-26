# Dune CLI

A command-line interface for interacting with the [Dune](https://dune.com/) API — query data, manage visualizations, and build dashboards from your terminal.

## Installation

### Homebrew (macOS and Linux)

```bash
brew install duneanalytics/tap/dune-cli
```

### Nix

```bash
# Run without installing
nix run github:duneanalytics/cli

# Install into your profile
nix profile add github:duneanalytics/cli
```

### Install script

```bash
curl -sSfL https://github.com/duneanalytics/cli/raw/main/install.sh | bash
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
| `query create --name <name> --sql <sql> [--description] [--private] [--temp] [--folder-id\|--folder-name] [--parameters <json>]` | Create a new saved query |
| `query create-and-run --name <name> --sql <sql> [create flags] [--performance] [--timeout] [--limit] [--chart-type <type> --config <json>] [--viz-name]` | Save a query, run it, optionally chart it, and show the first rows |
| `query search --query <text> [--owned-only] [--limit]` | Find saved queries by name |
| `query get <query-id>` | Get a saved query's details, SQL, and latest execution |
| `query update <query-id> [--name] [--sql] [--description] [--private] [--archived] [--temp] [--tags] [--parameters <json>]` | Update an existing query |
| `query archive <query-id>` | Archive a saved query |
| `query run <query-id> [--param key=value] [--performance small\|medium\|large] [--limit] [--timeout] [--no-wait]` | Execute a saved query and display results |
| `query run-sql --sql <sql> [--param key=value] [--performance small\|medium\|large] [--limit] [--timeout] [--no-wait]` | Execute raw SQL directly |
| `query alert <query-id> --cron <expr> --performance small\|medium\|large --email <addr>... \| --slack-webhook <url> \| --discord-webhook <url> \| --webhook <url>` | Schedule a query and get notified each time a run returns rows |

### `dune matview` (alias: `mv`)

Create and manage materialized views — query results persisted into a queryable table, optionally refreshed on a schedule.

| Command | Description |
|---------|-------------|
| `matview create --name <name> --query-id <id> [--performance small\|medium\|large] [--cron <expr>] [--expires-at <RFC3339>]` | Materialize a saved query into a table |
| `matview get <name-or-id>` | Get a matview's metadata, size, and refresh schedule |
| `matview list [--limit] [--offset] [--all]` | List the materialized views you own |
| `matview update <name> [--performance] [--cron <expr>] [--no-schedule] [--expires-at]` | Update settings or the schedule (preserves the schedule unless changed) |
| `matview refresh <name> [--performance small\|medium\|large]` | Trigger an on-demand refresh |
| `matview delete <name>` | Permanently delete a matview and its schedule |

`create` takes a bare name (e.g. `result_token_summary`); `get`, `update`, `refresh`, and `delete` take the fully-qualified SQL name (e.g. `dune.my_team.result_token_summary`). `get` also takes the matview ID that `matview list -o json` returns.

A matview always has its source query's privacy. To make one private, make the query private with `dune query update <query-id> --private`. The old `--private` flag is deprecated: it fails when it asks for a privacy the query does not have.

### `dune execution`

Manage query executions.

| Command | Description |
|---------|-------------|
| `execution results <execution-id> [--limit] [--offset] [--timeout] [--no-wait] [--columns a,b] [--filter <expr>] [--sort-by <expr>] [--sample <n>]` | Fetch results of a query execution, optionally shaped on the server |

`--columns`, `--filter` (a WHERE-like expression such as `'"amount" > 0'`), `--sort-by` (an ORDER BY fragment such as `'"block_time" DESC'`), and `--sample` shape the results before they are sent, so you download only the rows and columns you need. `--sample` returns about that many rows, chosen at random, and cannot be combined with `--limit` or `--offset`.

### `dune dataset`

Search the Dune dataset catalog.

| Command | Description |
|---------|-------------|
| `dataset search [--query] [--categories] [--blockchains] [--schemas] [--dataset-types] [--owner-scope] [--include-private] [--include-schema] [--include-metadata] [--limit] [--offset]` | Search for datasets; the top 3 hits include their columns |
| `dataset search-by-contract --contract-address <address> [--blockchains] [--include-schema] [--limit] [--offset]` | Search for decoded tables by contract address |
| `dataset blockchains [--categories] [--query] [--limit] [--offset]` | List blockchains with the number of tables on each |
| `dataset size <table>` | Show a table's storage size, e.g. `dataset size ethereum.blocks` |

Categories: `canonical`, `decoded`, `spell`, `community`.

### `dune visualization` (alias: `viz`)

Create and manage visualizations on saved queries.

| Command | Description |
|---------|-------------|
| `viz create --query-id <id> --name <name> --chart-type <type> --config <json> [--description]` | Create a chart from a compact chart config |
| `viz create --query-id <id> --name <name> --options <json> [--type chart\|table\|counter\|...]` | Create a visualization from raw Dune options |
| `viz get <visualization-id>` | Get a visualization's details, options, chart type, and chart and image links |
| `viz update <visualization-id> [--name] [--description] [--counter-label] [--chart-type <type> --config <json>] [--type] [--options]` | Update an existing visualization |
| `viz delete <visualization-id>` | Permanently delete a visualization |
| `viz list --query-id <id> [--limit] [--offset]` | List all visualizations for a query |

Supported types: `chart`, `table`, `counter`, `pivot`, `cohort`, `funnel`, `choropleth`, `sankey`, `sunburst_sequence`, `word_cloud`.

`--chart-type` is one of `column`, `line`, `area`, `scatter`, `pie`, `counter`, or `table`, and `--config` describes the axes and series (run `dune viz create --help` for its format). Column names must match the query's result columns exactly; when the query has completed results, Dune rejects a config that names other columns. Look at the columns first with `dune execution results`. `--config` and `--options` accept inline JSON, `@path`, or `@-` for standard input.

### `dune dashboard` (alias: `dash`)

Create and manage dashboards.

| Command | Description |
|---------|-------------|
| `dashboard create --name <name> [--visualization-ids 1,2,3] [--text-widgets <json>] [--columns-per-row 1\|2\|3] [--private]` | Create a new dashboard |
| `dashboard get <dashboard-id>` or `dashboard get --owner <handle> --slug <slug>` | Get a dashboard's details and widgets |
| `dashboard update <dashboard-id> [--name] [--slug] [--private] [--tags] [--visualization-widgets <json>] [--text-widgets <json>]` | Update an existing dashboard |
| `dashboard update-widget <dashboard-id> --widget-type visualization\|text\|param --widget-id <id> [--row --col --size-x --size-y] [--text <md>]` | Move or resize one widget, or edit a text widget, keeping the others |
| `dashboard image <dashboard-id> [--file <path>]` | Save a JPEG of the whole dashboard to check its layout |
| `dashboard archive <dashboard-id>` | Archive a dashboard |
| `dashboard search --query <text> [--owned-only] [--limit]` | Find dashboards by name |

### `dune folder`

Organize saved queries and dashboards into library folders.

| Command | Description |
|---------|-------------|
| `folder list [--parent-path /] [--name-contains] [--limit] [--offset]` | List folders under a path |
| `folder get <folder-id> [--limit] [--offset]` | Show a folder and the queries, dashboards, and subfolders in it |
| `folder create --name <name> [--parent-id] [--description] [--color] [--icon]` | Create a folder |
| `folder update <folder-id> [--name] [--description] [--color] [--icon]` | Rename or restyle a folder; an empty `--description` clears it |
| `folder move <folder-id> --to <folder-id>\|--to-root` | Move a folder |
| `folder move-content --query-ids 1,2 --dashboard-ids 3 --to <folder-id>\|--to-root` | Move queries and dashboards into a folder |
| `folder delete <folder-id> [--recursive]` | Delete a folder; `--recursive` also archives everything inside |

### `dune schedule`

Run saved queries and refresh dashboards on a schedule. A query or dashboard has at most one schedule.

| Command | Description |
|---------|-------------|
| `schedule list [--limit] [--offset]` | List your schedules |
| `schedule set --query-id <id>\|--dashboard-id <id> --cron <expr> --performance small\|medium\|large [--expires-at <RFC3339>]` | Create or replace a schedule |
| `schedule delete --query-id <id>\|--dashboard-id <id>` | Remove a schedule |

Cron expressions are 5-part and in UTC; run `dune schedule set --help` for the supported ones.

### `dune contract`

Submit smart contracts for decoding and track the submissions.

| Command | Description |
|---------|-------------|
| `contract submit --blockchain <chain> --address <addr> --project <name> --contract-name <name> [--abi <json>] [--proxy] [--factory] [--multiple-instances] [--manual-abi] [--type] [--reason] [--idempotency-key]` | Submit one contract |
| `contract submit --batch <json\|@path\|@->` | Submit up to 100 contracts; exits with an error if any is rejected |
| `contract submissions [--status] [--blockchain] [--address] [--project] [--contract-name] [--limit] [--cursor]` | List your submissions and their status |

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

## How the CLI talks to Dune

Commands call the Dune REST API through [duneapi-client-go](https://github.com/duneanalytics/duneapi-client-go). Where the REST API has no equivalent yet (name search for queries and dashboards, `dataset blockchains`, `dataset size`, `query create`, `query create-and-run`, `dataset search`, chart configs, `query update --temp`, and a few fields of `query get`, `viz get`, and `matview get`), the CLI calls the [Dune MCP server](https://docs.dune.com/docs/agents/mcp) at `<api host>/mcp/v1` with the same API key. The MCP tools `renderVisualization` (an interactive view for MCP apps) and `submit_feedback` (feedback about the MCP tools themselves) have no CLI command.
