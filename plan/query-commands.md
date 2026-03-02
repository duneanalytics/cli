# Dune CLI — `dune query` Implementation Plan

## Commands

| Command | Maps to MCP tool | SDK method |
|---------|-----------------|------------|
| `create` | `createDuneQuery` | `CreateQuery` (new — added to SDK in Step 2) |
| `get` | `getDuneQuery` | `GetQuery` (new — added to SDK in Step 2) |
| `update` | `updateDuneQuery` | `UpdateQuery` (new — added to SDK in Step 2) |
| `archive` | `updateDuneQuery` (is_archived) | `ArchiveQuery` (new — added to SDK in Step 2) |
| `run` | `executeQueryById` + `getExecutionResults` | `RunQuery` + `Execution.WaitGetResults` |
| `results` | `getExecutionResults` | `QueryResultsV2` |
| `run-sql` | (ad-hoc SQL) + `getExecutionResults` | `RunSQL` + `Execution.WaitGetResults` |

All commands use **only** the SDK's `dune.DuneClient` interface. No separate HTTP client in the CLI.

## Framework: Cobra + Charmbracelet Fang

- `github.com/spf13/cobra` — CLI framework (35k+ stars, industry standard)
- `github.com/charmbracelet/fang` — styled help pages, man pages, theming (wraps Cobra)
- Entry point: `fang.Execute(context.Background(), rootCmd)` instead of raw `rootCmd.Execute()`

## Guiding principle: Everything goes through the SDK

All Dune API interactions go through `github.com/duneanalytics/duneapi-client-go`. Missing functionality is added to the SDK itself — the CLI has **no custom HTTP calls**.

The SDK already provides: auth, config, HTTP utils, execution (`RunQuery`, `RunSQL`, `QueryExecute`, `SQLExecute`), results (`QueryResultsV2` with pagination), polling (`Execution.WaitGetResults`), models.

Always reuse existing structs from the SDK's `models/` package.

## Key dependency: duneapi-client-go

**Local development**: The CLI's `go.mod` uses a `replace` directive to point to the local SDK checkout while changes are in development:

```
replace github.com/duneanalytics/duneapi-client-go => ../duneapi-client-go
```

This allows `import "github.com/duneanalytics/duneapi-client-go/dune"` as normal — Go resolves from the local filesystem. Remove the `replace` line once the SDK changes are merged.

**Existing SDK methods used by the CLI (signatures updated in Step 2):**

| Method | Signature (after Step 2) | Used by |
|--------|--------------------------|---------|
| `RunQuery` | `(queryID int, params map[string]any, performance string) (Execution, error)` | `run` (wait mode) |
| `QueryExecute` | `(queryID int, params map[string]any, performance string) (*ExecuteResponse, error)` | `run` (no-wait mode) |
| `RunSQL` | `(sql string, performance string, params map[string]any) (Execution, error)` | `run-sql` |
| `SQLExecute` | `(sql string, performance string, params map[string]any) (*ExecuteResponse, error)` | (available) |
| `Execution.WaitGetResults` | `(pollInterval time.Duration, maxRetries int) (*ResultsResponse, error)` | `run`, `run-sql` |
| `QueryResultsV2` | `(executionID string, options ResultOptions) (*ResultsResponse, error)` | `results` |
| `QueryStatus` | `(executionID string) (*StatusResponse, error)` | (optional) |

**New SDK methods added in Step 2:**

| Method | Signature | Used by |
|--------|-----------|---------|
| `CreateQuery` | `(req CreateQueryRequest) (*CreateQueryResponse, error)` | `create` |
| `GetQuery` | `(queryID int) (*GetQueryResponse, error)` | `get` |
| `UpdateQuery` | `(queryID int, req UpdateQueryRequest) (*UpdateQueryResponse, error)` | `update` |
| `ArchiveQuery` | `(queryID int) (*UpdateQueryResponse, error)` | `archive` |

