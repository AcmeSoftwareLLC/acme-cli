package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acme",
	Short: "ACME developer tooling — secrets management and service runner",
	Long: `acme wraps your secrets provider (Infisical by default) to provide a
consistent interface across all ACME projects, regardless of language.

Configuration is read from acme.toml in the current or any parent directory.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(envCmd, runCmd, setupCmd)
}
