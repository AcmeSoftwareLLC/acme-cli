package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acmesoftwarellc/acme-cli/internal/config"
	"github.com/acmesoftwarellc/acme-cli/internal/provider"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Authenticate with the secrets provider and write the token to .env",
	Long: `Checks the current auth status, opens a browser login if expired,
then writes INFISICAL_TOKEN (and related config vars) to the project's env file.`,
	RunE: runEnv,
}

func runEnv(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	cfg, cfgDir, err := config.Load(cwd)
	if err != nil {
		return err
	}

	p, err := provider.New(cfg)
	if err != nil {
		return err
	}

	envFile := cfg.Secrets.EnvFile
	if !filepath.IsAbs(envFile) {
		envFile = filepath.Join(cfgDir, envFile)
	}

	environment := cfg.Secrets.Environment

	fmt.Printf("[acme env] provider=%s  env_file=%s  environment=%s\n",
		cfg.Secrets.Provider, envFile, environment)

	return p.WriteEnvFile(envFile, environment)
}
