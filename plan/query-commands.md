# Dune CLI — `dune query` Implementation Plan

## Commands

| Command | Maps to MCP tool | Reuses duneapi-client-go |
|---------|-----------------|--------------------------|
| `create` | `createDuneQuery` | No (needs new API calls) |
| `get` | `getDuneQuery` | No (needs new API calls) |
| `update` | `updateDuneQuery` | No (needs new API calls) |
| `archive` | `updateDuneQuery` (is_archived) | No (needs new API calls) |
| `run` | `executeQueryById` + `getExecutionResults` | Yes: `QueryExecute`, `QueryResultsV2` |
| `results` | `getExecutionResults` | Yes: `QueryResultsV2` |
| `run-sql` | (ad-hoc SQL) + `getExecutionResults` | Yes: `SQLExecute`, `QueryResultsV2` |

## Framework: Cobra + Charmbracelet Fang

- `github.com/spf13/cobra` — CLI framework (35k+ stars, industry standard)
- `github.com/charmbracelet/fang` — styled help pages, man pages, theming (wraps Cobra)
- Entry point: `fang.Execute(context.Background(), rootCmd)` instead of raw `rootCmd.Execute()`

## Guiding principle: Prefer SDK structs

Always prefer reusing existing structs and types from `duneapi-client-go` over creating new ones. Only define new types in the CLI when the SDK does not already provide them (e.g. query CRUD request/response types). Before adding a new struct to `api/models.go`, check whether a suitable type already exists in the SDK's `models/` package.

## Key dependency: duneapi-client-go

Provides: auth (`X-DUNE-API-KEY` header), config (`DUNE_API_KEY`/`DUNE_API_HOST` env vars), HTTP utils, execution (`QueryExecute`, `SQLExecute`, `QueryStatus`), results (`QueryResultsV2` with pagination), models (`ExecuteResponse`, `ResultsResponse`, `StatusResponse`).

**Gap**: No query CRUD endpoints (create/get/update). The CLI needs an internal `api/` package for these — but only for types not already in the SDK.

---

## Step 1: Project Scaffolding + Cobra + Fang

- [ ] Done

Add `github.com/spf13/cobra`, `github.com/charmbracelet/fang`, and `github.com/duneanalytics/duneapi-client-go` deps. Create root command (`cmd/root.go`) with persistent `--api-key` flag (overrides `DUNE_API_KEY` env). Create `query` parent command (`cmd/query/query.go`). Refactor `cmd/main.go` to use `fang.Execute(context.Background(), rootCmd)`.

File structure: `cmd/main.go`, `cmd/root.go`, `cmd/query/query.go`.

MCP reference: None — infrastructure step.

Reuses: `config.Env`, `config.FromEnvVars()`, `dune.NewClient(env)`.

**Acceptance criteria:**
- `dune --help` lists `query` as subcommand with Fang-styled help
- `dune query --help` lists available subcommands
- Missing API key prints error to stderr, exits 1
- `make build` produces binary

**Tests:**
- Root command initializes without error
- Missing API key returns error
- Query command registered as subcommand

---

## Step 2: Query API Client (CRUD)

- [ ] Done

Create `api/client.go` — thin HTTP client reusing duneapi-client-go's config/auth. Generic `doRequest(method, path, body)` helper, sets `X-DUNE-API-KEY` header.

Create `api/query.go` with:
- `CreateQuery(req) → (query_id, error)` — POST `/api/v1/query`
- `GetQuery(queryID) → (*QueryResponse, error)` — GET `/api/v1/query/{id}`
- `UpdateQuery(queryID, req) → error` — PATCH `/api/v1/query/{id}`

Create `api/models.go` with request/response types **only for types not already in duneapi-client-go**. Before defining a new struct, check the SDK's `models/` package for an existing match. `UpdateQueryRequest` uses pointer fields so only provided fields are serialized (`*string`, `*bool` with `omitempty`).

MCP reference: `createDuneQuery` (POST, name+query_sql+description+is_private+parameters → query_id), `getDuneQuery` (GET → full query object), `updateDuneQuery` (PATCH, only changed fields).

