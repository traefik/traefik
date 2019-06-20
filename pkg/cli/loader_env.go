package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/containous/traefik/pkg/config/env"
	"github.com/containous/traefik/pkg/log"
)

// EnvLoader loads a configuration from all the environment variables prefixed with "TRAEFIK_".
type EnvLoader struct{}

// Load loads the command's configuration from the environment variables.
func (e *EnvLoader) Load(_ []string, cmd *Command) (bool, error) {
	if ok, _ := strconv.ParseBool(os.Getenv("TRAEFIK_ENABLE_ENV_VARS")); !ok {
		return false, nil
	}

	if err := env.Decode(os.Environ(), cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from environment variables: %v", err)
	}

	log.WithoutContext().Println("Configuration loaded from environment variables.")

	return true, nil
}
