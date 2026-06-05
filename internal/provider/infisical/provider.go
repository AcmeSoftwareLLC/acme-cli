package infisical

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/acmesoftwarellc/acme-cli/internal/config"
	"github.com/acmesoftwarellc/acme-cli/internal/dotenv"
)

// Provider implements provider.Provider using the Infisical CLI as a subprocess.
type Provider struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Provider {
	return &Provider{cfg: cfg}
}

// Login checks the current auth status and opens a browser login if needed.
func (p *Provider) Login() error {
	if p.isAuthenticated() {
		fmt.Println("Already authenticated.")
		return nil
	}
	fmt.Println("Opening browser login ...")
	cmd := exec.Command("infisical", "login", "--domain="+p.cfg.Secrets.Infisical.Domain)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WriteEnvFile fetches the personal user token and writes provider vars to .env.
func (p *Provider) WriteEnvFile(envFile, environment string) error {
	if err := p.Login(); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	out, err := exec.Command("infisical", "user", "get", "token", "--plain").Output()
	if err != nil {
		return fmt.Errorf("fetching token: %w", err)
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return errors.New("got empty token from infisical")
	}

	fmt.Printf("Writing token to %s...\n", envFile)
	return dotenv.Upsert(envFile, map[string]string{
		"ENVIRONMENT":          environment,
		"INFISICAL_API_URL":    p.cfg.Secrets.Infisical.Domain,
		"INFISICAL_PROJECT_ID": p.cfg.Secrets.Infisical.ProjectID,
		"INFISICAL_TOKEN":      token,
	})
}

// Exec replaces the current process with infisical run + the given command.
func (p *Provider) Exec(environment string, cmd []string, extraEnv map[string]string) error {
	token, err := p.resolveToken()
	if err != nil {
		return err
	}

	infisicalPath, err := exec.LookPath("infisical")
	if err != nil {
		return fmt.Errorf("infisical not found in PATH — run `acme setup` first")
	}

	env := p.mergeEnv(extraEnv)
	args := p.buildRunArgs(token, environment, cmd)
	return syscall.Exec(infisicalPath, args, env)
}

// RunCmd runs a single command via infisical run and waits for it to complete.
func (p *Provider) RunCmd(environment string, cmd []string, extraEnv map[string]string) error {
	token, err := p.resolveToken()
	if err != nil {
		return err
	}

	args := p.buildRunArgs(token, environment, cmd)
	c := exec.Command(args[0], args[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = p.mergeEnv(extraEnv)
	return c.Run()
}

// Install installs the Infisical CLI if not already present.
func (p *Provider) Install() error {
	if _, err := exec.LookPath("infisical"); err == nil {
		fmt.Println("Infisical CLI already installed.")
		return nil
	}
	fmt.Println("Installing Infisical CLI...")
	switch runtime.GOOS {
	case "darwin":
		return p.run("brew", "install", "infisical/get-cli/infisical")
	case "linux":
		return p.installLinux()
	default:
		return fmt.Errorf("unsupported OS %q — install manually: https://infisical.com/docs/cli/overview", runtime.GOOS)
	}
}

// resolveToken returns INFISICAL_TOKEN from the environment, deriving it via
// universal auth when INFISICAL_CLIENT_ID + INFISICAL_CLIENT_SECRET are set.
func (p *Provider) resolveToken() (string, error) {
	if clientID := os.Getenv("INFISICAL_CLIENT_ID"); clientID != "" {
		secret := os.Getenv("INFISICAL_CLIENT_SECRET")
		if secret == "" {
			return "", errors.New("INFISICAL_CLIENT_ID set but INFISICAL_CLIENT_SECRET is missing")
		}
		domain := envOrDefault("INFISICAL_API_URL", p.cfg.Secrets.Infisical.Domain)
		out, err := exec.Command("infisical", "login",
			"--method=universal-auth",
			"--client-id="+clientID,
			"--client-secret="+secret,
			"--domain="+domain,
			"--plain", "--silent",
		).Output()
		if err != nil {
			return "", fmt.Errorf("universal auth login: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	if token := os.Getenv("INFISICAL_TOKEN"); token != "" {
		return token, nil
	}
	return "", errors.New("no Infisical credentials found — set INFISICAL_TOKEN or run `acme env`")
}

func (p *Provider) buildRunArgs(token, environment string, cmd []string) []string {
	projectID := envOrDefault("INFISICAL_PROJECT_ID", p.cfg.Secrets.Infisical.ProjectID)
	domain := envOrDefault("INFISICAL_API_URL", p.cfg.Secrets.Infisical.Domain)

	args := []string{
		"infisical", "run",
		"--token", token,
		"--projectId", projectID,
		"--env", environment,
		"--domain=" + domain,
		"--",
	}
	return append(args, cmd...)
}

func (p *Provider) mergeEnv(extra map[string]string) []string {
	env := os.Environ()
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

type loginStatus struct {
	Sessions []struct {
		Status string `json:"status"`
		Token  struct {
			Exp int64 `json:"exp"`
		} `json:"token"`
	} `json:"sessions"`
}

func (p *Provider) isAuthenticated() bool {
	out, err := exec.Command("infisical", "login", "status", "--json").Output()
	if err != nil {
		return false
	}
	var s loginStatus
	if err := json.Unmarshal(out, &s); err != nil {
		return false
	}
	for _, session := range s.Sessions {
		if session.Status == "authenticated" && (session.Token.Exp == 0 || session.Token.Exp > time.Now().Unix()) {
			return true
		}
	}
	return false
}

func (p *Provider) installLinux() error {
	isDebian := func() bool {
		_, err := os.Stat("/etc/debian_version")
		return err == nil
	}
	isRPM := func() bool {
		_, err := os.Stat("/etc/redhat-release")
		return err == nil
	}

	switch {
	case isDebian():
		if err := p.run("sh", "-c", "curl -1sLf 'https://artifacts-cli.infisical.com/setup.deb.sh' | sudo -E bash"); err != nil {
			return err
		}
		return p.run("sudo", "apt-get", "install", "-y", "infisical")
	case isRPM():
		if err := p.run("sh", "-c", "curl -1sLf 'https://artifacts-cli.infisical.com/setup.rpm.sh' | sudo -E bash"); err != nil {
			return err
		}
		return p.run("sudo", "yum", "install", "-y", "infisical")
	default:
		return errors.New("unsupported Linux distribution — install manually: https://infisical.com/docs/cli/overview")
	}
}

func (p *Provider) run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
