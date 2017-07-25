package command

import (
	"strings"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

type OperatorAutopilotCommand struct {
	base.Command
}

func (c *OperatorAutopilotCommand) Help() string {
	helpText := `
Usage: consul operator autopilot <subcommand> [options]

The Autopilot operator command is used to interact with Consul's Autopilot
subsystem. The command can be used to view or modify the current configuration.

`

	return strings.TrimSpace(helpText)
}

func (c *OperatorAutopilotCommand) Synopsis() string {
	return "Provides tools for modifying Autopilot configuration"
}

func (c *OperatorAutopilotCommand) Run(args []string) int {
	return cli.RunResultHelp
}
