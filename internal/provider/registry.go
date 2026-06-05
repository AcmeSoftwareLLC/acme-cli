package provider

import (
	"fmt"

	"github.com/acmesoftwarellc/acme-cli/internal/config"
	"github.com/acmesoftwarellc/acme-cli/internal/provider/infisical"
)

// New returns the Provider implementation for the provider named in cfg.
// Adding a new vendor: implement Provider, then add a case here.
func New(cfg *config.Config) (Provider, error) {
	switch cfg.Secrets.Provider {
	case "infisical", "":
		return infisical.New(cfg), nil
	default:
		return nil, fmt.Errorf("unknown secrets provider %q — supported: infisical", cfg.Secrets.Provider)
	}
}