**Existing SDK models fixed in Step 2:**

| Model | Field added | Why |
|-------|-------------|-----|
| `ExecuteRequest` | `Performance string` | API accepts `performance` in execute body; needed for `--performance` flag |
| `ExecuteSQLRequest` | `QueryParameters map[string]any` | API accepts params in SQL execute body; needed for `--param` flag |

## Architecture: Single SDK client in context

```go
// Stored in Cobra command context, accessed via ClientFromCmd(cmd)
dune.DuneClient
```

One client, created from `*config.Env` in `PersistentPreRunE`. No wrapper structs needed.

---

## Step 1: Project Scaffolding + SDK Integration

- [x] Done

Add `github.com/spf13/cobra`, `github.com/charmbracelet/fang`, and `github.com/duneanalytics/duneapi-client-go` deps. Create root command (`cli/root.go`) with persistent `--api-key` flag (overrides `DUNE_API_KEY` env). Create `query` parent command (`cmd/query/query.go`). Use `fang.Execute(context.Background(), rootCmd)`.

**SDK integration:**
- Delete local `config/` package — use SDK's `config` package instead (identical API: `FromEnvVars()`, `FromAPIKey()`, `Env{APIKey, Host}`)
- Delete local `models/error.go` — use SDK error patterns
- In `PersistentPreRunE`: build `*config.Env` from SDK, create `dune.NewDuneClient(env)`, store in context
- Add `replace` directive to `go.mod` pointing to `../duneapi-client-go`
- Provide `cmdutil.ClientFromCmd(cmd) dune.DuneClient` helper

File structure: `cmd/main.go`, `cli/root.go`, `cmdutil/client.go`, `cmd/query/query.go`.

Reuses: `config.Env`, `config.FromEnvVars()`, `config.FromAPIKey()`, `dune.NewDuneClient(env)`.

**Acceptance criteria:**
- `dune --help` lists `query` as subcommand with Fang-styled help
- `dune query --help` lists available subcommands
- Missing API key prints error to stderr, exits 1
- `make build` produces binary
- No local `config/` or `models/` packages remain
- `go vet ./...` passes

**Tests:**
- Root command initializes without error
- Missing API key returns error
- `ClientFromCmd` returns non-nil DuneClient when API key is set
- Query command registered as subcommand

---

## Step 2: Add Query CRUD to SDK

- [ ] Done

**This step modifies the SDK repo** at `/Users/ivpusic/github/dune/duneapi-client-go`, not the CLI.

Add query CRUD endpoints to the `DuneClient` interface and fix incomplete execution models. Verified against API server source (`duneapi/models/querycrud.go`) and docs (`docs.dune.com/api-reference/queries`).

### Updated file: `models/execute.go` — fix incomplete models

```go
type ExecuteRequest struct {
    QueryParameters map[string]any `json:"query_parameters,omitempty"`
    Performance     string         `json:"performance,omitempty"` // NEW — "medium" or "large"
}

type ExecuteSQLRequest struct {
    SQL             string         `json:"sql"`
    Performance     string         `json:"performance,omitempty"`
    QueryParameters map[string]any `json:"query_parameters,omitempty"` // NEW — parameterized SQL
}
```

Also update `SQLExecute` and `RunSQL` signatures to accept `queryParameters`:

```go
// Current: SQLExecute(sql string, performance string)
// Updated: SQLExecute(sql string, performance string, queryParameters map[string]any)
// Current: RunSQL(sql string, performance string)
// Updated: RunSQL(sql string, performance string, queryParameters map[string]any)
```

And update `QueryExecute` to pass `Performance` through:

```go
// The existing QueryExecute already takes queryParameters map[string]any.
// Just need to populate ExecuteRequest.Performance from the request.
// Add performance parameter to signature:
// Current: QueryExecute(queryID int, queryParameters map[string]any)
// Updated: QueryExecute(queryID int, queryParameters map[string]any, performance string)
// Similarly for RunQuery:
// Current: RunQuery(queryID int, queryParameters map[string]any)
// Updated: RunQuery(queryID int, queryParameters map[string]any, performance string)
```

