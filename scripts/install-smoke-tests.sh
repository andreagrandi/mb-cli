#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

GO_BIN="${GO_BIN:-go}"
GORELEASER_BIN="${GORELEASER_BIN:-goreleaser}"

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

echo "==> Smoke testing go install from local source..."
"$GO_BIN" install ./cmd/mb-cli
MBCLI_BIN=$("$GO_BIN" env GOPATH)/bin/mb-cli
test -x "$MBCLI_BIN" || fail "mb-cli binary not found at $MBCLI_BIN"
"$MBCLI_BIN" --help >/dev/null || fail "mb-cli --help failed"
"$MBCLI_BIN" version >/dev/null || fail "mb-cli version failed"
echo "    go install smoke test passed"

if command -v "$GORELEASER_BIN" >/dev/null 2>&1; then
	echo "==> Smoke testing release artifacts with GoReleaser snapshot..."
	"$GORELEASER_BIN" release --snapshot --clean

	GOOS=$("$GO_BIN" env GOOS)
	GOARCH=$("$GO_BIN" env GOARCH)
	mapfile -d '' -t candidates < <(find dist -type f \( -name 'mb-cli' -o -name 'mb-cli.exe' \) -path "*mb-cli_${GOOS}_${GOARCH}_*" -print0)
	test ${#candidates[@]} -gt 0 || fail "no release binary found for $GOOS/$GOARCH"
	BINARY=${candidates[0]}
	test -x "$BINARY" || fail "release binary is not executable: $BINARY"
	"$BINARY" --help >/dev/null || fail "release binary --help failed"
	"$BINARY" version >/dev/null || fail "release binary version failed"

	test -f dist/checksums.txt || fail "checksums.txt not found"
	echo "    release artifacts smoke test passed"
else
	echo "    skipping GoReleaser smoke test (goreleaser not installed)"
fi

echo "==> All install smoke tests passed"
