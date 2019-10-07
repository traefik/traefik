package cli

import (
	"fmt"

	"github.com/containous/traefik/v2/pkg/config/flag"
	"github.com/containous/traefik/v2/pkg/log"
)

// FlagLoader loads configuration from flags.
type FlagLoader struct{}

// Load loads the command's configuration from flag arguments.
func (*FlagLoader) Load(args []string, cmd *Command) (bool, error) {
	if len(args) == 0 {
		return false, nil
	}

	if err := flag.Decode(args, cmd.Configuration); err != nil {
		return false, fmt.Errorf("failed to decode configuration from flags: %v", err)
	}

	log.WithoutContext().Println("Configuration loaded from flags.")

	return true, nil
}