### New file: `models/query.go`

Types match the Dune API spec (reference: `duneapi/models/querycrud.go`, docs: `docs.dune.com/api-reference/queries`).

```go
// QueryParameter represents a parameterized query variable.
// Supported types: "text", "number", "datetime", "enum".
type QueryParameter struct {
    Key         string   `json:"key"`
    Type        string   `json:"type"`
    Value       string   `json:"value"`
    EnumOptions []string `json:"enumOptions,omitempty"`
}

// POST /api/v1/query
type CreateQueryRequest struct {
    Name        string           `json:"name"`
    QuerySQL    string           `json:"query_sql"`
    Description string           `json:"description,omitempty"`
    IsPrivate   bool             `json:"is_private,omitempty"`
    Parameters  []QueryParameter `json:"parameters,omitempty"`
    Tags        []string         `json:"tags,omitempty"`
}

type CreateQueryResponse struct {
    QueryID int `json:"query_id"`
}

// GET /api/v1/query/{queryId}
type GetQueryResponse struct {
    QueryID     int              `json:"query_id"`
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Tags        []string         `json:"tags"`
    Version     int              `json:"version"`
    Parameters  []QueryParameter `json:"parameters"`
    QueryEngine string           `json:"query_engine"`
    QuerySQL    string           `json:"query_sql"`
    IsPrivate   bool             `json:"is_private"`
    IsArchived  bool             `json:"is_archived"`
    IsUnsaved   bool             `json:"is_unsaved"`
    Owner       string           `json:"owner"`
}

// PATCH /api/v1/query/{queryId}
// Pointer fields with omitempty — only non-nil fields are serialized.
type UpdateQueryRequest struct {
    Name        *string           `json:"name,omitempty"`
    Description *string           `json:"description,omitempty"`
    QuerySQL    *string           `json:"query_sql,omitempty"`
    Tags        *[]string         `json:"tags,omitempty"`
    Parameters  *[]QueryParameter `json:"parameters,omitempty"`
    IsPrivate   *bool             `json:"is_private,omitempty"`
    IsArchived  *bool             `json:"is_archived,omitempty"`
}

type UpdateQueryResponse struct {
    QueryID int `json:"query_id"`
}
```

### New file: `dune/query.go`

```go
// POST /api/v1/query
func (c *duneClient) CreateQuery(req models.CreateQueryRequest) (*models.CreateQueryResponse, error)

// GET /api/v1/query/{queryId}
func (c *duneClient) GetQuery(queryID int) (*models.GetQueryResponse, error)

// PATCH /api/v1/query/{queryId}
func (c *duneClient) UpdateQuery(queryID int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error)

// POST /api/v1/query/{queryId}/archive
func (c *duneClient) ArchiveQuery(queryID int) (*models.UpdateQueryResponse, error)
```

Uses existing `httpRequest()` and `decodeBody()` helpers from `dune/http.go`. Follows the same pattern as `QueryExecute` / `SQLExecute`.

### Updated file: `dune/dune.go`

Add to `DuneClient` interface:

```go
// Query CRUD
CreateQuery(req models.CreateQueryRequest) (*models.CreateQueryResponse, error)
GetQuery(queryID int) (*models.GetQueryResponse, error)
UpdateQuery(queryID int, req models.UpdateQueryRequest) (*models.UpdateQueryResponse, error)
ArchiveQuery(queryID int) (*models.UpdateQueryResponse, error)
```

Update existing method signatures:

