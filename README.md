# mb-cli

<img src="mb-cli-logo.png" width="200" alt="mb-cli">

A read-only CLI for the Metabase API. Designed for terminal users and AI coding agents (Claude, Codex) to explore databases, inspect schemas, and run ad-hoc queries.

## Installation

### Homebrew

```bash
brew tap andreagrandi/tap
brew install mb-cli
```

### Binary download

Download the latest release from the [releases page](https://github.com/andreagrandi/mb-cli/releases).

### Go install

```bash
go install github.com/andreagrandi/mb-cli/cmd/mb-cli@latest
```

## Configuration

Set `MB_HOST` plus **one** of the two authentication variables.

### Option 1: API key (recommended)

```bash
export MB_HOST=https://your-metabase-instance.com
export MB_API_KEY=your-api-key
```

`MB_API_KEY` is a long-lived [Metabase API key](https://www.metabase.com/docs/latest/people-and-groups/api-keys). Generating one requires Metabase admin access.

### Option 2: Session token

```bash
export MB_HOST=https://your-metabase-instance.com
export MB_SESSION_TOKEN=your-session-token
```

Use this when you do not have admin access to mint an API key. Grab the value of the `metabase.SESSION` cookie from your browser (DevTools → Application → Cookies). Session tokens expire on logout or when the Metabase server's session timeout fires, so they are best for short-lived use.

`MB_API_KEY` and `MB_SESSION_TOKEN` are mutually exclusive — set one or the other, not both.

### Optional

```bash
export MB_REDACT_PII=false  # Disable PII redaction (enabled by default)
```

## Usage

### Global flags

| Flag | Description | Default |
|------|-------------|---------|
| `--format`, `-f` | Output format: `json`, `table` | `json` |
| `--verbose`, `-v` | Show request details on stderr | `false` |
| `--redact-pii` | Redact PII values in query results | `true` |
| `--timeout` | Timeout for a command's API requests (`0` disables) | `30s` |

### Database commands

```bash
# List all databases
mb-cli database list

# Get database details
mb-cli database get 1

# Full metadata (tables + fields)
mb-cli database metadata 1

# List all fields in a database
mb-cli database fields 1

# List schema names
mb-cli database schemas 1

# List tables in a specific schema
mb-cli database schema 1 public
```

The `database` command has an alias `db`:

```bash
mb-cli db list
```

### Table commands

```bash
# List all tables
mb-cli table list

# Get table details
mb-cli table get 42

# Table metadata with field details
mb-cli table metadata 42

# Foreign key relationships
mb-cli table fks 42

# Raw table data
mb-cli table data 42
```

### Field commands

```bash
# Get field details
mb-cli field get 100

# Summary statistics
mb-cli field summary 100

# Distinct values
mb-cli field values 100
```

### SQL queries

```bash
# Run a SQL query by database ID
mb-cli query sql --db 1 --sql "SELECT * FROM users LIMIT 10"

# Run a SQL query by database name (case-insensitive substring match)
mb-cli query sql --db prod --sql "SELECT count(*) FROM orders"

# Append a LIMIT clause
mb-cli query sql --db 1 --sql "SELECT * FROM users" --limit 5

# Export results as CSV
mb-cli query sql --db 1 --sql "SELECT * FROM users" --export csv
```

The `--db` flag accepts a numeric database ID or a name substring. If the name matches multiple databases, you'll be asked to use the ID instead.

### Saved questions (cards)

```bash
# List all saved questions
mb-cli card list

# Get card details
mb-cli card get 10

# Get the full card definition, including dataset_query and template tags
mb-cli card get 10 --full

# Execute a saved question
mb-cli card run 10

# Execute a saved question with parameters
mb-cli card run 10 --param timeframe_days=14
```

### Dashboards

```bash
# List all dashboards
mb-cli dashboard list

# Get dashboard metadata, tabs, filters, and grouped cards
mb-cli dashboard get 298

# List the saved questions used by a dashboard
mb-cli dashboard cards 298

# Discover valid values for a dashboard filter
mb-cli dashboard params values 298 merchant_name

# Search dashboard filter values
mb-cli dashboard params search 298 merchant_name acme

# Execute a dashboard card with dashboard parameters applied
mb-cli dashboard run-card 298 1201 398 --param merchant_name="Acme Corp"

# Summarize tabs, parameter mappings, and source-card dependencies
mb-cli dashboard analyze 298
```

### Search

```bash
# Search across all Metabase items
mb-cli search "users"

# Filter by type
mb-cli search "revenue" --models table,card
```

### Version

```bash
mb-cli version
```

## Recipes

[docs/RECIPES.md](docs/RECIPES.md) collects end-to-end recipes for common
Metabase exploration tasks — schema discovery, saved-question inspection,
dashboard analysis, parameterized card and dashboard runs, and PII-safe
querying. Each recipe is a worked sequence of commands with sample output.

## Agent integration

mb-cli is designed to be used by AI coding agents. When piped or used in scripts, output defaults to JSON for easy parsing.

Typical agent workflow:

```bash
# 1. Discover databases
mb-cli database list

# 2. Find relevant tables
mb-cli search "users" --models table

# 3. Inspect table schema
mb-cli table metadata 42

# 4. Query data
mb-cli query sql --db 1 --sql "SELECT id, email FROM users WHERE created_at > '2024-01-01' LIMIT 10"

# 5. Inspect a dashboard that depends on saved questions
mb-cli dashboard analyze 298
```

## Dashboard analysis workflow

Example flow for dashboard `298`:

```bash
# 1. Inspect dashboard structure
mb-cli dashboard get 298

# 2. List the saved questions behind the dashboard
mb-cli dashboard cards 298

# 3. Inspect a card's full SQL or MBQL definition
mb-cli card get 398 --full

# 4. Discover valid parameter values
mb-cli dashboard params values 298 merchant_name

# 5. Summarize dependencies and assumption-backed cards
mb-cli dashboard analyze 298
```

## API coverage

The dashboard and parameter workflows use these Metabase endpoints:

| Command | Endpoint |
|---------|----------|
| `dashboard list` | `GET /api/dashboard/` |
| `dashboard get` | `GET /api/dashboard/:id` |
| `dashboard cards` | `GET /api/dashboard/:id` |
| `dashboard params values` | `GET /api/dashboard/:id/params/:param-key/values` |
| `dashboard params search` | `GET /api/dashboard/:id/params/:param-key/search/:query` |
| `dashboard run-card` | `POST /api/dashboard/:dashboard-id/dashcard/:dashcard-id/card/:card-id/query` |
| `card get --full` | `GET /api/card/:id` |
| `card run --param` | `POST /api/card/:card-id/query` |
| `dashboard analyze` | `GET /api/dashboard/:id`, `GET /api/card/:id` |

## PII Redaction

When AI agents use mb-cli directly (via shell commands), query results containing PII (emails, names, phone numbers) flow through stdout into the model's context. mb-cli prevents that by replacing PII column values with `[REDACTED]` before data leaves the client layer. Agents can still cross-reference records using IDs; for actual PII values, the user can check directly in Metabase.

Redaction is **ON by default**. Disabling it is a visible, auditable action: when the flag or env var is set to `false`, mb-cli prints `Warning: PII redaction is disabled` to stderr on every invocation.

### What gets redacted

Columns whose Metabase field semantic type is one of:

`type/Email`, `type/Name`, `type/Phone`, `type/Address`, `type/City`, `type/State`, `type/ZipCode`, `type/Country`, `type/Latitude`, `type/Longitude`, `type/Birthdate`, `type/AvatarURL`, `type/URL`, `type/ImageURL`, `type/Company`.

For native SQL results, Metabase does not always populate the semantic type on result columns. mb-cli fills it in by fetching field metadata from the source database and matching by column name. When two fields share a name and disagree, mb-cli picks the PII type (err on the side of caution).

### Where redaction applies

| Command | Redacted? |
|---------|-----------|
| `query sql` | Yes (with column-name enrichment) |
| `query filter` (MBQL) | Yes |
| `card run` (with or without `--param`) | Yes (with column-name enrichment) |
| `dashboard run-card` | Yes (with column-name enrichment) |
| `table data` | Yes |
| `field values` | Yes, when the field's semantic type is PII |

Schema and metadata commands (`database metadata`, `table metadata`, `field get`, etc.) return field definitions, not data, so there is nothing to redact.

### Limits and known gaps

- Redaction relies on Metabase semantic types. A column holding PII without a semantic type (or with a custom type not in the list above) will not be redacted. Set semantic types in Metabase to close that gap.
- Derived columns from joins, aggregations, or `CASE` expressions may lose the upstream semantic type. Native SQL enrichment matches by column name only — alias derived columns to match a known PII field name if you want them caught.
- Free-form `type/Description` fields are not treated as PII; if you store names or emails in description-style columns, set their semantic type explicitly.
- Redaction operates on result rows, not column headers. Column names like `email` remain visible — values are replaced.

### Export behavior

When redaction is enabled, the `--export` flag on `query sql` and `query filter` is **blocked** with this error:

```
export is not supported when PII redaction is enabled (use JSON or table format instead)
```

Raw export bytes (CSV/XLSX/JSON straight from Metabase) cannot be post-processed reliably for redaction, so the safe default is to refuse the operation. To export, explicitly opt out of redaction for that invocation:

```bash
MB_REDACT_PII=false mb-cli query sql --db 1 --sql "..." --export csv
# or
mb-cli query sql --redact-pii=false --db 1 --sql "..." --export csv
```

### Opt-out precedence

The `--redact-pii` flag wins if set; otherwise `MB_REDACT_PII` is consulted; otherwise the default is `true`. Any value other than the literal string `false` is treated as enabled.

## License

MIT
