package cmd

import (
	"fmt"
	"os"

	"github.com/acmesoftwarellc/acme-cli/internal/config"
	"github.com/acmesoftwarellc/acme-cli/internal/provider"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install the secrets provider CLI (e.g. Infisical) if not already present",
	RunE:  runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	cfg, _, err := config.Load(cwd)
	if err != nil {
		// No acme.toml found — default to infisical
		fmt.Println("No acme.toml found, defaulting to infisical provider.")
		cfg = &config.Config{}
		cfg.Secrets.Provider = "infisical"
	}

	p, err := provider.New(cfg)
	if err != nil {
		return err
	}
	return p.Install()
}
