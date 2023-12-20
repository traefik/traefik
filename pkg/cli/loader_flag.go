package cli

import (
	"errors"
	"fmt"

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

	if deprecationNotice(args) {
		// An incompatible configuration is in use and need to be removed/adapted.
		return false, errors.New("incompatible static configuration detected")
	}

	if err := flag.Decode(args, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %w", err)
	}

	log.Print("Configuration loaded from flags")

	return true, nil
}

func deprecationNotice(args []string) bool {
	rawConfig := &rawConfiguration{}
	if err := flag.Decode(args, rawConfig); err != nil {
		log.Debug().Err(err).Msgf("failed to decode configuration from flags")
		return false
	}

	logger := log.With().Str("loader", "FLAG").Logger()
	return rawConfig.deprecationNotice(logger)
}
