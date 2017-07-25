package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestSnapshotCommand_implements(t *testing.T) {
	var _ cli.Command = &SnapshotCommand{}
}

func TestSnapshotCommand_noTabs(t *testing.T) {
	assertNoTabs(t, new(SnapshotCommand))
}
