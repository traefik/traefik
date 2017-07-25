package command

import (
	"fmt"
	"github.com/hashicorp/consul/command/base"
	"strings"
)

// ReloadCommand is a Command implementation that instructs
// the Consul agent to reload configurations
type ReloadCommand struct {
	base.Command
}

func (c *ReloadCommand) Help() string {
	helpText := `
Usage: consul reload

  Causes the agent to reload configurations. This can be used instead
  of sending the SIGHUP signal to the agent.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *ReloadCommand) Run(args []string) int {
	c.Command.NewFlagSet(c)

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	if err := client.Agent().Reload(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error reloading: %s", err))
		return 1
	}

	c.Ui.Output("Configuration reload triggered")
	return 0
}

func (c *ReloadCommand) Synopsis() string {
	return "Triggers the agent to reload configuration files"
}
