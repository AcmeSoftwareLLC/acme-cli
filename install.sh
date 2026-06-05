#!/bin/sh
set -e

VERSION="${ACME_VERSION:-latest}"
REPO="github.com/acmesoftwarellc/acme-cli"

die() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

warn_path() {
    case ":$PATH:" in
        *":$GOPATH_BIN:"*) ;;
        *)
            printf '\nWarning: %s is not in your PATH.\n' "$GOPATH_BIN"
            printf 'Add this to your shell profile (~/.zshrc or ~/.bashrc) and restart your terminal:\n'
            printf '  export PATH="%s:$PATH"\n' "$GOPATH_BIN"
            ;;
    esac
}

# Go must be present before anything else
command -v go >/dev/null 2>&1 || die "Go is not installed. Install it from https://go.dev/dl/ and re-run."

GOPATH_BIN="$(go env GOPATH)/bin"
BINARY="$GOPATH_BIN/acme-cli"

rm -f "$BINARY"

printf 'Installing acme-cli@%s...\n' "$VERSION"

if ! go install "${REPO}@${VERSION}" 2>&1; then
    die "go install failed. Check your network connection and that '${REPO}@${VERSION}' exists."
fi

if [ ! -f "$BINARY" ]; then
    die "Installation appeared to succeed but binary not found at $BINARY."
fi

printf 'Installed acme-cli to %s\n' "$BINARY"
warn_path
