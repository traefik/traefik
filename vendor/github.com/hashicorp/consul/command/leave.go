package command

import (
	"fmt"
	"github.com/hashicorp/consul/command/base"
	"strings"
)

// LeaveCommand is a Command implementation that instructs
// the Consul agent to gracefully leave the cluster
type LeaveCommand struct {
	base.Command
}

func (c *LeaveCommand) Help() string {
	helpText := `
Usage: consul leave [options]

  Causes the agent to gracefully leave the Consul cluster and shutdown.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *LeaveCommand) Run(args []string) int {
	f := c.Command.NewFlagSet(c)
	if err := c.Command.Parse(args); err != nil {
		return 1
	}
	nonFlagArgs := f.Args()
	if len(nonFlagArgs) > 0 {
		c.Ui.Error(fmt.Sprintf("Error found unexpected args: %v", nonFlagArgs))
		c.Ui.Output(c.Help())
		return 1
	}

	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	if err := client.Agent().Leave(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error leaving: %s", err))
		return 1
	}

	c.Ui.Output("Graceful leave complete")
	return 0
}

func (c *LeaveCommand) Synopsis() string {
	return "Gracefully leaves the Consul cluster and shuts down"
}
