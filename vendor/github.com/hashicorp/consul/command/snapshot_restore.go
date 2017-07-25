package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/consul/command/base"
)

// SnapshotRestoreCommand is a Command implementation that is used to restore
// the state of the Consul servers for disaster recovery.
type SnapshotRestoreCommand struct {
	base.Command
}

func (c *SnapshotRestoreCommand) Help() string {
	helpText := `
Usage: consul snapshot restore [options] FILE

  Restores an atomic, point-in-time snapshot of the state of the Consul servers
  which includes key/value entries, service catalog, prepared queries, sessions,
  and ACLs.

  Restores involve a potentially dangerous low-level Raft operation that is not
  designed to handle server failures during a restore. This command is primarily
  intended to be used when recovering from a disaster, restoring into a fresh
  cluster of Consul servers.

  If ACLs are enabled, a management token must be supplied in order to perform
  snapshot operations.

  To restore a snapshot from the file "backup.snap":

    $ consul snapshot restore backup.snap

  For a full list of options and examples, please see the Consul documentation.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *SnapshotRestoreCommand) Run(args []string) int {
	flagSet := c.Command.NewFlagSet(c)

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	var file string

	args = flagSet.Args()
	switch len(args) {
	case 0:
		c.Ui.Error("Missing FILE argument")
		return 1
	case 1:
		file = args[0]
	default:
		c.Ui.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	// Open the file.
	f, err := os.Open(file)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error opening snapshot file: %s", err))
		return 1
	}
	defer f.Close()

	// Restore the snapshot.
	err = client.Snapshot().Restore(nil, f)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error restoring snapshot: %s", err))
		return 1
	}

	c.Ui.Info("Restored snapshot")
	return 0
}

func (c *SnapshotRestoreCommand) Synopsis() string {
	return "Restores snapshot of Consul server state"
}