```go
// Updated signatures (add performance/params where missing):
QueryExecute(queryID int, queryParameters map[string]any, performance string) (*models.ExecuteResponse, error)
RunQuery(queryID int, queryParameters map[string]any, performance string) (Execution, error)
RunQueryGetRows(queryID int, queryParameters map[string]any, performance string) ([]map[string]any, error)
SQLExecute(sql string, performance string, queryParameters map[string]any) (*models.ExecuteResponse, error)
RunSQL(sql string, performance string, queryParameters map[string]any) (Execution, error)
```

Add URL templates:

```go
createQueryURLTemplate  = "%s/api/v1/query"
queryURLTemplate        = "%s/api/v1/query/%d"          // GET + PATCH
archiveQueryURLTemplate = "%s/api/v1/query/%d/archive"  // POST
```

**Acceptance criteria:**
- `CreateQuery` POSTs to `/api/v1/query` with JSON body containing name, query_sql, description, is_private, parameters, tags
- `GetQuery` GETs `/api/v1/query/{id}`, parses full response including version, query_engine, is_unsaved, owner
- `UpdateQuery` PATCHes `/api/v1/query/{id}` with only non-nil fields in body
- `ArchiveQuery` POSTs to `/api/v1/query/{id}/archive` with empty body
- `ExecuteRequest` now includes `Performance` field in JSON body
- `ExecuteSQLRequest` now includes `QueryParameters` field in JSON body
- `QueryExecute` / `RunQuery` accept and pass `performance` parameter
- `SQLExecute` / `RunSQL` accept and pass `queryParameters` parameter
- All methods use existing `httpRequest` helper (sets `X-DUNE-API-KEY` header)
- Non-2xx responses follow existing SDK error pattern (`ErrorReqUnsuccessful`)
- Existing SDK tests still pass (`go test ./...`)

**Tests (new file `dune/query_test.go`, using httptest):**
- CreateQuery: verify POST method, path `/api/v1/query`, request body fields; parse `query_id` response
- GetQuery: verify GET method, path `/api/v1/query/123`; parse all response fields
- UpdateQuery: verify PATCH method, path `/api/v1/query/123`; body omits nil fields, includes non-nil fields
- ArchiveQuery: verify POST method, path `/api/v1/query/123/archive`; parse `query_id` response
- Error responses (400, 401, 404) return `ErrorReqUnsuccessful`

**Tests (update existing execution tests):**
- `QueryExecute` with performance: verify `performance` field in request body
- `SQLExecute` with queryParameters: verify `query_parameters` field in request body

---

## Step 3: Output Formatting

- [x] Deferred — create `output/` inline when the first command needs it (Step 4).

---

## Step 4: `dune query create`

- [x] Done

`cmd/query/create.go` — flags: `--name` (required), `--sql` (required), `--description`, `--private`, `-o`. Gets client via `cmdutil.ClientFromCmd(cmd)`, calls `client.CreateQuery(models.CreateQueryRequest{...})`.

API reference: POST `/api/v1/query` — name (max 600 chars), query_sql (max 500k chars), description (max 1k chars), is_private, parameters, tags → `{"query_id": int}`.

**Output:** text: `Created query 4125432` / json: `{"query_id": 4125432}`

**Acceptance criteria:**
- `dune query create --name "Test" --sql "SELECT 1"` creates query, prints ID
- `--private` sets is_private=true
- Missing `--name` or `--sql` errors
- API error printed to stderr, exits 1
- `-o json` works

**Tests:**
- Required flags validation
- Successful create prints ID (mock DuneClient)
- Private flag passed correctly
- JSON output format

---

## Step 5: `dune query get`

- [ ] Done

`cmd/query/get.go` — positional arg: query ID (required, integer). Flag: `-o`. Calls `client.GetQuery(queryID)`.

API reference: GET `/api/v1/query/{queryId}` → query_id, name, description, query_sql, owner, is_private, is_archived, is_unsaved, version, query_engine, tags, parameters.

**Output:** text: key-value with SQL block / json: full response.

