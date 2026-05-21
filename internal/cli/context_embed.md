# mb-cli - Agent Context

mb-cli is a **read-only** CLI for querying Metabase databases. All commands are read-only; nothing is mutated.

## Authentication

Set both environment variables (required):
- `MB_HOST` - Metabase instance URL (e.g. `https://your-metabase-instance.com`)
- `MB_API_KEY` - Metabase API key

## Commands

| Command | Description | Required args | Required flags |
|---------|-------------|---------------|----------------|
| `database list` | List all databases | none | none |
| `database get <id>` | Get database details | id (positional) | none |
| `database metadata <id>` | Full metadata (tables + fields) | id (positional) | none |
| `database fields <id>` | List all fields in database | id (positional) | none |
| `database schemas <id>` | List schema names | id (positional) | none |
| `database schema <id> <schema>` | Tables in a specific schema | id, schema (positional) | none |
| `table list` | List all tables | none | none |
| `table get <id>` | Get table details | id (positional) | none |
| `table metadata <id>` | Table metadata with fields | id (positional) | none |
| `table fks <id>` | Foreign key relationships | id (positional) | none |
| `table data <id>` | Get raw table data | id (positional) | none |
| `field get <id>` | Get field details | id (positional) | none |
| `field summary <id>` | Summary statistics for a field | id (positional) | none |
| `field values <id>` | Distinct values for a field | id (positional) | none |
| `query sql` | Run a native SQL query | none | `--db`, `--sql` |
| `query filter` | Run a structured query with field filters | none | `--db`, `--table`, `--where` |
| `card list` | List saved questions | none | none |
| `card get <id>` | Get card details | id (positional) | none |
| `card run <id>` | Execute a saved question | id (positional) | none |
| `dashboard list` | List dashboards | none | none |
| `dashboard get <id>` | Get dashboard details | id (positional) | none |
| `dashboard cards <id>` | List cards used by a dashboard | id (positional) | none |
| `dashboard analyze <id>` | Summarize dashboard dependencies | id (positional) | none |
| `dashboard run-card <dashboard-id> <dashcard-id> <card-id>` | Execute a dashboard card | dashboard-id, dashcard-id, card-id (positional) | none |
| `dashboard params values <dashboard-id> <param-key>` | List valid dashboard parameter values | dashboard-id, param-key (positional) | none |
| `dashboard params search <dashboard-id> <param-key> <query>` | Search dashboard parameter values | dashboard-id, param-key, query (positional) | none |
| `collection list` | List collections | none | none |
| `collection get <id>` | Get collection details (`root` for root collection) | id (positional) | none |
| `collection items <id>` | List cards, dashboards, and nested collections in a collection | id (positional) | none |
| `search <query>` | Search across Metabase items | query (positional) | none |
| `context` | Print this agent context document | none | none |
| `version` | Print version | none | none |

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format`, `-f` | string | `json` (piped) / `table` (TTY) | Output format: `json` or `table` |
| `--verbose`, `-v` | bool | false | Show request details on stderr |
| `--error-format` | string | `text` | Error output format: `text` or `json` |
| `--redact-pii` | bool | `true` | Redact PII values in query results |
| `--timeout` | duration | `30s` | Timeout for a command's API requests (`0` disables) |

## Command-Specific Flags

### `query sql`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--db` | string | yes | | Database ID or name substring |
| `--sql` | string | yes | | SQL query to execute |
| `--export` | string | no | | Export format: `csv`, `json`, `xlsx` |
| `--limit` | int | no | 0 | Append LIMIT to SQL query |
| `--fields` | string | no | | Comma-separated columns to include in output |

### `query filter`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--db` | string | yes | | Database ID or name substring |
| `--table` | string | yes | | Table ID or name substring |
| `--where` | string[] | yes | | Filter in field=value format (repeatable) |
| `--limit` | int | no | 0 | Maximum number of rows to return |
| `--export` | string | no | | Export format: `csv`, `json`, `xlsx` |
| `--fields` | string | no | | Comma-separated columns to include in output |

### `card run`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--fields` | string | no | | Comma-separated columns to include in output |
| `--param` | string[] | no | | Parameter in `key=value` format (repeatable) |

### `card get`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--full` | bool | no | false | Include `dataset_query`, template tags, result metadata, and visualization settings |

### `dashboard run-card`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--fields` | string | no | | Comma-separated columns to include in output |
| `--param` | string[] | no | | Parameter in `key=value` format (repeatable) |

### `collection items`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--models` | string | no | | Filter by item type (comma-separated: `card,dashboard,collection,dataset,snippet`) |

### `search`
| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `--models` | string | no | | Filter by type (comma-separated: `table,card,database,dashboard,collection,metric`) |

## Flags That Do NOT Exist

