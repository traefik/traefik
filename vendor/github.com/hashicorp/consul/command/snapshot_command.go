package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

// SnapshotCommand is a Command implementation that just shows help for
// the subcommands nested below it.
type SnapshotCommand struct {
	Ui cli.Ui
}

func (c *SnapshotCommand) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *SnapshotCommand) Help() string {
	helpText := `
Usage: consul snapshot <subcommand> [options] [args]

  This command has subcommands for saving, restoring, and inspecting the state
  of the Consul servers for disaster recovery. These are atomic, point-in-time
  snapshots which include key/value entries, service catalog, prepared queries,
  sessions, and ACLs.

  If ACLs are enabled, a management token must be supplied in order to perform
  snapshot operations.

  Create a snapshot:

      $ consul snapshot save backup.snap

  Restore a snapshot:

      $ consul snapshot restore backup.snap

  Inspect a snapshot:

      $ consul snapshot inspect backup.snap


  For more examples, ask for subcommand help or view the documentation.

`
	return strings.TrimSpace(helpText)
}

func (c *SnapshotCommand) Synopsis() string {
	return "Saves, restores and inspects snapshots of Consul server state"
}
