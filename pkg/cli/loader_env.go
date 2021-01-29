package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/env"
	tcmd "github.com/traefik/traefik/v2/cmd"
	"github.com/traefik/traefik/v2/pkg/log"
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
		log.WithoutContext().Debug("environment variables", strings.Join(vars, ", "))
		return false, fmt.Errorf("failed to decode configuration from environment variables: %w ", err)
	}

	// Checks if the newly loaded configuration specifies JSON logging, and if so sets the format so that the configuration loaded message will be in JSON
	configuration := cmd.Configuration.(*tcmd.TraefikCmdConfiguration)
	if configuration.Log != nil && configuration.Log.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	log.WithoutContext().Println("Configuration loaded from environment variables.")

	return true, nil
}
