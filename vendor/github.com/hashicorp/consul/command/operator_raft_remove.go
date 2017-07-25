package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
)

type OperatorRaftRemoveCommand struct {
	base.Command
}

func (c *OperatorRaftRemoveCommand) Help() string {
	helpText := `
Usage: consul operator raft remove-peer [options]

Remove the Consul server with given -address from the Raft configuration.

There are rare cases where a peer may be left behind in the Raft quorum even
though the server is no longer present and known to the cluster. This command
can be used to remove the failed server so that it is no longer affects the Raft
quorum. If the server still shows in the output of the "consul members" command,
it is preferable to clean up by simply running "consul force-leave" instead of
this command.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorRaftRemoveCommand) Synopsis() string {
	return "Remove a Consul server from the Raft configuration"
}

func (c *OperatorRaftRemoveCommand) Run(args []string) int {
	f := c.Command.NewFlagSet(c)

	var address, id string
	f.StringVar(&address, "address", "",
		"The address to remove from the Raft configuration.")
	f.StringVar(&id, "id", "",
		"The ID to remove from the Raft configuration.")

	if err := c.Command.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		c.Ui.Error(fmt.Sprintf("Failed to parse args: %v", err))
		return 1
	}

	// Set up a client.
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Fetch the current configuration.
	if err := raftRemovePeers(address, id, client.Operator()); err != nil {
		c.Ui.Error(fmt.Sprintf("Error removing peer: %v", err))
		return 1
	}
	if address != "" {
		c.Ui.Output(fmt.Sprintf("Removed peer with address %q", address))
	} else {
		c.Ui.Output(fmt.Sprintf("Removed peer with id %q", id))
	}

	return 0
}

func raftRemovePeers(address, id string, operator *api.Operator) error {
	if len(address) == 0 && len(id) == 0 {
		return fmt.Errorf("an address or id is required for the peer to remove")
	}
	if len(address) > 0 && len(id) > 0 {
		return fmt.Errorf("cannot give both an address and id")
	}

	// Try to kick the peer.
	if len(address) > 0 {
		if err := operator.RaftRemovePeerByAddress(address, nil); err != nil {
			return err
		}
	} else {
		if err := operator.RaftRemovePeerByID(id, nil); err != nil {
			return err
		}
	}

	return nil
}
