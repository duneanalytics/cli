# Dune CLI

A command-line interface for interacting with the Dune Analytics API.

## Authentication

```bash
# Save your API key to ~/.config/dune/config.yaml
dune auth --api-key <key>

# Or set via environment variable
export DUNE_API_KEY=<key>
```

The `--api-key` flag is available on all commands to override the stored key.

## Commands

### `dune query`

Manage and execute Dune queries.

| Command | Description |
|---------|-------------|
| `query create --name <name> --sql <sql> [--description] [--private]` | Create a new saved query |
| `query get <query-id>` | Get a saved query's details and SQL |
| `query update <query-id> [--name] [--sql] [--description] [--private] [--tags]` | Update an existing query |
| `query archive <query-id>` | Archive a saved query |
| `query run <query-id> [--param key=value] [--performance medium\|large] [--limit] [--no-wait]` | Execute a saved query and display results |
| `query run-sql --sql <sql> [--param key=value] [--performance medium\|large] [--limit] [--no-wait]` | Execute raw SQL directly |

### `dune execution`

Manage query executions.

| Command | Description |
|---------|-------------|
| `execution results <execution-id> [--limit] [--offset]` | Fetch results of a query execution |

### `dune dataset`

Search the Dune dataset catalog.

| Command | Description |
|---------|-------------|
| `dataset search [--query] [--categories] [--blockchains] [--schemas] [--dataset-types] [--owner-scope] [--include-private] [--include-schema] [--include-metadata] [--limit] [--offset]` | Search for datasets |

Categories: `canonical`, `decoded`, `spell`, `community`

### `dune usage`

Show credit and resource usage for your account.

```bash
dune usage [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]
```

## Output Format

Most commands support `-o, --output <format>` with `text` (default) or `json`.
