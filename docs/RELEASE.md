# Release Verification

This document describes how to verify that the public install paths for `mb-cli` produce a working binary.

## Automated smoke tests

The [Install Smoke Tests](../.github/workflows/install-smoke-tests.yml) workflow runs on every push and pull request:

- `go install` from the local source tree builds an `mb-cli` binary and runs `mb-cli --help` and `mb-cli version`.
- GoReleaser builds release artifacts in snapshot mode, the Linux amd64 binary is executed, and the expected archives and checksum file are checked.
- The generated Homebrew formula is syntax-checked with `ruby -c`.

These tests do not require Metabase credentials.

## Manual release checks

After a release is published, verify the public install paths:

### Go install

```bash
go install github.com/andreagrandi/mb-cli/cmd/mb-cli@latest
$(go env GOPATH)/bin/mb-cli --help
$(go env GOPATH)/bin/mb-cli version
```

### Homebrew

```bash
brew update
brew install andreagrandi/tap/mb-cli
mb-cli --help
mb-cli version
```

### Binary download

1. Download the archive for your platform from the [releases page](https://github.com/andreagrandi/mb-cli/releases).
2. Extract the `mb-cli` binary.
3. Run `./mb-cli --help` and `./mb-cli version`.

### Expected output

Both `mb-cli --help` and `mb-cli version` must exit with status `0` and print the help text and version respectively.
