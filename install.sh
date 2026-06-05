#!/bin/sh
set -e

VERSION="${ACME_VERSION:-latest}"
REPO="github.com/AcmeSoftwareLLC/acme-cli"

die() {
    printf 'Error: %s\n' "$1" >&2
    exit 1
}

# Go must be present before anything else
command -v go >/dev/null 2>&1 || die "Go is not installed. Install it from https://go.dev/dl/ and re-run."

GOPATH_BIN="$(go env GOPATH)/bin"
BINARY="$GOPATH_BIN/acme-cli"

if [ -f "$BINARY" ]; then
    CURRENT=$("$BINARY" version 2>/dev/null || echo "unknown")
    if [ "$VERSION" = "latest" ]; then
        printf 'acme-cli is already installed (%s). To upgrade, set ACME_VERSION=latest and re-run with FORCE=1.\n' "$CURRENT"
        exit 0
    fi
fi

printf 'Installing acme-cli@%s...\n' "$VERSION"

if ! go install "${REPO}@${VERSION}" 2>&1; then
    die "go install failed. Check your network connection and that '${REPO}@${VERSION}' exists."
fi

if [ ! -f "$BINARY" ]; then
    die "Installation appeared to succeed but binary not found at $BINARY."
fi

INSTALLED=$("$BINARY" version 2>/dev/null || echo "unknown")
printf 'Installed acme-cli %s to %s\n' "$INSTALLED" "$BINARY"

# Warn if GOPATH/bin is not on PATH
case ":$PATH:" in
    *":$GOPATH_BIN:"*) ;;
    *)
        printf '\nWarning: %s is not in your PATH.\n' "$GOPATH_BIN"
        printf 'Add this to your shell profile (~/.zshrc or ~/.bashrc) and restart your terminal:\n'
        printf '  export PATH="%s:$PATH"\n' "$GOPATH_BIN"
        ;;
esac
