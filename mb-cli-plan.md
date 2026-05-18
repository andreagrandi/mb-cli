# mb-cli: Metabase CLI Tool - Implementation Plan

## Context

Build a **read-only** Go CLI for the Metabase API focused on database exploration and data querying. Auth via `MB_HOST` + `MB_API_KEY` env vars. Primary use case: coding agents (Claude, Codex) querying Metabase databases to verify data, inspect schemas, and run ad-hoc queries. The tool mirrors the project structure and conventions of `logbasset` and `sentire`.

**Constraint: This tool is strictly read-only. No data modification endpoints will be implemented.**

---

## API Endpoints Inventory

All endpoints used by this CLI. Despite some being POST, they are all **read-only** operations (Metabase uses POST for query execution).

### Database (all GET)
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| GET | `/api/database/` | `mb-cli database list` | List all databases |
| GET | `/api/database/{id}` | `mb-cli database get <id>` | Get database details |
| GET | `/api/database/{id}/metadata` | `mb-cli database metadata <id>` | Full metadata (tables + fields) |
| GET | `/api/database/{id}/fields` | `mb-cli database fields <id>` | List all fields in database |
| GET | `/api/database/{id}/schemas` | `mb-cli database schemas <id>` | List schema names |
| GET | `/api/database/{id}/schema/{schema}` | `mb-cli database schema <id> <schema>` | Tables in a specific schema |

### Table (all GET)
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| GET | `/api/table/` | `mb-cli table list` | List all tables |
| GET | `/api/table/{id}` | `mb-cli table get <id>` | Get table details |
| GET | `/api/table/{id}/query_metadata` | `mb-cli table metadata <id>` | Table metadata with fields |
| GET | `/api/table/{id}/fks` | `mb-cli table fks <id>` | Foreign key relationships |
| GET | `/api/table/{table-id}/data` | `mb-cli table data <id>` | Get raw table data |

### Field (all GET)
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| GET | `/api/field/{id}` | `mb-cli field get <id>` | Get field details |
| GET | `/api/field/{id}/summary` | `mb-cli field summary <id>` | Field summary statistics |
| GET | `/api/field/{id}/values` | `mb-cli field values <id>` | Distinct values for a field |

### Dataset / Query (POST but read-only)
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| POST | `/api/dataset/` | `mb-cli query sql --db <id-or-name> --sql "..."` | Run native SQL query |
| POST | `/api/dataset/{export-format}` | `mb-cli query sql --db <id> --sql "..." --export csv` | Export query results |

### Card / Saved Questions
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| GET | `/api/card/` | `mb-cli card list` | List saved questions |
| GET | `/api/card/{id}` | `mb-cli card get <id>` | Get card details |
| POST | `/api/card/{card-id}/query` | `mb-cli card run <id>` | Execute a saved question |

### Search (GET)
| Method | Endpoint | CLI Command | Purpose |
|--------|----------|-------------|---------|
| GET | `/api/search/` | `mb-cli search <query>` | Search across all Metabase items |

---

## Project Structure