**Acceptance criteria:**
- `dune query get 4125432` displays metadata + SQL
- Missing/non-integer ID errors
- 404 prints clear error
- `-o json` outputs full response

**Tests:**
- Valid ID renders text output (mock DuneClient)
- JSON output matches response
- Missing argument errors
- Non-integer argument errors
- 404 handled

---

## Step 6: `dune query update`

- [ ] Done

`cmd/query/update.go` — positional arg: query ID. Flags: `--name`, `--sql`, `--description`, `--private`, `--tags` — all optional but at least one required. Only sends provided fields (pointer/omitempty pattern). Calls `client.UpdateQuery(queryID, models.UpdateQueryRequest{...})`.

API reference: PATCH `/api/v1/query/{queryId}` — name, query_sql, description, parameters, tags, is_private, is_archived (all optional) → `{"query_id": int}`.

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

`cmd/query/archive.go` — positional arg: query ID. Calls `client.ArchiveQuery(queryID)`.

API reference: POST `/api/v1/query/{queryId}/archive` — dedicated endpoint, no request body → `{"query_id": int}`.

**Output:** `Archived query 4125432`

**Acceptance criteria:**
- Sends POST to `/api/v1/query/{id}/archive`
- Missing ID errors
- API errors handled

**Tests:**
- Correct HTTP method and path (mock DuneClient)
- Missing argument errors
- 404 handled

---

## Step 8: `dune query run`

- [ ] Done

`cmd/query/run.go` — positional arg: query ID. Flags: `--param key=value` (repeatable), `--performance medium|large`, `--limit`, `--no-wait`, `-o`.

**SDK-first approach — no custom polling logic:**

- `--no-wait` mode: calls `client.QueryExecute(queryID, params, performance)`, prints execution ID, exits
- Wait mode (default): calls `client.RunQuery(queryID, params, performance)` → `exec.WaitGetResults(pollInterval, maxRetries)` — SDK handles all polling internally
- `--performance` flag: passed directly to SDK methods (Step 2 adds `Performance` field to `ExecuteRequest`)

API reference: POST `/api/v1/query/{query_id}/execute` — body: `{"query_parameters": {...}, "performance": "medium"|"large"}`. Response: `{"execution_id": string, "state": string}`.

No `poll.go` needed — the SDK's `Execution.WaitGetResults()` replaces all custom polling logic.

Reuses: SDK's `RunQuery`, `Execution.WaitGetResults`, `QueryExecute`, `ResultsResponse`.

**Output:** `--no-wait`: `Execution ID: 01JG...` / table: rows + footer with row count / json: full result object / csv: standard CSV.

**Acceptance criteria:**
- Executes and prints results as table
- `--param` flags parsed and passed as `map[string]any`
- `--performance large` passed to SDK's `QueryExecute`/`RunQuery`
- `--limit` limits displayed rows
- `--no-wait` prints execution ID only
- Failed execution prints error, exits 1
- Progress shown on stderr during polling (SDK handles this)
- `-o json` and `-o csv` work

**Tests:**
- Param parsing ("key=value" → map)
- Performance flag passed to SDK methods
- No-wait mode returns execution ID
- Successful execution renders table (mock DuneClient interface)
- Failed execution prints error, exits 1
- JSON and CSV output formats

---

## Step 9: `dune query results`

- [ ] Done

`cmd/query/results.go` — positional arg: execution ID (string). Flags: `--limit`, `--offset`, `-o`.

One-shot fetch via `client.QueryResultsV2(executionID, models.ResultOptions{Page: &models.ResultPageOption{Offset, Limit}})` — no polling. If still running: print status, exit 0. If complete: display results. If failed: print error, exit 1.

API reference: GET `/api/v1/execution/{execution_id}/results` — query params: limit, offset → state, result metadata, data rows. States: `QUERY_STATE_COMPLETED`, `QUERY_STATE_PENDING`, `QUERY_STATE_EXECUTING`, `QUERY_STATE_FAILED`, `QUERY_STATE_CANCELLED`, `QUERY_STATE_EXPIRED`.