Reuses: `config.Env` for host+key, HTTP header pattern from `dune/http.go`, error model from `models/error.go`. Reuse any existing SDK structs (e.g. `ExecuteResponse`, `ResultsResponse`, `StatusResponse`, error types) wherever applicable.

**Acceptance criteria:**
- CreateQuery sends correct POST body, returns query_id
- GetQuery sends GET to correct path, parses full response
- UpdateQuery sends PATCH with only non-nil fields
- All methods set `X-DUNE-API-KEY` header
- API errors (4xx/5xx) returned as structured errors

**Tests (httptest):**
- CreateQuery: verify request body and parse response
- GetQuery: verify URL path and parse response
- UpdateQuery: verify PATCH body omits nil fields
- Error responses (400, 401, 404, 500) return structured error

---

## Step 3: Output Formatting

- [ ] Done

Create `output/` package. `AddOutputFlag(cmd, default)` adds `-o/--output` flag.

- `PrintJSON(w, v)` — indented JSON
- `PrintKeyValue(w, pairs)` — aligned `Key: Value` for single objects
- `PrintTable(w, columns, rows)` — column-aligned text table
- `PrintCSV(w, columns, rows)` — stdlib `encoding/csv`

All write to `io.Writer` (no direct stdout coupling).

MCP reference: mirrors output shapes — `getExecutionResults` → table/csv, `getDuneQuery` → text/json, `createDuneQuery` → text/json.

**Acceptance criteria:**
- PrintJSON outputs valid indented JSON
- PrintKeyValue renders aligned pairs
- PrintTable renders aligned columns with header
- PrintCSV outputs valid CSV with header

**Tests:**
- JSON: marshal struct, verify valid output
- Text: key-value alignment
- Table: columns and rows aligned, handles empty
- CSV: valid output, handles commas/quotes in values

---

## Step 4: `dune query create`

- [ ] Done

`cmd/query/create.go` — flags: `--name` (required), `--sql` (required), `--description`, `--private`, `-o`. Calls `api.CreateQuery()`. CLI sets `is_temp: false` (unlike MCP which defaults temp).

MCP reference: `createDuneQuery` — name (max 600 chars), query (max 500k chars), description (max 1k chars), is_private, parameters → query_id.

**Output:** text: `Created query 4125432` / json: `{"query_id": 4125432}`

**Acceptance criteria:**
- `dune query create --name "Test" --sql "SELECT 1"` creates query, prints ID
- `--private` sets is_private=true
- Missing `--name` or `--sql` errors
- API error printed to stderr, exits 1
- `-o json` works

**Tests:**
- Required flags validation
- Successful create prints ID (mock API)
- Private flag passed correctly
- JSON output format

---

## Step 5: `dune query get`

- [ ] Done

`cmd/query/get.go` — positional arg: query ID (required, integer). Flag: `-o`. Calls `api.GetQuery()`.

MCP reference: `getDuneQuery` — query_id → query_id, name, query_sql, description, is_private, is_archived, tags, owner, parameters.

**Output:** text: key-value with SQL block / json: full response.

**Acceptance criteria:**
- `dune query get 4125432` displays metadata + SQL
- Missing/non-integer ID errors
- 404 prints clear error
- `-o json` outputs full response

**Tests:**
- Valid ID renders text output (mock API)
- JSON output matches response
- Missing argument errors
- Non-integer argument errors
- 404 handled

---

## Step 6: `dune query update`

- [ ] Done

`cmd/query/update.go` — positional arg: query ID. Flags: `--name`, `--sql`, `--description`, `--private` — all optional but at least one required. Only sends provided fields (pointer/omitempty pattern).

MCP reference: `updateDuneQuery` — PATCH with queryId + optional fields. Optimistic locking.

**Output:** text: `Updated query 4125432` / json: `{"query_id": 4125432}`

**Acceptance criteria:**
- Single flag update sends only that field
- Multiple flags all included
- No flags prints error
- API errors (404, 409) printed clearly

**Tests:**
- Single flag → only that field in PATCH body
- Multiple flags all present
- No flags → usage error
- 404 handled

---

## Step 7: `dune query archive`