```
mb-cli/
├── cmd/mb-cli/main.go                      # Entry point
├── internal/
│   ├── cli/                            # Cobra commands
│   │   ├── root.go                     # Root command, global flags
│   │   ├── database.go                 # database subcommands
│   │   ├── table.go                    # table subcommands
│   │   ├── field.go                    # field subcommands
│   │   ├── query.go                    # query sql command
│   │   ├── card.go                     # card subcommands
│   │   ├── search.go                   # search command
│   │   ├── version.go                  # version command
│   │   ├── context.go                  # agent context document (Step 13)
│   │   ├── context_embed.md            # embedded agent reference (Step 13)
│   │   ├── schema.go                   # JSON schema introspection (Step 13)
│   │   └── tty.go                      # TTY auto-detection (Step 13)
│   ├── client/                         # Metabase HTTP client
│   │   ├── client.go                   # Client struct, Do/Get/Post, x-api-key auth
│   │   ├── databases.go               # Database API methods
│   │   ├── tables.go                   # Table API methods
│   │   ├── fields.go                   # Field API methods
│   │   ├── dataset.go                  # Dataset/query API methods
│   │   ├── cards.go                    # Card API methods
│   │   ├── search.go                   # Search API method
│   │   └── types.go                    # Request/response structs
│   ├── config/                         # Configuration
│   │   └── config.go                   # LoadConfig from MB_HOST + MB_API_KEY env vars
│   ├── validation/                     # Input hardening (Step 13)
│   │   └── validation.go              # Control char rejection, SQL validation
│   ├── formatter/                      # Output formatting
│   │   ├── formatter.go               # Formatter interface + factory
│   │   ├── json.go                    # JSON formatter (pretty-printed)
│   │   └── table.go                   # Table formatter (text/tabwriter)
│   └── version/
│       └── version.go                  # Version var (injected via ldflags)
├── tests/                              # Test files (sentire convention)
│   ├── client_test.go
│   ├── config_test.go
│   ├── databases_test.go
│   ├── tables_test.go
│   ├── fields_test.go
│   ├── dataset_test.go
│   ├── cards_test.go
│   ├── search_test.go
│   ├── formatter_test.go
│   ├── context_test.go                 # (Step 13)
│   ├── schema_test.go                  # (Step 13)
│   ├── tty_test.go                     # (Step 13)
│   ├── error_format_test.go            # (Step 13)
│   ├── validation_test.go              # (Step 13)
│   └── fields_filter_test.go           # (Step 13)
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                      # CI: test, vet, fmt, build
│   │   └── release.yml                 # Release: tag-triggered goreleaser
│   └── dependabot.yml                  # Weekly gomod + github-actions updates
├── Makefile
├── .goreleaser.yaml
├── go.mod
├── AGENTS.md
├── CHANGELOG.md
├── LICENSE                             # (exists)
└── README.md                           # (exists, will update)
```

---

## Implementation Steps

> **Progress: 13/14 steps completed**
> ✅ = done, ⬚ = not started

### Step 1: Project Scaffolding + Config + Version ✅

Set up Go module, entry point, config loading, root CLI, and version command.

**Files to create:**
- `go.mod` — module `github.com/andreagrandi/mb-cli`, Go 1.25
- `cmd/mb-cli/main.go` — calls `cli.Execute()` (pattern from sentire `cmd/sentire/main.go`)
- `internal/version/version.go` — `var Version = "dev"` (injected via ldflags at build)
- `internal/config/config.go` — `LoadConfig()` reads `MB_HOST` and `MB_API_KEY` from env vars; returns error if either is missing (pattern from sentire `internal/config/config.go`)
- `internal/cli/root.go` — Cobra root command with persistent flags: `--format` (json/table, default json), `--verbose`
- `internal/cli/version.go` — `version` subcommand printing version string
- `Makefile` — targets: `build`, `test`, `test-verbose`, `fmt`, `vet`, `lint`, `clean`, `deps`, `build-all`, `help`
- `AGENTS.md` — development conventions (adapted from logbasset/sentire AGENTS.md)

**Dependencies:** `github.com/spf13/cobra`

**Tests:** `tests/config_test.go` — env var loading, missing MB_HOST error, missing MB_API_KEY error

**Verification:** `make build && ./bin/mb-cli version`

---

### Step 2: HTTP Client ✅

Build the core Metabase API client with API key authentication.

**Files to create:**
- `internal/client/client.go`:
  - `Client` struct: `BaseURL string`, `HTTPClient *http.Client`, `APIKey string`, `Verbose bool`
  - `NewClient(cfg *config.Config) *Client` constructor (30s timeout)
  - `Do(req *http.Request) (*http.Response, error)` — sets `x-api-key` header, `Content-Type: application/json`, `User-Agent: mb-cli/<version>`. Returns error for 4xx/5xx with status + body
  - `Get(endpoint string, params url.Values) (*http.Response, error)` — builds GET request
  - `Post(endpoint string, body any) (*http.Response, error)` — builds POST request, marshals body to JSON
  - `DecodeJSON(resp *http.Response, v any) error` — decodes response body, defers Close
  - `HTTPDoer` interface (`Do(*http.Request) (*http.Response, error)`) for test injection

**Design decisions (following sentire patterns):**
- Auth via `x-api-key` header (Metabase API key auth)
- Errors: `fmt.Errorf("API request failed with status %d: %s", statusCode, body)`
- Always `defer resp.Body.Close()` after error check
- Verbose mode: log request URL and response status to stderr

