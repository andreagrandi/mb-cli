# AGENTS.md - mb-cli Development Guide

## Project Structure

mb-cli follows the standard Go project layout:
```
mb-cli/
├── cmd/mb/main.go         # Application entry point
├── internal/
│   ├── cli/               # CLI command implementations (Cobra)
│   ├── client/            # Metabase API client
│   ├── config/            # Configuration management
│   ├── formatter/         # Output formatting (JSON, table)
│   └── version/           # Version information
├── tests/                 # Test files
└── Makefile               # Build targets
```

## Build/Test Commands

- `make build` - Build binary to bin/mb
- `make test` - Run all tests
- `make test-verbose` - Run tests with verbose output
- `make fmt` - Format code with gofmt
- `make vet` - Static analysis with go vet
- `make lint` - Run golangci-lint
- `make clean` - Remove build artifacts
- `make deps` - Download and tidy dependencies
- `make build-all` - Cross-platform builds

## Configuration

Environment variables (both required):
- `MB_HOST` - Metabase instance URL (e.g. `https://your-metabase-instance.com`)
- `MB_API_KEY` - Metabase API key

## Code Style Guidelines

- Use Go standard formatting (gofmt)
- Package names: lowercase, single word
- Types: PascalCase
- Functions/methods: PascalCase (exported), camelCase (unexported)
- Variables: camelCase
- Constants: PascalCase or ALL_CAPS
- Import ordering: standard library, third-party, local packages
- Use `fmt.Errorf` for error wrapping
- Always defer `resp.Body.Close()` after error check for HTTP responses
- Use table-driven tests with `tests := []struct{}` pattern
- Test files go in the `tests/` directory
- Use `httptest.NewServer` for HTTP client testing

## Smoke Testing

- After implementing API methods or CLI commands, always smoke test **every** method against the real Metabase API (`make build && ./bin/mb-cli <command>`)
- Do not consider a step complete until all endpoints have been verified end-to-end, not just unit tested

## Project Style

- When you generate or update the CHANGELOG.md, be concise
- New additions go in [Unreleased] section
- Don't bump version unless requested

## Changelog Updates

Every PR that introduces a user-visible change must add an entry under the
`## [Unreleased]` section of `CHANGELOG.md` before it is opened for review.
This is the same workflow LogBasset follows and the entry is what becomes the
public release note.

### When a changelog entry is required

Add an entry whenever a change affects what users observe or depend on:

- New, removed, or renamed commands, subcommands, flags, or environment
  variables
- Changes to command output (format, schema, default format, sort order,
  truncation)
- Changes to default behavior (e.g. PII redaction defaults, exit codes,
  validation, error messages users may script against)
- Bug fixes that change observable behavior
- Packaging and install changes (Homebrew formula, `go install` path,
  release artifacts, supported Go versions for distributed binaries)
- Security or privacy behavior (redaction, logging of sensitive values,
  authentication handling)
- Documentation changes that describe a new capability shipped in the same PR
  (the entry covers the capability, not the docs)

### When an entry is NOT required

Skip the changelog only for changes invisible to users:

- Internal refactors with no behavior change
- Test-only changes
- Pure CI / tooling tweaks that do not affect built artifacts
- Repo housekeeping (README typos, comments, formatting)
- Dependency bumps that do not change behavior (Dependabot PRs are exempt by
  default; if a bump changes behavior, add an entry manually)

If you are unsure, add an entry — over-documenting is cheaper than a missing
release note.

### How to write the entry

- Add a single bullet under `## [Unreleased]` at the top of `CHANGELOG.md`
- One line, present tense, user-facing wording ("Add `query filter`
  command", "Redact PII in native SQL results", "Fix Homebrew install on
  macOS")
- Reference the issue or PR number in parentheses when it adds context
  (e.g. `(#20)`); do not include personal information
- Do not bump the version or move entries out of `[Unreleased]` — that
  happens during the release process

### Opting out

The PR template has a "Changelog updated" checkbox. If a PR genuinely needs
no entry, tick the "No changelog entry needed" box and briefly say why. The
`changelog` CI check also honours the `no-changelog` label for the same
purpose.

## Release Process

When asked to create a new release, follow this exact sequence:

1. Ensure tests pass without errors:

```bash
make test
```

2. Set the release version in `internal/version/version.go`:

- Use the provided version if one is given.
- Otherwise bump the patch version (example: `0.1.0` -> `0.1.1`).
- Update `var Version = "..."` in `internal/version/version.go`.

3. Update `CHANGELOG.md`:

- Add a short bullet-point summary of changes since the last release.
- Follow the existing changelog format.

4. Commit the release-prep changes.
5. Push the release-prep commit.
6. Create the tag using the version from `internal/version/version.go`:

```bash
git tag v<version>
```

7. Push the tag:

```bash
git push origin v<version>
```

Notes:

- The tag push triggers `.github/workflows/release.yml`.
- Do not manually create a GitHub release before the workflow runs.
