package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/ryanuber/columnize"
)

type OperatorRaftListCommand struct {
	base.Command
}

func (c *OperatorRaftListCommand) Help() string {
	helpText := `
Usage: consul operator raft list-peers [options]

Displays the current Raft peer configuration.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorRaftListCommand) Synopsis() string {
	return "Display the current Raft peer configuration"
}

func (c *OperatorRaftListCommand) Run(args []string) int {
	c.Command.NewFlagSet(c)

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
	result, err := raftListPeers(client.Operator(), c.Command.HTTPStale())
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting peers: %v", err))
	}
	c.Ui.Output(result)

	return 0
}

func raftListPeers(operator *api.Operator, stale bool) (string, error) {
	q := &api.QueryOptions{
		AllowStale: stale,
	}
	reply, err := operator.RaftGetConfiguration(q)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve raft configuration: %v", err)
	}

	// Format it as a nice table.
	result := []string{"Node|ID|Address|State|Voter"}
	for _, s := range reply.Servers {
		state := "follower"
		if s.Leader {
			state = "leader"
		}
		result = append(result, fmt.Sprintf("%s|%s|%s|%s|%v",
			s.Node, s.ID, s.Address, state, s.Voter))
	}

	return columnize.SimpleFormat(result), nil
}