**Tests:** `tests/client_test.go`
- `x-api-key` header correctly set
- `Content-Type` and `User-Agent` headers set
- GET request with query params
- POST request with JSON body
- 4xx/5xx error handling (returns error with status + body)
- Uses `httptest.NewServer` (pattern from sentire `tests/client_test.go`)

**Verification:** `make test`

---

### Step 3: Output Formatters ✅

Implement JSON and table output formatters before adding data commands.

**Files to create:**
- `internal/formatter/formatter.go`:
  - `Formatter` interface: `Format(data any, writer io.Writer) error`
  - `NewFormatter(format string) (Formatter, error)` factory
  - `Output(cmd *cobra.Command, data any) error` — reads `--format` flag, creates formatter, writes to stdout
  - `FormatQueryResults(columns []string, rows [][]any, writer io.Writer) error` — special method for tabular query results
- `internal/formatter/json.go` — pretty-printed JSON (`json.MarshalIndent`)
- `internal/formatter/table.go` — tabular output using `text/tabwriter` (stdlib, no extra dependency)

**Design:** Query results from `/api/dataset` have a specific shape (columns + rows) that needs special formatting. The table formatter renders column headers + aligned rows. The JSON formatter just passes through the raw data.

**Tests:** `tests/formatter_test.go` — JSON output, table output, query result formatting, edge cases (empty data, nil)

**Verification:** `make test`

---

### Step 4: Database Commands ✅

Implement all database exploration commands.

**Files to create:**
- `internal/client/types.go` — `Database`, `Table`, `Field`, `Schema` structs matching Metabase API responses
- `internal/client/databases.go`:
  - `ListDatabases(includeTables bool) ([]Database, error)` — `GET /api/database/?include=tables`
  - `GetDatabase(id int) (*Database, error)` — `GET /api/database/{id}`
  - `GetDatabaseMetadata(id int) (*DatabaseMetadata, error)` — `GET /api/database/{id}/metadata`
  - `GetDatabaseFields(id int) ([]Field, error)` — `GET /api/database/{id}/fields`
  - `ListDatabaseSchemas(id int) ([]string, error)` — `GET /api/database/{id}/schemas`
  - `GetDatabaseSchema(id int, schema string) ([]Table, error)` — `GET /api/database/{id}/schema/{schema}`
- `internal/cli/database.go` — Cobra subcommands:
  - `mb-cli database list` — list databases (id, name, engine)
  - `mb-cli database get <id>` — database details
  - `mb-cli database metadata <id>` — full metadata with tables + fields
  - `mb-cli database fields <id>` — all fields
  - `mb-cli database schemas <id>` — list schema names
  - `mb-cli database schema <id> <schema>` — tables in schema

**Tests:** `tests/databases_test.go` — mock responses for each endpoint, verify struct parsing

**Verification:** `make test && make build && ./bin/mb-cli database list`

---

### Step 5: Table Commands ✅

Implement table exploration commands.

**Files to create:**
- `internal/client/tables.go`:
  - `ListTables() ([]Table, error)` — `GET /api/table/`
  - `GetTable(id int) (*Table, error)` — `GET /api/table/{id}`
  - `GetTableMetadata(id int) (*TableMetadata, error)` — `GET /api/table/{id}/query_metadata`
  - `GetTableFKs(id int) ([]ForeignKey, error)` — `GET /api/table/{id}/fks`
  - `GetTableData(id int) (*QueryResult, error)` — `GET /api/table/{table-id}/data`
- `internal/cli/table.go` — Cobra subcommands:
  - `mb-cli table list` — list all tables
  - `mb-cli table get <id>` — table details
  - `mb-cli table metadata <id>` — table metadata with field details
  - `mb-cli table fks <id>` — foreign key relationships
  - `mb-cli table data <id>` — raw table data (uses query result formatter)

**Tests:** `tests/tables_test.go` — mock responses, verify parsing, table data formatting

**Verification:** `make test && ./bin/mb-cli table list && ./bin/mb-cli table data <id>`

---

### Step 6: Field Commands ✅

Implement field inspection commands.

