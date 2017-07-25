package command

import (
	"github.com/mitchellh/cli"
	"testing"
)

func TestVersionCommand_implements(t *testing.T) {
	var _ cli.Command = &VersionCommand{}
}
