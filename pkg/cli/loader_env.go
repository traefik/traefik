package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/containous/traefik/pkg/config/env"
	"github.com/containous/traefik/pkg/log"
)

// EnvLoader loads a configuration from all the environment variables prefixed with "TRAEFIK_".
type EnvLoader struct{}

// Load loads the command's configuration from the environment variables.
func (e *EnvLoader) Load(_ []string, cmd *Command) (bool, error) {
	return e.load(os.Environ(), cmd)
}

func (*EnvLoader) load(environ []string, cmd *Command) (bool, error) {
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
		return false, fmt.Errorf("failed to decode configuration from environment variables: %v", err)
	}

	log.WithoutContext().Println("Configuration loaded from environment variables.")

	return true, nil
}