**Files to create:**
- `internal/client/fields.go`:
  - `GetField(id int) (*Field, error)` — `GET /api/field/{id}`
  - `GetFieldSummary(id int) ([]FieldSummary, error)` — `GET /api/field/{id}/summary`
  - `GetFieldValues(id int) (*FieldValues, error)` — `GET /api/field/{id}/values`
- `internal/cli/field.go` — Cobra subcommands:
  - `mb-cli field get <id>` — field details (type, base_type, semantic_type)
  - `mb-cli field summary <id>` — summary stats (count, distinct, min, max)
  - `mb-cli field values <id>` — distinct values

**Tests:** `tests/fields_test.go` — mock responses, verify parsing

**Verification:** `make test && ./bin/mb-cli field get <id>`

---

### Step 7: SQL Query Command ✅

Implement native SQL query execution with database name resolution.

**Files to create/modify:**
- `internal/client/types.go` — add `DatasetQuery`, `NativeQuery`, `QueryResult`, `ResultColumn` types
- `internal/client/dataset.go`:
  - `RunNativeQuery(databaseID int, sql string) (*QueryResult, error)` — `POST /api/dataset/`
    ```json
    { "database": <id>, "type": "native", "native": { "query": "<sql>" } }
    ```
  - `ExportNativeQuery(databaseID int, sql string, format string) ([]byte, error)` — `POST /api/dataset/{format}`
- `internal/cli/query.go` — Cobra commands:
  - `mb-cli query sql --db <id-or-name> --sql "SELECT ..."` — run native SQL
  - `--db` flag: accepts numeric ID or name substring
  - `--export` flag: optional export format (csv, json, xlsx)
  - `--limit` flag: append LIMIT to SQL if provided

**Database name resolution:**
- If `--db` value is numeric -> use as database ID directly
- If non-numeric -> call `ListDatabases()`, case-insensitive substring match
- Error on zero matches: "no database matching '<name>' found"
- Error on multiple matches: "ambiguous database name '<name>', matches: [list]. Use database ID instead."

**Tests:** `tests/dataset_test.go`:
- SQL query request body construction
- Response parsing (columns + rows)
- Database name resolution: exact match, substring, no match, ambiguous match
- Export format request

**Verification:** `make test && ./bin/mb-cli query sql --db prod --sql "SELECT 1"`

---

### Step 8: Card (Saved Questions) Commands ✅

Implement listing and running saved questions.

**Files to create:**
- `internal/client/cards.go`:
  - `ListCards() ([]Card, error)` — `GET /api/card/`
  - `GetCard(id int) (*Card, error)` — `GET /api/card/{id}`
  - `RunCard(id int) (*QueryResult, error)` — `POST /api/card/{card-id}/query`
- `internal/client/types.go` — add `Card` struct (id, name, description, database_id, display, query_type, etc.)
- `internal/cli/card.go` — Cobra subcommands:
  - `mb-cli card list` — list saved questions (id, name, database, display type)
  - `mb-cli card get <id>` — card details
  - `mb-cli card run <id>` — execute saved question and show results

**Note:** `POST /api/card/{id}/query` is read-only — it runs the saved question without modifying anything.

**Tests:** `tests/cards_test.go` — mock list/get/run responses, verify request headers

**Verification:** `make test && ./bin/mb-cli card list && ./bin/mb-cli card run <id>`

---

### Step 9: Search Command ✅

Implement cross-entity search.

**Files to create:**
- `internal/client/search.go`:
  - `Search(query string, models []string) ([]SearchResult, error)` — `GET /api/search/?q=<query>&models=<models>`
- `internal/client/types.go` — add `SearchResult` struct
- `internal/cli/search.go` — Cobra command:
  - `mb-cli search <query> [--models table,card,database]` — search Metabase items
  - Default: search all models
  - `--models` flag: filter by type (table, card, database, dashboard, collection, metric)

**Tests:** `tests/search_test.go` — mock search responses with various model types

**Verification:** `make test && ./bin/mb-cli search "users"`

---

### Step 10: GitHub CI/CD + Dependabot ✅

Set up CI pipeline, release workflow, and dependabot.

