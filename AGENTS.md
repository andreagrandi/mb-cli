# AGENTS.md - mb-cli Development Guide

## Session Start Workflow

Before making code or documentation changes in this repo:

1. Switch back to `master`.
2. Pull the latest changes with `git pull --ff-only`.
3. Create a new branch with a short, descriptive name related to the feature being added or the bug being fixed.
4. Make the requested changes on that branch.

Do not start work from an old feature branch unless the user explicitly asks to continue that branch.

## Adding a Ticket, Issue, or Bug

When the user asks to add a "ticket", "issue", or "bug" to this repo, all of the
steps below are required — creating the GitHub issue alone is not enough.

**1. Create the issue** with the repo label and an area label:

```bash
gh issue create --repo andreagrandi/mb-cli \
  --title "<concise title>" \
  --body "<description>" \
  --label "mb-cli" \
  --label "<area>"
```

- Always apply the `mb-cli` label — every work item in this repo carries it.
- Add the matching area label when one exists: `feature`, `ux`, `agent`,
  `docs`, `security`, `testing`, `release`. The `Reliability` and `Packaging`
  areas have no label; for those, set only the project Area field (step 3).
- Add a type label when it fits: `bug` (bug report), `enhancement` (feature
  request), or `documentation` (docs-only work).

**2. Add the issue to the "CLI Tools" project** and capture the item ID
(https://github.com/users/andreagrandi/projects/1):

```bash
ITEM_ID=$(gh project item-add 1 --owner andreagrandi \
  --url <issue-url> --format json --jq .id)
```

**3. Set Priority, Area, and Status** on the project item. The project and
field IDs are identical for every repo on this board:

- Project ID: `PVT_kwHOAAm1584BYDlZ`
- Priority — field `PVTSSF_lAHOAAm1584BYDlZzhTMDck`:
  High `ed3787e3`, Medium `3e3ea407`, Low `994234f4`
- Area — field `PVTSSF_lAHOAAm1584BYDlZzhTMDco`:
  Reliability `6595432d`, Packaging `6895c50a`, UX `2bc024bb`,
  Testing `0d5bc016`, Feature `6390f97d`, Agent `3a2d6f7e`,
  Docs `f5c50514`, Security `062d12a3`, Release `b344aeab`
- Status — field `PVTSSF_lAHOAAm1584BYDlZzhTMDNQ`:
  Todo `f75ad846`, In Progress `47fc9ee4`, Done `98236657`

```bash
# Priority — always set it
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDck \
  --single-select-option-id <priority-option-id>

# Area — match the area label from step 1
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDco \
  --single-select-option-id <area-option-id>

# Status — new tickets start as Todo
gh project item-edit --id "$ITEM_ID" --project-id PVT_kwHOAAm1584BYDlZ \
  --field-id PVTSSF_lAHOAAm1584BYDlZzhTMDNQ \
  --single-select-option-id f75ad846
```

If the user does not state a priority or area, ask before creating the issue.
Follow the conventions of existing project issues — do not invent new labels or
fields.

## Project Structure

mb-cli follows the standard Go project layout:
```
mb-cli/
├── cmd/mb-cli/main.go     # Application entry point
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

- `make build` - Build binary to bin/mb-cli
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
no entry, tick the "No changelog entry needed" box and briefly say why.

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

8. After the release workflow finishes, verify the public install paths:

- Follow the manual release checks in `docs/RELEASE.md`.
- Confirm `go install`, Homebrew, and binary downloads all produce a working `mb-cli --help` and `mb-cli version`.

Notes:

- The tag push triggers `.github/workflows/release.yml`.
- Do not manually create a GitHub release before the workflow runs.