Reuses: SDK's `QueryResultsV2`, `models.ResultOptions`, `models.ResultPageOption`, `models.ResultsResponse`.

**Acceptance criteria:**
- Completed execution displays results
- `--limit` and `--offset` work
- Running execution prints status, exits 0
- Failed execution prints error, exits 1
- `-o json` and `-o csv` work

**Tests:**
- Completed execution renders results (mock DuneClient)
- Running execution prints status
- Failed execution prints error
- Offset and limit passed correctly
- Missing argument errors

---

## Step 10: `dune query run-sql`

- [ ] Done

`cmd/query/run_sql.go` — flags: `--sql` (required), `--performance medium|large`, `--limit`, `--param key=value`, `-o`.

Calls `client.RunSQL(sql, performance, params)` which returns an `Execution` interface, then `exec.WaitGetResults(pollInterval, maxRetries)`. Fully SDK-driven — the SDK accepts `performance` and `queryParameters` (Step 2 fixes `ExecuteSQLRequest` to include both).

API reference: POST `/api/v1/sql/execute` — body: `{"sql": string, "performance": string, "query_parameters": {...}}` → `{"execution_id": string, "state": string}`.

Reuses: SDK's `RunSQL`, `Execution.WaitGetResults`, `ResultsResponse`.

**Acceptance criteria:**
- `dune query run-sql --sql "SELECT 1"` executes and prints results
- `--performance large` passed to SDK's `RunSQL`
- `--param key=value` passed as `queryParameters` to SDK's `RunSQL`
- `--limit` limits displayed rows
- Missing `--sql` errors
- SQL syntax error prints error with hint
- Progress shown on stderr

**Tests:**
- Successful execution and display (mock DuneClient)
- Missing --sql flag errors
- Performance flag passed to `RunSQL`
- Param flags passed as queryParameters to `RunSQL`
- SQL error prints details
- Uses SDK Execution for polling (same pattern as `run`)

---

## File Structure

```
cli/                                    # CLI repo
  cli/
    root.go                             # Step 1: root command, Execute()
  cmd/
    main.go                             # Entry point (exists)
    query/
      query.go                          # Query parent command (exists)
      create.go                         # Step 4
      get.go                            # Step 5
      update.go                         # Step 6
      archive.go                        # Step 7
      run.go                            # Step 8
      results.go                        # Step 9
      run_sql.go                        # Step 10
  cmdutil/
    client.go                           # SetClient, ClientFromCmd (context helpers)
  output/
    output.go                           # Shared output formatting (text, JSON, CSV)
  go.mod                                # Has replace directive → ../duneapi-client-go
  plan/
    query-commands.md                   # This plan

duneapi-client-go/                      # SDK repo (separate)
  models/
    query.go                            # Step 2: new — query CRUD types
    execute.go                          # Step 2: updated — add Performance to ExecuteRequest,
                                        #   QueryParameters to ExecuteSQLRequest
  dune/
    dune.go                             # Step 2: updated — 4 new methods + updated signatures
    query.go                            # Step 2: new — CreateQuery, GetQuery, UpdateQuery, ArchiveQuery
    query_test.go                       # Step 2: new — tests
```

## Dependency Graph

```
Step 1 (scaffolding + SDK integration + replace directive)
  ├── Step 2 (add query CRUD to SDK — separate repo)
  │     └── Steps 4-7 (CRUD commands — need Step 2)
  ├── Steps 8-9 (execution commands — need Step 1, SDK already has methods)
  └── Step 10 (run-sql — need Step 1, SDK already has RunSQL)
```

Output formatting (`output/`) is created inline with the first command that needs it.
Steps 4-7 depend on Step 2. Steps 8-10 only need Step 1 (SDK already has execution methods).