**Files to create:**
- `.github/workflows/ci.yml` (modeled on logbasset `.github/workflows/ci.yml`):
  - Triggers: push to master, PRs to master
  - Test job: checkout, setup Go 1.25, cache modules, download deps, verify deps, `make test`, `make vet`, gofmt check
  - Build job: `make build`, upload artifact
  - Cross-platform build job: `make build-all`, upload artifacts
- `.github/workflows/release.yml` (modeled on logbasset `.github/workflows/release.yml`):
  - Triggers: `v*` tags
  - Steps: checkout (fetch-depth 0), setup Go, cache, `make test`, goreleaser v2
  - Permissions: contents write
  - Env: `GITHUB_TOKEN`, `GORELEASER_GITHUB_TOKEN` from `HOMEBREW_TAP_TOKEN` secret
- `.github/dependabot.yml` (identical to logbasset `.github/dependabot.yml`):
  - gomod: weekly, max 5 PRs
  - github-actions: weekly, max 5 PRs

**Verification:** push branch, confirm CI passes

---

### Step 11: GoReleaser + Homebrew ✅

Configure multi-platform release builds and homebrew tap.

**Files to create:**
- `.goreleaser.yaml` (modeled on logbasset `.goreleaser.yaml`):
  - Project name: `mb-cli`
  - Binary name: `mb-cli`
  - Entry: `./cmd/mb-cli/main.go`
  - CGO_ENABLED=0
  - Platforms: linux (amd64/arm64), darwin (amd64/arm64), windows (amd64, no arm64)
  - ldflags: `-s -w -X github.com/andreagrandi/mb-cli/internal/version.Version={{.Version}}`
  - Archives: tar.gz (linux/darwin), zip (windows); include README.md + LICENSE
  - Checksums: `checksums.txt`
  - Changelog: exclude docs:/test:/ci:/merge commits
  - Release: `Release v{{ .Version }}`
  - Homebrew: publish to `andreagrandi/homebrew-tap`, formula name `mb-cli`

**GitHub setup required:**
- Repository secret `HOMEBREW_TAP_TOKEN`: PAT with `repo` scope for `andreagrandi/homebrew-tap`
- `GITHUB_TOKEN` is automatic

**Verification:** `goreleaser check` locally, then `git tag v0.1.0 && git push origin v0.1.0`

---

### Step 12: Documentation + Final Polish ✅

- Update `README.md`:
  - Installation: homebrew (`brew install andreagrandi/tap/mb-cli`), binary download, `go install`
  - Configuration: `MB_HOST` and `MB_API_KEY` env vars
  - Usage examples for every command
  - Agent integration section: how Claude/Codex should use the tool
