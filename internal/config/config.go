package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultDomain = "https://app.infisical.com"

type Config struct {
	Secrets  SecretsConfig            `toml:"secrets"`
	Services map[string]ServiceConfig `toml:"services"`
}

type SecretsConfig struct {
	Provider    string          `toml:"provider"`
	EnvFile     string          `toml:"env_file"`
	Environment string          `toml:"environment"`
	Infisical   InfisicalConfig `toml:"infisical"`
}

type InfisicalConfig struct {
	ProjectID string `toml:"project_id"`
	Domain    string `toml:"domain"`
}

type ServiceConfig struct {
	Command  string            `toml:"command"`
	Commands []string          `toml:"commands"`
	Env      map[string]string `toml:"env"`
	Secrets  *bool             `toml:"secrets"`
}

func (s ServiceConfig) NeedsSecrets() bool {
	if s.Secrets == nil {
		return true
	}
	return *s.Secrets
}

func (s ServiceConfig) GetCommands() []string {
	if len(s.Commands) > 0 {
		return s.Commands
	}
	if s.Command != "" {
		return []string{s.Command}
	}
	return nil
}

// Load reads acme.toml from the given directory (walks up to find it).
func Load(startDir string) (*Config, string, error) {
	dir := startDir
	for {
		path := filepath.Join(dir, "acme.toml")
		if _, err := os.Stat(path); err == nil {
			var cfg Config
			if _, err := toml.DecodeFile(path, &cfg); err != nil {
				return nil, "", fmt.Errorf("parsing %s: %w", path, err)
			}
			applyDefaults(&cfg)
			return &cfg, dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, "", fmt.Errorf("acme.toml not found in %s or any parent directory", startDir)
}

func applyDefaults(cfg *Config) {
	if cfg.Secrets.Provider == "" {
		cfg.Secrets.Provider = "infisical"
	}
	if cfg.Secrets.EnvFile == "" {
		cfg.Secrets.EnvFile = ".env"
	}
	if cfg.Secrets.Environment == "" {
		cfg.Secrets.Environment = "local"
	}
	if cfg.Secrets.Infisical.Domain == "" {
		cfg.Secrets.Infisical.Domain = DefaultDomain
	}
}

// ParseCommand splits a shell-style command string into argv.
// Handles basic quoting but delegates complex cases to the shell.
func ParseCommand(raw string) []string {
	return strings.Fields(raw)
}
