# acme-cli

The unified developer CLI for Acme Software projects. It provides a consistent interface for common development tasks across all Acme Software projects, regardless of language or stack. Secret management is the first supported feature, with more tooling to be added as our needs grow.

## Installation

```sh
# Using go install
go install github.com/acmesoftwarellc/acme-cli@latest

# Using the install script
curl -fsSL https://raw.githubusercontent.com/acmesoftwarellc/acme-cli/main/install.sh | sh

# Install a specific version
ACME_VERSION=v1.2.3 curl -fsSL https://raw.githubusercontent.com/acmesoftwarellc/acme-cli/main/install.sh | sh
```

Both methods install the binary as `acme-cli` into `$(go env GOPATH)/bin`. Make sure that directory is in your `PATH`.

## Configuration

Place an `acme.toml` in your project root (or any parent directory — the CLI walks up to find it):

```toml
[secrets]
provider    = "infisical"           # default: "infisical"
env_file    = ".env"                # default: ".env"
environment = "local"               # default: "local"

[secrets.infisical]
project_id = "your-infisical-project-id"
domain     = "https://app.infisical.com"  # default

# A service that needs secrets injected (default behaviour)
[services.api]
command = "go run ./cmd/api"

# A service with multiple steps — earlier commands run as subprocesses,
# the last one replaces the current process (exec)
[services.worker]
commands = [
  "go generate ./...",
  "go run ./cmd/worker",
]

# A service with extra environment variables
[services.migrate]
command = "go run ./cmd/migrate"
[services.migrate.env]
DB_POOL_SIZE = "1"

# A service that does not need secrets (skips provider injection entirely)
[services.frontend]
command = "npm run dev"
secrets = false
```

## Commands

### `acme-cli setup`

Install any required tooling (e.g. the configured secrets provider) if not already present.

```sh
acme-cli setup
```

### `acme-cli env`

Authenticate with the secrets provider and write the token to your env file.

```sh
acme-cli env
```

Checks the current auth status, opens a browser login if the token is expired, then writes the provider token and related config vars to the project's env file (`.env` by default).

### `acme-cli run`

Run a named service or an arbitrary command with secrets injected into the environment.

```sh
# Run a named service from acme.toml
acme-cli run api

# Run an arbitrary command with secrets injected
acme-cli run -- npm run dev
```

The last command in a service's `commands` list replaces the current process (`exec`). Earlier commands run as subprocesses and must exit 0 before the next one starts.

Services with `secrets = false` skip secret injection entirely and run directly.

## Environment variables

| Variable | Description |
|---|---|
| `ENVIRONMENT` | Overrides `secrets.environment` from `acme.toml` at runtime |
| `INFISICAL_TOKEN` | Pre-set auth token — skips interactive login |
| `INFISICAL_CLIENT_ID` | Universal auth client ID for CI/CD and non-interactive environments |
| `INFISICAL_CLIENT_SECRET` | Universal auth client secret (used with `INFISICAL_CLIENT_ID`) |

## Contributing

`acme-cli` is the shared developer toolbelt for all Acme Software projects. New commands and integrations are welcome — follow the patterns below to keep things consistent.

### Getting started

```sh
git clone https://github.com/acmesoftwarellc/acme-cli
cd acme-cli
go build -o acme-cli .
./acme-cli --help
```

### Adding a new command

1. Create `cmd/<name>.go` and define a `*cobra.Command` (see [cmd/run.go](cmd/run.go) as a reference).
2. Register it in [cmd/root.go](cmd/root.go) via `rootCmd.AddCommand(...)`.
3. Put any non-trivial logic in a new `internal/<name>/` package, not in the `cmd` layer.

### Adding a new secrets provider

1. Create `internal/provider/<name>/provider.go` implementing the `provider.Provider` interface (see [internal/provider/provider.go](internal/provider/provider.go)).
2. Register the provider name in [internal/provider/registry.go](internal/provider/registry.go).
3. Add the corresponding config fields to `SecretsConfig` in [internal/config/config.go](internal/config/config.go) and apply any defaults in `applyDefaults`.

### Guidelines

- Keep the `cmd` layer thin — it parses flags and delegates to `internal`.
- `acme.toml` is the source of truth for project config; avoid adding hidden conventions that aren't reflected there.
- New commands should work without an `acme.toml` where it makes sense (see how `setup` handles a missing config file).
- No external dependencies without discussion — the binary should stay small and easy to install.