- Finalize `AGENTS.md` with all project conventions
- Add a **Release Process** section to `AGENTS.md` documenting the exact release sequence, modeled on [mcp-wire AGENTS.md](https://github.com/andreagrandi/mcp-wire/blob/master/AGENTS.md#release-process): ensure tests pass, update version in `internal/version/version.go`, update CHANGELOG.md, commit, push, tag, push tag. Note that the tag push triggers `.github/workflows/release.yml` and no manual GitHub release should be created before the workflow runs.
- Create `CHANGELOG.md` with `[Unreleased]` section

**Verification:** review README, ensure all commands documented with examples

---

### Step 13: Agent-Friendly Enhancements ✅

Make the CLI optimised for AI agent consumption, following the best practices from [Rewrite your CLI for AI Agents](https://justin.poehnelt.com/posts/rewrite-your-cli-for-ai-agents/) and mirroring what was done in [logbasset PR #33](https://github.com/andreagrandi/logbasset/pull/33).

#### 13a. `mb-cli context` command — embedded agent reference document

**Files to create:**
- `internal/cli/context.go` — Cobra command that prints an embedded markdown document (pattern from logbasset `internal/cli/context.go`)
- `internal/cli/context_embed.md` — ~100-line agent context document embedded via `//go:embed`

**Content of `context_embed.md`:**
- Tool description: mb-cli is a **read-only** CLI for querying Metabase databases
- Authentication: `MB_HOST` and `MB_API_KEY` env vars
- Command table: every command with required args and flags
- Global flags table: `--format`, `--verbose`
- **Flags that do NOT exist** section: common hallucinations agents might try (e.g. `--host`, `--token`, `--database` as global flag, `--output` instead of `--format`, `--query` instead of `--sql`)
- Output formats: json (default when piped), table (default in TTY)
- Database name resolution: explain that `--db` accepts ID or name substring, and what happens on ambiguous/no match
- Structured error output: format and exit codes
- Worked examples for common agent workflows (list dbs, find table, query data)

**Tests:** `tests/context_test.go` — verify command runs, output contains key sections

#### 13b. `mb-cli schema [command]` — JSON schema introspection

**Files to create:**
- `internal/cli/schema.go` — Cobra command (pattern from logbasset `internal/cli/schema.go`)
  - `mb-cli schema` — lists all commands with descriptions as JSON array
  - `mb-cli schema <command>` — prints JSON schema for that command's args, flags (name, type, required, default, enum, description), and output keys
  - `--pretty` flag for indented output

**Schema types:**
```go
type commandSummary struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type paramSchema struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`
    Required    bool        `json:"required"`
    Default     any         `json:"default,omitempty"`
    Enum        []string    `json:"enum,omitempty"`
    Description string      `json:"description"`
}

type commandSchema struct {
    Command    string        `json:"command"`
    Args       []paramSchema `json:"args,omitempty"`
    Flags      []paramSchema `json:"flags"`
    OutputKeys []string      `json:"output_keys,omitempty"`
}
```

**Tests:** `tests/schema_test.go` — verify JSON output is valid, verify all commands are listed, verify schema for a specific command

#### 13c. TTY auto-detection — smart output format defaults

**Files to create:**
- `internal/cli/tty.go` — `IsTTY` function using `os.Stdout.Stat()` (pattern from logbasset `internal/cli/tty.go`)

**Behaviour change in `root.go`:**
- When stdout is a TTY: default `--format` to `table` (human-readable)
- When stdout is piped (non-TTY): default `--format` to `json` (machine-readable)
- Explicit `--format` flag always overrides the auto-detection

**Tests:** `tests/tty_test.go` — verify auto-detection logic (mock `IsTTY` var)

#### 13d. `--error-format json` — structured error output on stderr

**Behaviour change in `root.go`:**
- Add `--error-format` persistent flag: `text` (default) or `json`
- When `json`: errors written to stderr as:
  ```json
  {"error":{"type":"CONFIG_ERROR","message":"MB_API_KEY is required","suggestion":"Set MB_API_KEY environment variable","exit_code":1}}
  ```
- When `text`: errors printed as normal (current behaviour)

**Files to modify:**
- `internal/cli/root.go` — add `--error-format` flag, wire it into error handling
- `cmd/mb-cli/main.go` — structured error handler that checks format

**Tests:** `tests/error_format_test.go` — verify JSON error output on stderr for config error, API error

#### 13e. Input validation — reject control characters and path traversal

**Files to create:**
- `internal/validation/validation.go` — (adapted from logbasset `internal/validation/validation.go`)
  - `ValidateNoControlChars(input, fieldName string) error` — reject ASCII 0x00-0x1F (except tab/newline/CR) and 0x7F
  - `ValidateSQL(sql string) error` — max length check, control char rejection
  - `ValidateSearchQuery(query string) error` — control char rejection

**Files to modify:**
- `internal/cli/query.go` — validate `--sql` input before sending to API
- `internal/cli/search.go` — validate search query input

**Tests:** `tests/validation_test.go` — control chars rejected, path traversal rejected, valid inputs pass

#### 13f. `--fields` flag for JSON field masking on query results

**Behaviour:** Add `--fields` flag to `mb-cli query sql` and `mb-cli card run` commands. When set (e.g. `--fields id,name,email`), the JSON output includes only those columns from the result set. This reduces output size for agent context windows.

**Files to modify:**
- `internal/cli/query.go` — add `--fields` flag
- `internal/cli/card.go` — add `--fields` flag
- `internal/formatter/json.go` — filter columns/rows based on fields list

**Tests:** `tests/fields_filter_test.go` — verify field filtering on query results

---

**Summary of new files for Step 13:**
```
internal/cli/context.go           # context command
internal/cli/context_embed.md     # embedded agent reference doc
internal/cli/schema.go            # schema introspection command
internal/cli/tty.go               # TTY detection
internal/validation/validation.go # input hardening
tests/context_test.go
tests/schema_test.go
tests/tty_test.go
tests/error_format_test.go
tests/validation_test.go
tests/fields_filter_test.go
```

**New commands added:**
```
mb-cli context                         # Print agent context document
mb-cli schema                          # List all commands as JSON
mb-cli schema <command>                # JSON schema for command inputs
```

**Verification:**
- `mb-cli context | head` — prints agent reference
- `mb-cli schema | jq .` — valid JSON, lists all commands
- `mb-cli schema "query sql" | jq .` — shows args/flags/defaults
- `echo '{}' | mb-cli database list` — JSON output (non-TTY detected)
- `mb-cli database list` — table output in terminal (TTY detected)
- `MB_API_KEY="" mb-cli database list --error-format json 2>&1` — structured JSON error on stderr
- `mb-cli query sql --db prod --sql $'\x01SELECT 1'` — rejected with validation error
- `mb-cli query sql --db prod --sql "SELECT id, name FROM users" --format json --fields id,name` — only id,name in output

---

### Step 14: Structured Query (Filter) Command ⬚

Add a `query filter` command that builds Metabase structured queries (MBQL) instead of requiring raw SQL. This lets users and agents look up rows by field values with simple flags.

**Usage:**
```bash
mb-cli query filter --db prod --table products --where "id=prod_1234"
mb-cli query filter --db 1 --table users --where "name=alice" --where "active=true"
mb-cli query filter --db 1 --table orders --where "status=pending" --limit 10
mb-cli query filter --db 1 --table products --where "id=prod_1234" --export csv
```

**Files to create/modify:**
- `internal/client/types.go` — add `StructuredQuery` struct, update `DatasetQuery` to use pointer fields (`*NativeQuery`, `*StructuredQuery`) so JSON omits the unused one
- `internal/client/dataset.go` — add `RunStructuredQuery(databaseID, tableID int, filters [][]any, limit int) (*QueryResult, error)`, update existing methods for pointer `Native`
- `internal/cli/query.go` — add `queryFilterCmd` with flags:
  - `--db` (string, required) — database ID or name substring (reuses `resolveDatabaseID`)
  - `--table` (string, required) — table ID or name substring (new `resolveTableID`)
  - `--where` (string slice, required) — filter in `field=value` format (repeatable)
  - `--limit` (int, optional) — max rows
  - `--export` (string, optional) — export format (csv, json, xlsx)
- `tests/dataset_test.go` — add tests for structured query request body, table/field resolution

**Table name resolution** (new `resolveTableID` in `query.go`):
- Numeric → use as table ID
- Non-numeric → call `GetDatabaseMetadata(dbID)`, case-insensitive substring match on table names within that database
- Error on zero/multiple matches (same pattern as `matchDatabaseByName`)

**Field name resolution** (new `resolveFieldID` in `query.go`):
- Parse `--where "field_name=value"` into field name + value
- Call `GetTableMetadata(tableID)` to get fields
- Case-insensitive exact match on field `name` or `display_name`
- Return field ID for MBQL filter construction

**MBQL query format** — POST `/api/dataset/`:
```json
{
  "database": 1,
  "type": "query",
  "query": {
    "source-table": 42,
    "filter": ["and", ["=", ["field", 100, null], "prod_1234"]],
    "limit": 10
  }
}
```

- Single filter: `["=", ["field", <field_id>, null], <value>]`
- Multiple filters: `["and", <filter1>, <filter2>, ...]`
- Only `=` operator for v1 (can extend later with `!=`, `>`, `<`, `contains`, etc.)

**Tests:**
- `TestRunStructuredQuery` — verify request body structure
- `TestRunStructuredQueryMultipleFilters` — verify AND combination
- `TestRunStructuredQueryWithLimit` — verify limit in query body
- `TestResolveTableID` — numeric passthrough + name resolution
- `TestParseWhereClause` — parsing `field=value` strings

**Verification:**
```bash
make test
make build
./bin/mb-cli query filter --help
./bin/mb-cli query filter --db prod --table products --where "id=prod_1234"
./bin/mb-cli query filter --db prod --table products --where "id=prod_1234" --format table
./bin/mb-cli query filter --db prod --table products --where "id=prod_1234" --export csv
```

---

## Development Note

`MB_HOST` and `MB_API_KEY` are available in the project `.envrc` (loaded by direnv). After each step, **always smoke-test against the real Metabase API** as soon as there is something runnable — don't wait until all steps are done. For example, after Step 4 (database commands), run `make build && ./bin/mb-cli database list` to confirm the API integration works end-to-end, not just the unit tests.

---

## Setup Instructions

### Environment Variables
```bash
export MB_HOST="https://your-metabase-instance.com"
export MB_API_KEY="your-api-key-here"
```

### GitHub Repository Secrets
- `HOMEBREW_TAP_TOKEN`: Personal access token with `repo` scope for `andreagrandi/homebrew-tap`

### Release Process
```bash
git tag v0.1.0
git push origin v0.1.0
# GoReleaser builds binaries, creates GitHub release, updates homebrew tap
```

---

## Command Summary

```
mb-cli database list                              # List all databases
mb-cli database get <id>                          # Get database details
mb-cli database metadata <id>                     # Full metadata (tables + fields)
mb-cli database fields <id>                       # List all fields
mb-cli database schemas <id>                      # List schema names
mb-cli database schema <id> <schema>              # Tables in a schema

mb-cli table list                                 # List all tables
mb-cli table get <id>                             # Get table details
mb-cli table metadata <id>                        # Table metadata with fields
mb-cli table fks <id>                             # Foreign key relationships
mb-cli table data <id>                            # Raw table data

mb-cli field get <id>                             # Field details
mb-cli field summary <id>                         # Summary statistics
mb-cli field values <id>                          # Distinct values

mb-cli query sql --db <id-or-name> --sql "..."    # Run SQL query
mb-cli query sql --db prod --sql "..." --export csv  # Export results
mb-cli query filter --db <id-or-name> --table <name> --where "field=value"  # Structured query
mb-cli query filter --db prod --table products --where "id=prod_1234"      # Filter by field

mb-cli card list                                  # List saved questions
mb-cli card get <id>                              # Card details
mb-cli card run <id>                              # Execute saved question

mb-cli search <query>                             # Search Metabase items
mb-cli search <query> --models table,card         # Search specific types

mb-cli context                                    # Print agent reference document
mb-cli schema                                     # List all commands as JSON
mb-cli schema <command>                           # JSON schema for command inputs

mb-cli version                                    # Show version

# Global flags (all commands):
--format json|table    (default: json when piped, table in TTY)
--verbose              (show request details on stderr)
--error-format text|json  (default: text; json for structured errors on stderr)
```

---

## Key Patterns Reused

| Pattern | Source | File Reference |
|---------|--------|----------------|
| Cobra CLI + subcommands | both | `sentire/internal/cli/root.go`, `logbasset/internal/cli/root.go` |
| Simple env var config | sentire | `sentire/internal/config/config.go` |
| `fmt.Errorf` wrapping | sentire | `sentire/internal/client/client.go` |
| `httptest.NewServer` tests | both | `sentire/tests/client_test.go`, `logbasset/internal/client/client_test.go` |
| Formatter interface | sentire | `sentire/internal/cli/formatter/formatter.go` |
| `text/tabwriter` tables | stdlib | (no extra dependency) |
| GoReleaser + Homebrew | both | `logbasset/.goreleaser.yaml` |
| Dependabot config | both | `logbasset/.github/dependabot.yml` |
| CI/release workflows | both | `logbasset/.github/workflows/ci.yml`, `logbasset/.github/workflows/release.yml` |
| Version injection via ldflags | both | `logbasset/.goreleaser.yaml` line 26 |
| DB name fuzzy matching | new | For agent-friendly UX |
| `context` embedded doc (go:embed) | logbasset PR #33 | `internal/cli/context.go` + `context_embed.md` |
| `schema` JSON introspection | logbasset PR #33 | `internal/cli/schema.go` |
| TTY auto-detection | logbasset PR #33 | `internal/cli/tty.go` |
| `--error-format json` on stderr | logbasset PR #33 | `internal/cli/root.go`, `cmd/mb-cli/main.go` |
| Input hardening (control chars) | logbasset PR #33 | `internal/validation/validation.go` |
| `--fields` for context window savings | logbasset PR #33 | `internal/cli/query.go`, `internal/formatter/json.go` |
