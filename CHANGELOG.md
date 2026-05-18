# Changelog

## [Unreleased]

- Add `MB_SESSION_TOKEN` as an alternative authentication method for users without admin access to mint an API key; mutually exclusive with `MB_API_KEY` (#9)
- Document changelog update workflow in `AGENTS.md` and add a pull request template prompting contributors to update `CHANGELOG.md` for user-visible changes (#20)

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
