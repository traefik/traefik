package cli

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/flag"
)

// FlagLoader loads configuration from flags.
type FlagLoader struct{}

// Load loads the command's configuration from flag arguments.
func (*FlagLoader) Load(args []string, cmd *cli.Command) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	if cmd.Name == "healthcheck" {
		for _, a := range args {
			if strings.HasPrefix(a, "--url") {
				log.Print("Skipping flag load for healthcheck (url provided)")
				return false, nil
			}
		}
	}

	if err := flag.Decode(args, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %w", err)
	}

	log.Print("Configuration loaded from flags")

	return true, nil
}
