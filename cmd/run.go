package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/acmesoftwarellc/acme-cli/internal/config"
	"github.com/acmesoftwarellc/acme-cli/internal/dotenv"
	"github.com/acmesoftwarellc/acme-cli/internal/provider"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run [service | -- command...]",
	Short:              "Run a named service or direct command with secrets injected",
	DisableFlagParsing: true,
	RunE:               runRun,
}

func runRun(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		return cmd.Help()
	}

	cwd, _ := os.Getwd()
	cfg, cfgDir, err := config.Load(cwd)
	if err != nil {
		return err
	}

	loadEnvFile(cfg, cfgDir)

	p, err := provider.New(cfg)
	if err != nil {
		return err
	}

	environment := envOrCfg("ENVIRONMENT", cfg.Secrets.Environment)

	// Passthrough mode: acme run -- <cmd>
	if args[0] == "--" {
		if len(args) < 2 {
			return errors.New("expected a command after --")
		}
		return p.Exec(environment, args[1:], nil)
	}

	// Named service mode: acme run <service>
	serviceName := args[0]
	svc, ok := cfg.Services[serviceName]
	if !ok {
		return fmt.Errorf("service %q not defined in acme.toml\navailable: %s", serviceName, serviceList(cfg))
	}

	cmds := svc.GetCommands()
	if len(cmds) == 0 {
		return fmt.Errorf("service %q has no command or commands defined", serviceName)
	}

	if !svc.NeedsSecrets() {
		return execDirect(cmds, svc.Env)
	}

	// Run intermediate commands, exec the last one (replaces process)
	for i, rawCmd := range cmds {
		argv := config.ParseCommand(rawCmd)
		isLast := i == len(cmds)-1
		if isLast {
			return p.Exec(environment, argv, svc.Env)
		}
		if err := p.RunCmd(environment, argv, svc.Env); err != nil {
			exitWithError(err)
		}
	}
	return nil
}

// execDirect runs commands without secrets injection; the last command replaces the process.
func execDirect(cmds []string, extraEnv map[string]string) error {
	env := os.Environ()
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}

	for i, rawCmd := range cmds {
		argv := config.ParseCommand(rawCmd)
		isLast := i == len(cmds)-1

		if isLast {
			bin, err := exec.LookPath(argv[0])
			if err != nil {
				return fmt.Errorf("command not found: %s", argv[0])
			}
			return syscall.Exec(bin, argv, env)
		}

		c := exec.Command(argv[0], argv[1:]...)
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		c.Env = env
		if err := c.Run(); err != nil {
			exitWithError(err)
		}
	}
	return nil
}

// loadEnvFile reads env_file from disk and sets any vars not already in the environment.
// This lets `acme run` pick up INFISICAL_TOKEN etc. from the project .env when
// running outside Docker (where compose doesn't inject them).
func loadEnvFile(cfg *config.Config, cfgDir string) {
	envFile := cfg.Secrets.EnvFile
	if envFile == "" {
		envFile = ".env"
	}
	if !filepath.IsAbs(envFile) {
		envFile = filepath.Join(cfgDir, envFile)
	}
	for k, v := range dotenv.Load(envFile) {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}

func envOrCfg(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func serviceList(cfg *config.Config) string {
	names := make([]string, 0, len(cfg.Services))
	for k := range cfg.Services {
		names = append(names, k)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "(none)"
	}
	result := names[0]
	for _, n := range names[1:] {
		result += ", " + n
	}
	return result
}

func exitWithError(err error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
