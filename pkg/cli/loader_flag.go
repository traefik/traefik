package cli

import (
	"fmt"

	"github.com/containous/traefik/pkg/config/flag"
	"github.com/containous/traefik/pkg/log"
)

// FlagLoader loads configuration from flags.
type FlagLoader struct{}

// Load loads the configuration.
func (f *FlagLoader) Load(args []string, cmd *Command) (bool, error) {
	if err := flag.Decode(args, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %v", err)
	}

	log.WithoutContext().Println("Configuration loaded from flags.")

	return true, nil
}
