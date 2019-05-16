package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/containous/traefik/pkg/config/env"
	"github.com/containous/traefik/pkg/log"
)

// EnvLoader loads configuration from environment variables.
type EnvLoader struct{}

// Load loads the configuration.
func (f *EnvLoader) Load(_ []string, cmd *Command) (bool, error) {
	environ := os.Environ()

	var found bool
	for _, value := range environ {
		if strings.HasPrefix(value, "TRAEFIK_") {
			found = true
			break
		}
	}

	if !found {
		return false, nil
	}

	if err := env.Decode(environ, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %v", err)
	}

	log.WithoutContext().Println("Configuration loaded from environment variables.")

	return true, nil
}