- [ ] Done

`cmd/query/archive.go` — positional arg: query ID. Calls `api.UpdateQuery(id, {IsArchived: true})`.

MCP reference: `updateDuneQuery` with is_archived=true.

**Output:** `Archived query 4125432`

**Acceptance criteria:**
- Sends PATCH with `is_archived: true`
- Missing ID errors
- API errors handled

**Tests:**
- Correct PATCH body
- Missing argument errors
- 404 handled

---

## Step 8: `dune query run`

- [ ] Done

`cmd/query/run.go` — positional arg: query ID. Flags: `--param key=value` (repeatable), `--performance medium|large`, `--limit`, `--no-wait`, `-o`.

Calls `QueryExecute(queryID, params)`. If `--no-wait`: print execution ID, exit. Otherwise: poll `QueryResultsV2(executionID, options)` every 2s. Print progress to stderr. On complete: display results. On fail: print error with line/column hint, exit 1.

Extract polling + display into shared helper (`internal/poll.go`) — reused by `run-sql`.

MCP reference: `executeQueryById` (query_id, performance, query_parameters → execution_id, state) + `getExecutionResults` (executionId, limit, offset, timeout → state, resultMetadata, data rows. Error states: FAILED with errorMessage/errorMetadata, CANCELLED, EXPIRED).

Reuses: `QueryExecute`, `QueryResultsV2`, `ResultOptions`, `IsExecutionFinished`, execution polling pattern.

**Output:** `--no-wait`: `Execution ID: 01JG...` / table: rows + footer with row count and credits / json: full result object / csv: standard CSV.

**Acceptance criteria:**
- Executes and prints results as table
- `--param` flags parsed and passed
- `--performance large` passed to API
- `--limit` limits rows
- `--no-wait` prints execution ID only
- Failed execution prints error, exits 1
- Progress shown on stderr during polling
- `-o json` and `-o csv` work
- Credits shown in table footer

**Tests:**
- Param parsing ("key=value" → map)
- No-wait mode returns execution ID
- Successful execution renders table (mock API)
- Failed execution prints error, exits 1
- JSON and CSV output formats

---

## Step 9: `dune query results`

- [ ] Done

`cmd/query/results.go` — positional arg: execution ID (string). Flags: `--limit`, `--offset`, `-o`.

One-shot fetch via `QueryResultsV2(executionID, options)` — no polling. If still running: print status, exit 0. If complete: display results. If failed: print error, exit 1.

MCP reference: `getExecutionResults` — executionId, limit (1-100), offset → state, resultMetadata, data rows.

Reuses: `QueryResultsV2`, `ResultOptions{Page: &ResultPageOption{Offset, Limit}}`.

**Acceptance criteria:**
- Completed execution displays results
- `--limit` and `--offset` work
- Running execution prints status, exits 0
- Failed execution prints error, exits 1
- `-o json` and `-o csv` work

**Tests:**
- Completed execution renders results (mock API)
- Running execution prints status
- Failed execution prints error
- Offset and limit passed correctly
- Missing argument errors

---

## Step 10: `dune query run-sql`

- [ ] Done

`cmd/query/run_sql.go` — flags: `--sql` (required), `--name` (optional), `--performance`, `--limit`, `--param key=value`, `-o`.

Calls `SQLExecute(sql, performance)` → polls results using shared helper from step 8.

MCP reference: combines `createDuneQuery` (temp) + `executeQueryById` + `getExecutionResults`. But since duneapi-client-go has `SQLExecute` hitting `/api/v1/sql/execute`, we skip the create step.

Reuses: `SQLExecute(sql, performance)` → `*ExecuteResponse`, shared polling helper, `QueryResultsV2`.

**Acceptance criteria:**
- `dune query run-sql --sql "SELECT 1"` executes and prints results
- `--performance large` passed to API
- `--limit` limits rows
- Missing `--sql` errors
- SQL syntax error prints error with hint
- Progress shown on stderr

**Tests:**
- Successful execution and display (mock API)
- Missing --sql flag errors
- Performance flag passed
- SQL error prints details
- Uses shared polling logic (no duplication with run)
