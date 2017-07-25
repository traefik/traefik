package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/snapshot"
)

// SnapshotSaveCommand is a Command implementation that is used to save the
// state of the Consul servers for disaster recovery.
type SnapshotSaveCommand struct {
	base.Command
}

func (c *SnapshotSaveCommand) Help() string {
	helpText := `
Usage: consul snapshot save [options] FILE

  Retrieves an atomic, point-in-time snapshot of the state of the Consul servers
  which includes key/value entries, service catalog, prepared queries, sessions,
  and ACLs.

  If ACLs are enabled, a management token must be supplied in order to perform
  snapshot operations.

  To create a snapshot from the leader server and save it to "backup.snap":

    $ consul snapshot save backup.snap

  To create a potentially stale snapshot from any available server (useful if no
  leader is available):

    $ consul snapshot save -stale backup.snap

  For a full list of options and examples, please see the Consul documentation.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *SnapshotSaveCommand) Run(args []string) int {
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

	// Take the snapshot.
	snap, qm, err := client.Snapshot().Save(&api.QueryOptions{
		AllowStale: c.Command.HTTPStale(),
	})
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error saving snapshot: %s", err))
		return 1
	}
	defer snap.Close()

	// Save the file.
	f, err := os.Create(file)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error creating snapshot file: %s", err))
		return 1
	}
	if _, err := io.Copy(f, snap); err != nil {
		f.Close()
		c.Ui.Error(fmt.Sprintf("Error writing snapshot file: %s", err))
		return 1
	}
	if err := f.Close(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error closing snapshot file after writing: %s", err))
		return 1
	}

	// Read it back to verify.
	f, err = os.Open(file)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error opening snapshot file for verify: %s", err))
		return 1
	}
	if _, err := snapshot.Verify(f); err != nil {
		f.Close()
		c.Ui.Error(fmt.Sprintf("Error verifying snapshot file: %s", err))
		return 1
	}
	if err := f.Close(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error closing snapshot file after verify: %s", err))
		return 1
	}

	c.Ui.Info(fmt.Sprintf("Saved and verified snapshot to index %d", qm.LastIndex))
	return 0
}

func (c *SnapshotSaveCommand) Synopsis() string {
	return "Saves snapshot of Consul server state"
}
