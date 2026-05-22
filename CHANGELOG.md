# Changelog

## [Unreleased]

## [0.3.0] - 2026-05-22

- Add `docs/RECIPES.md` with worked end-to-end recipes for schema discovery, saved-question inspection, dashboard analysis, parameterized card/dashboard runs, and PII-safe querying, linked from the README (#16)
- Expand the `context` agent document with end-to-end workflows for schema, dashboard, and saved-question exploration plus an explicit safe-querying section (#15)
- Add `card params` and `dashboard params list` commands to discover the parameters a saved question or dashboard accepts before running it; failed parameterized runs now point to the exact discovery command (#14)
- Add `collection` commands (`list`, `get`, `items`) to browse collections and discover the cards, dashboards, and nested collections inside them, with `--models` filtering and `root` collection support (#13)
- Classify `--error-format json` errors from typed errors instead of message substrings, making structured error output reliable when wording changes (#12)
- Add `--timeout` flag and propagate request contexts so commands abort cleanly on Ctrl+C, with clear timeout and cancellation errors (#11)
- Add `MB_SESSION_TOKEN` as an alternative authentication method for users without admin access to mint an API key; mutually exclusive with `MB_API_KEY` (#9)
- Document changelog update workflow in `AGENTS.md` and add a pull request template prompting contributors to update `CHANGELOG.md` for user-visible changes (#20)
- Move main package to `cmd/mb-cli` so `go install github.com/andreagrandi/mb-cli/cmd/mb-cli@latest` produces an `mb-cli` binary that matches the documented command name (#10)
- Expand PII redaction documentation with covered commands, semantic-type enrichment behavior, known gaps, and export opt-out instructions (#17)
- Enrich semantic types on non-parameterized `card run` results so PII columns from native saved questions are redacted consistently with the parameterized path (#17)

## [0.2.0] - 2026-03-12

- Add dashboard inspection commands for listing dashboards, viewing tabs and dashcards, and summarizing dashboard dependencies
- Expand card inspection with `card get --full`, dashboard parameter lookup, and parameterized card/dashboard execution
- Improve dashboard query safety with clearer error messages, redaction-aware parameter handling, and updated documentation/tests

## [0.1.3] - 2026-03-05

- PII redaction enabled by default: query result columns with Metabase PII semantic types (Email, Name, Phone, etc.) are replaced with `[REDACTED]`
- Add `--redact-pii` global flag and `MB_REDACT_PII` environment variable
- Block `--export` when PII redaction is enabled
- Enrich native SQL query results with field semantic types from database metadata
- Update agent context document with PII redaction directive

## [0.1.2] - 2026-03-05

- Add `query filter` command for structured MBQL queries with `--where` field filters
- Table and field name resolution (case-insensitive substring match)
- Agent-friendly enhancements: `context`, `schema`, TTY auto-detection, `--error-format json`, input validation, `--fields` filtering

## [0.1.1] - 2026-03-05

- Fix Homebrew installation: switch from cask to formula to avoid macOS Gatekeeper blocking
- Add project logo to README

## [0.1.0] - 2026-03-05

- Database commands: list, get, metadata, fields, schemas, schema
- Table commands: list, get, metadata, fks, data
- Field commands: get, summary, values
- SQL query execution with database name resolution
- Card (saved questions) commands: list, get, run
- Search command with model filtering
- JSON and table output formatters
- CI/CD workflows and Dependabot configuration
- GoReleaser configuration for automated releases
