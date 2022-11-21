package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/env"
)

// EnvLoader loads a configuration from all the environment variables prefixed with "TRAEFIK_".
type EnvLoader struct{}

// Load loads the command's configuration from the environment variables.
func (e *EnvLoader) Load(_ []string, cmd *cli.Command) (bool, error) {
	vars := env.FindPrefixedEnvVars(os.Environ(), env.DefaultNamePrefix, cmd.Configuration)
	if len(vars) == 0 {
		return false, nil
	}

	if err := env.Decode(vars, env.DefaultNamePrefix, cmd.Configuration); err != nil {
		log.Debug().Msgf("environment variables: %s", strings.Join(vars, ", "))
		return false, fmt.Errorf("failed to decode configuration from environment variables: %w", err)
	}

	log.Print("Configuration loaded from environment variables")

	return true, nil
}
