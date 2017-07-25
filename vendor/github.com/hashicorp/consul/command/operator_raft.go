package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/command/base"
)

type OperatorRaftCommand struct {
	base.Command
}

func (c *OperatorRaftCommand) Help() string {
	helpText := `
Usage: consul operator raft <subcommand> [options]

The Raft operator command is used to interact with Consul's Raft subsystem. The
command can be used to verify Raft peers or in rare cases to recover quorum by
removing invalid peers.

Subcommands:

    list-peers     Display the current Raft peer configuration
    remove-peer    Remove a Consul server from the Raft configuration

`

	return strings.TrimSpace(helpText)
}

func (c *OperatorRaftCommand) Synopsis() string {
	return "Provides cluster-level tools for Consul operators"
}

func (c *OperatorRaftCommand) Run(args []string) int {
	if result := c.raft(args); result != nil {
		c.Ui.Error(result.Error())
		return 1
	}
	return 0
}

// raft handles the raft subcommands.
func (c *OperatorRaftCommand) raft(args []string) error {
	f := c.Command.NewFlagSet(c)

	// Parse verb arguments.
	var listPeers, removePeer bool
	f.BoolVar(&listPeers, "list-peers", false,
		"If this flag is provided, the current Raft peer configuration will be "+
			"displayed. If the cluster is in an outage state without a leader, you may need "+
			"to set -stale to 'true' to get the configuration from a non-leader server.")
	f.BoolVar(&removePeer, "remove-peer", false,
		"If this flag is provided, the Consul server with the given -address will be "+
			"removed from the Raft configuration.")

	// Parse other arguments.
	var address string
	f.StringVar(&address, "address", "",
		"The address to remove from the Raft configuration.")

	// Leave these flags for backwards compatibility, but hide them
	// TODO: remove flags/behavior from this command in Consul 0.9
	c.Command.HideFlags("list-peers", "remove-peer", "address")

	if err := c.Command.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// Set up a client.
	client, err := c.Command.HTTPClient()
	if err != nil {
		return fmt.Errorf("error connecting to Consul agent: %s", err)
	}
	operator := client.Operator()

	// Dispatch based on the verb argument.
	if listPeers {
		result, err := raftListPeers(operator, c.Command.HTTPStale())
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error getting peers: %v", err))
		}
		c.Ui.Output(result)
	} else if removePeer {
		if err := raftRemovePeers(address, "", operator); err != nil {
			return fmt.Errorf("Error removing peer: %v", err)
		}
		c.Ui.Output(fmt.Sprintf("Removed peer with address %q", address))
	} else {
		c.Ui.Output(c.Help())
		return nil
	}

	return nil
}
