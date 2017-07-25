package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestKVCommand_implements(t *testing.T) {
	var _ cli.Command = &KVCommand{}
}

func TestKVCommand_noTabs(t *testing.T) {
	assertNoTabs(t, new(KVCommand))
}
