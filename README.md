# acme-cli

The unified developer CLI for Acme Software projects. It provides a consistent interface for common development tasks across all Acme Software projects, regardless of language or stack. Secret management is the first supported feature, with more tooling to be added as our needs grow.

## Installation

```sh
go install github.com/acmesoftwarellc/acme-cli@latest
```

## Configuration

Place an `acme.toml` in your project root (or any parent directory — the CLI walks up to find it). See [acme.toml](acme.toml) in this repo for an annotated example covering all supported options.

## Commands

### `acme setup`

Install any required tooling (e.g. the configured secrets provider) if not already present.

```sh
acme setup
```

### `acme env`

Authenticate with the secrets provider and write the token to your env file.

```sh
acme env
```

Checks the current auth status, opens a browser login if the token is expired, then writes the provider token and related config vars to the project's env file (`.env` by default).

### `acme run`

Run a named service or an arbitrary command with secrets injected into the environment.

```sh
# Run a named service from acme.toml
acme run api

# Run an arbitrary command with secrets injected
acme run -- npm run dev
```

The last command in a service's `commands` list replaces the current process (`exec`). Earlier commands run as subprocesses and must exit 0 before the next one starts.

Services with `secrets = false` skip secret injection entirely and run directly.

## Environment variables

| Variable      | Description                                              |
|---------------|----------------------------------------------------------|
| `ENVIRONMENT` | Overrides `secrets.environment` from `acme.toml` at runtime |

## Contributing

`acme-cli` is the shared developer toolbelt for all Acme Software projects. New commands and integrations are welcome — follow the patterns below to keep things consistent.

### Getting started

```sh
git clone https://github.com/acmesoftwarellc/acme-cli
cd acme-cli
go build -o acme .
./acme --help
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
