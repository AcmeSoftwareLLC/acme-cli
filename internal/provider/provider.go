package provider

// Provider abstracts a secrets manager backend.
// Adding a new vendor: implement this interface + register in registry.go.
type Provider interface {
	// Login performs the interactive dev-time auth flow (browser-based or credential-based).
	Login() error

	// WriteEnvFile writes the auth token and provider config to the target .env file.
	WriteEnvFile(envFile, environment string) error

	// Exec replaces the current process with the command wrapped in secrets injection.
	// For services with secrets=false the Provider implementation should exec the command directly.
	// This function must not return on success (it calls syscall.Exec or equivalent).
	Exec(environment string, cmd []string, extraEnv map[string]string) error

	// RunCmd runs a command with secrets injection and waits for completion.
	// Used for multi-step sequences where intermediate commands must not replace the process.
	RunCmd(environment string, cmd []string, extraEnv map[string]string) error

	// Install installs the provider's required CLI tooling if not already present.
	Install() error
}
