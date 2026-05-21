# mb-cli Recipes

End-to-end recipes for common Metabase exploration tasks. Each recipe is a
sequence of commands where every step feeds an ID or name into the next, so you
discover what you need instead of guessing.

These recipes complement the command reference in the [README](../README.md)
and the agent reference printed by `mb-cli context`.

## Conventions

- Examples use a generic e-commerce-style Metabase (a `Production` database
  with `users` and `orders` tables). Replace IDs and names with your own.
- Output is trimmed for brevity; real responses have more rows and fields.
- All examples were verified against mb-cli `v0.2.0`. mb-cli is **read-only** —
  no recipe mutates Metabase.
- Output format defaults to `table` in a terminal and `json` when piped. The
  examples pass `--format` explicitly so the output is reproducible.

## Before you start

Set the connection environment variables once per shell:

```bash
export MB_HOST=https://your-metabase-instance.com
export MB_API_KEY=your-api-key
```

See the [README](../README.md#configuration) for the session-token
alternative.

**For agents:** prefer the metadata-first paths below over guessing table,
column, or parameter names. When a command fails, its error message names the
exact discovery command to run next — follow it.

---

## Recipe 1 — Discover a database schema

**Goal:** go from "which databases exist?" to "what columns does this table
have?" without guessing names.

### 1. List databases to find the ID

```bash
mb-cli database list --format table
```

```
id  name        engine    details  tables
1   Production  postgres           []
2   Analytics   postgres           []
```

### 2. List the schemas in that database

```bash
mb-cli database schemas 1 --format table
```

```
public
```

### 3. List the tables in a schema

```bash
mb-cli database schema 1 public --format table
```

```
id  name    display_name  schema  db_id  entity_type
42  users   Users         public  1      entity/UserTable
51  orders  Orders        public  1      entity/GenericTable
```

If you do not know which schema a table lives in, search instead:

```bash
mb-cli search "orders" --models table --format table
```

```
id  name    description  model  database_id  table_id  collection_id  archived
51  Orders               table  1            51        <nil>          false
```

### 4. Inspect a table's columns

```bash
mb-cli table metadata 42 --format json
```

```json
{
  "id": 42,
  "name": "users",
  "display_name": "Users",
  "schema": "public",
  "db_id": 1,
  "fields": [
    {
      "id": 100,
      "name": "email",
      "display_name": "Email",
      "base_type": "type/Text",
      "database_type": "varchar",
      "semantic_type": "type/Email",
      "table_id": 42
    },
    {
      "id": 112,
      "name": "status",
      "display_name": "Status",
      "base_type": "type/Text",
      "database_type": "varchar",
      "semantic_type": "type/Category",
      "table_id": 42
    }
  ]
}
```

The `semantic_type` drives PII redaction — `type/Email` here means the `email`
column is redacted in query results (see [Recipe 6](#recipe-6--pii-safe-ad-hoc-querying)).

### 5. Follow foreign keys to related tables

```bash
mb-cli table fks 51 --format json
```

```json
[
  {
    "relationship": "Mt1",
    "origin": {
      "id": 305,
      "name": "user_id",
      "table": { "id": 51, "name": "orders" }
    },
    "destination": {
      "id": 99,
      "name": "id",
      "table": { "id": 42, "name": "users" }
    }
  }
]
```

`orders.user_id` references `users.id` — a many-to-one (`Mt1`) relationship.

### 6. Understand a column before filtering on it

```bash
mb-cli field summary 112 --format table
```

```
type       value
count      48210
distincts  3
```

```bash
mb-cli field values 112 --format table
```

```
field_id  112
values    [[active] [closed] [pending]]
```

Now you know `status` has three values and is safe to filter on.

---

## Recipe 2 — Inspect a saved question (card)

**Goal:** understand what a saved question does and what data it returns
before running it.

### 1. Find the card

```bash
mb-cli search "active users" --models card --format table
```

```
id  name                 description                    model  database_id  table_id  collection_id  archived
10  Weekly Active Users  Distinct users seen this week  card   1            0         3              false
```

`card list` works too if you want to browse everything.

### 2. Inspect the card definition

```bash
mb-cli card get 10 --format json
```

```json
{
  "id": 10,
  "name": "Weekly Active Users",
  "description": "Distinct users seen this week",
  "database_id": 1,
  "display": "scalar",
  "query_type": "native",
  "collection_id": 3,
  "archived": false
}
```

### 3. See the underlying SQL or MBQL

Add `--full` to get `dataset_query`, template tags, and result metadata:

```bash
mb-cli card get 10 --full --format json
```

```json
{
  "id": 10,
  "name": "Weekly Active Users",
  "query_type": "native",
  "dataset_query": {
    "database": 1,
    "type": "native",
    "native": {
      "query": "SELECT count(DISTINCT user_id) AS active_users\nFROM events\nWHERE occurred_at >= CURRENT_DATE - 7"
    }
  }
}
```

`--full` is how you discover which tables and columns a card depends on.

### 4. Run the card

```bash
mb-cli card run 10 --format table
```

```
active_users
1843
```

If the card takes parameters, see the next recipe.

---

## Recipe 3 — Run a parameterized saved question

**Goal:** run a saved question that accepts parameters, without guessing
parameter names or values.

### 1. Discover the card's parameters

```bash
mb-cli card params 10 --format table
```

```
card_id     10
card_name   Weekly Active Users
query_type  native

Parameters
name           display_name   type    widget_type  required  default
last_x_days    Last X Days    number               true      30
min_events     Min Events     number               true      1

Run with:
  mb-cli card run 10 --param last_x_days=<value> --param min_events=<value>
```

If a card takes no parameters, `card params` says so and points you at plain
`card run`.

### 2. Run with parameters

Pass each parameter `name` as `--param name=value` (repeat the flag per
parameter):

```bash
mb-cli card run 10 --param last_x_days=7 --param min_events=1 --format table
```

```
active_users
612
```

Omitting `--param` uses each parameter's default (here, `last_x_days=30`).

### 3. When a parameter is wrong

A bad parameter key or value fails with a message that names the fix:

```bash
mb-cli card run 10 --param wrong_name=7
```

```
Error: parameterized query failed: check parameter keys and values, then run
'mb-cli card params 10' to list valid parameters (...)
```

Use `--error-format json` for a machine-readable error whose `suggestion`
field carries the same `card params` command.

---

## Recipe 4 — Analyze a dashboard

**Goal:** understand a dashboard's structure, the saved questions behind it,
and which cards depend on shared parameters.

### 1. Find the dashboard

```bash
mb-cli search "sales overview" --models dashboard --format table
```

```
id   name            description        model      database_id  table_id  collection_id  archived
298  Sales Overview  Company-wide KPIs  dashboard  0            0         <nil>          false
```

`dashboard list` browses all of them.

### 2. Inspect dashboard structure

```bash
mb-cli dashboard get 298 --format table
```

```
id           298
name         Sales Overview
description  Company-wide KPIs
archived     false

Tabs
id   name
166  Revenue
167  Customers

Parameters
id        name    slug    type
839f43ad  Region  region  string/=

Cards
[Revenue]
dashcard_id  card_id  name                 query_type  display
1201         463      Revenue by Region    native      bar
1202         464      Revenue by Month     native      line
```

### 3. Summarize dependencies

`dashboard analyze` rolls up the cards, parameters, and assumptions:

```bash
mb-cli dashboard analyze 298 --format table
```

```
dashboard_id       298
name               Sales Overview
description        Company-wide KPIs
total_dashcards    12
total_parameters   1
unique_base_cards  11
flagged_cards      3
```

`flagged_cards` counts cards with assumptions worth checking — the per-card
`flags` column (visible in the full output) calls out things like
`contains hardcoded filter literals`. Use `--format json` for the complete
per-dashcard breakdown.

### 4. List just the cards

```bash
mb-cli dashboard cards 298 --format table
```

```
dashcard_id  tab      card_id  name               query_type  display
1201         Revenue  463      Revenue by Region  native      bar
1202         Revenue  464      Revenue by Month   native      line
```

The `dashcard_id` and `card_id` pair is what you need to run a card in
[Recipe 5](#recipe-5--run-a-parameterized-dashboard-card).

---

## Recipe 5 — Run a parameterized dashboard card

**Goal:** execute one card on a dashboard with the dashboard's filters
applied.

### 1. List the dashboard's parameters

```bash
mb-cli dashboard params list 298 --format table
```

```
dashboard_id    298
dashboard_name  Sales Overview

Parameters
id        name    slug    type      mapped_cards
839f43ad  Region  region  string/=  8

Inspect a parameter's valid values with:
  mb-cli dashboard params values 298 <param>
Run a card with:
  mb-cli dashboard run-card 298 <dashcard-id> <card-id> --param <param>=<value>
```

`mapped_cards` shows how many cards a parameter filters — `0` means changing it
has no effect.

### 2. Discover valid parameter values

```bash
mb-cli dashboard params values 298 region --format table
```

For long value lists, filter with a search query:

```bash
mb-cli dashboard params search 298 region "eu" --format table
```

### 3. Run a dashboard card

Pass the dashboard ID, the `dashcard_id`, and the `card_id` from
[Recipe 4](#recipe-4--analyze-a-dashboard):

```bash
mb-cli dashboard run-card 298 1201 463 --param region="EU West" --format table
```

```
region   revenue_usd
EU West  248130
```

Without `--param`, the card runs with the dashboard's default filter state.

---

## Recipe 6 — PII-safe ad-hoc querying

**Goal:** answer a data question while keeping personal data out of the
terminal (and out of an agent's context).

PII redaction is **on by default**. Columns whose Metabase semantic type marks
them as personal data (`type/Email`, `type/Name`, `type/Phone`, and more) come
back as `[REDACTED]`.

### 1. Prefer `query filter` for simple lookups

`query filter` needs no SQL, resolves database and table names for you, and
avoids syntax mistakes:

```bash
mb-cli query filter --db Production --table 42 \
  --where "status=active" --limit 3 \
  --fields "id,email,first_name,status" --format table
```

```
id                                    email       first_name  status
user.a5e0be29-4702-4535-8ea7-3f979ab6  [REDACTED]  [REDACTED]  active
user.c2db17d9-cfbd-4043-8183-019a88f3  [REDACTED]  [REDACTED]  active
user.ced7e4cd-4243-4744-ae67-d9c9ca9f  [REDACTED]  [REDACTED]  active
```

The `email` and `first_name` values are redacted; the `id` stays visible so you
can still cross-reference records. If the table name is ambiguous, pass the
numeric table ID instead — the error message lists the candidates.

### 2. Use `query sql` for aggregations and joins

Aggregated columns rarely contain PII, so a `GROUP BY` is a safe way to explore
distributions. Always bound the result with `LIMIT` or `--limit`:

```bash
mb-cli query sql --db Production \
  --sql "SELECT status, count(*) AS users FROM users GROUP BY 1 ORDER BY 2 DESC" \
  --format table
```

```
status   users
active   31204
closed   12880
pending  4126
```

### 3. Know what redaction does not catch

Redaction relies on Metabase semantic types. A PII column with no semantic type
is **not** redacted, and derived columns from joins or `CASE` expressions can
lose the upstream type. For native SQL, alias a derived column to a known PII
field name so column-name enrichment catches it. See the
[README](../README.md#pii-redaction) for the full list of redacted types and
known gaps.

### 4. Exporting is blocked while redaction is on

`--export` (csv, json, xlsx) is refused when redaction is enabled, because raw
export bytes cannot be reliably redacted:

```
export is not supported when PII redaction is enabled (use JSON or table format instead)
```

**For agents:** do not disable redaction. Never pass `--redact-pii=false` or
set `MB_REDACT_PII=false`. Identify records by ID; if a human needs the actual
PII value, they can look it up directly in Metabase. Disabling redaction also
prints `Warning: PII redaction is disabled` to stderr on every invocation, so
the choice is always auditable.

---

## See also

- [README](../README.md) — full command reference and flag tables
- `mb-cli context` — the agent reference document, including command tables and
  the list of flags that do **not** exist
- [README › PII Redaction](../README.md#pii-redaction) — redaction scope,
  limits, and opt-out precedence