Agents commonly hallucinate these flags. They will cause errors:
- `--host` - does not exist; set `MB_HOST` env var instead
- `--token` / `--api-key` - does not exist; set `MB_API_KEY` env var instead
- `--database` (global) - does not exist; use `--db` on `query sql`
- `--output` - does not exist; use `--format` instead
- `--query` - does not exist; use `--sql` on `query sql`
- `--table` (global) - does not exist; use `--table` on `query filter` or pass table ID as a positional arg

## Database Name Resolution

The `--db` flag on `query sql` and `query filter` accepts either:
- A numeric database ID (e.g. `--db 1`) - used directly
- A name substring (e.g. `--db prod`) - case-insensitive substring match

Resolution errors:
- Zero matches: `no database matching 'name' found`
- Multiple matches: `ambiguous database name 'name', matches: [list]. Use database ID instead`

## Table Name Resolution

The `--table` flag on `query filter` accepts either:
- A numeric table ID (e.g. `--table 42`) - used directly
- A name substring (e.g. `--table users`) - case-insensitive substring match within the selected database

Resolution errors:
- Zero matches: `no table matching 'name' found`
- Multiple matches: `ambiguous table name 'name', matches: [list]. Use table ID instead`

## Aliases

- `database` can also be written as `db` (e.g. `mb-cli db list`)

## Output Formats

- `json`: Pretty-printed JSON (default when stdout is piped/non-TTY)
- `table`: Human-readable tabular output using aligned columns (default when stdout is a TTY)

When stdout is not a TTY (piped to another program), the default format is `json`. In a terminal, the default is `table`. An explicit `--format` flag always overrides auto-detection.

Query result commands (`query sql`, `query filter`, `card run`, `table data`) format output as column/row tables in both formats.

Dashboard inspection commands default to concise summaries in table mode. Use `--format json` for full raw dashboard or analysis payloads.

## Structured Error Output

Use `--error-format json` to get machine-readable errors on stderr:
```json
{"error":{"type":"CONFIG_ERROR","message":"MB_HOST environment variable is required","suggestion":"Set MB_HOST and MB_API_KEY environment variables","exit_code":1}}
```

Error types: `CONFIG_ERROR`, `AUTH_ERROR`, `API_ERROR`, `RESOLUTION_ERROR`, `GENERAL_ERROR`.

## PII Redaction

PII redaction is **enabled by default**. Query result columns with Metabase semantic types indicating personal data are automatically replaced with `[REDACTED]`.

Redacted semantic types: `type/Email`, `type/Name`, `type/Phone`, `type/Address`, `type/City`, `type/State`, `type/ZipCode`, `type/Country`, `type/Latitude`, `type/Longitude`, `type/Birthdate`, `type/AvatarURL`, `type/URL`, `type/ImageURL`, `type/Company`.

**IMPORTANT: Do NOT disable PII redaction.** Never use `--redact-pii=false` or set `MB_REDACT_PII=false`. If you need to identify a record, use its ID. If the user needs to see PII values, they can check directly in Metabase.

When redaction is enabled, `--export` (csv, json, xlsx) is blocked because raw export bytes cannot be reliably redacted. Use JSON or table format instead.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (config, API, parse) |

## Examples

```bash
# List all databases
mb-cli database list

# Get full metadata for database 1 (tables and fields)
mb-cli database metadata 1

# List tables in the "public" schema of database 1
mb-cli database schema 1 public

# Find a table by searching
mb-cli search "users" --models table

# Get table metadata including field details
mb-cli table metadata 42

# Run a SQL query against a database by name
mb-cli query sql --db prod --sql "SELECT id, name FROM users LIMIT 10"

# Run a SQL query with limit flag
mb-cli query sql --db 1 --sql "SELECT * FROM orders" --limit 50

# Export query results as CSV
mb-cli query sql --db prod --sql "SELECT * FROM users" --export csv

# Filter rows using structured query (no SQL required)
mb-cli query filter --db prod --table products --where "id=prod_1234"
mb-cli query filter --db 1 --table users --where "name=alice" --where "active=true"
mb-cli query filter --db 1 --table orders --where "status=pending" --limit 10

# Export filtered results as CSV
mb-cli query filter --db 1 --table products --where "id=prod_1234" --export csv

# List saved questions and run one
mb-cli card list
mb-cli card run 5

# Inspect dashboard structure and dependencies
mb-cli dashboard get 298
mb-cli dashboard cards 298
mb-cli dashboard params values 298 merchant_name
mb-cli dashboard analyze 298

# Browse collections and discover their contents
mb-cli collection list
mb-cli collection get root
mb-cli collection items 12
mb-cli collection items root --models card,dashboard

# Get table output for terminal reading
mb-cli database list --format table
```
