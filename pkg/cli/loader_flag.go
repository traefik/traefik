package cli

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/flag"
	tcmd "github.com/traefik/traefik/v2/cmd"
	"github.com/traefik/traefik/v2/pkg/log"
)

// FlagLoader loads configuration from flags.
type FlagLoader struct{}

// Load loads the command's configuration from flag arguments.
func (*FlagLoader) Load(args []string, cmd *cli.Command) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	if err := flag.Decode(args, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %w", err)
	}

	// Checks if the newly loaded configuration specifies JSON logging, and if so sets the format so that the configuration loaded message will be in JSON
	configuration := cmd.Configuration.(*tcmd.TraefikCmdConfiguration)
	if configuration.Log != nil && configuration.Log.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	log.WithoutContext().Println("Configuration loaded from flags.")

	return true, nil
}
