package command

import (
	"fmt"
	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/consul"
	"github.com/mitchellh/cli"
)

// VersionCommand is a Command implementation prints the version.
type VersionCommand struct {
	HumanVersion string
	Ui           cli.Ui
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Run(_ []string) int {
	c.Ui.Output(fmt.Sprintf("Consul %s", c.HumanVersion))

	config := agent.DefaultConfig()
	var supplement string
	if config.Protocol < consul.ProtocolVersionMax {
		supplement = fmt.Sprintf(" (agent will automatically use protocol >%d when speaking to compatible agents)",
			config.Protocol)
	}
	c.Ui.Output(fmt.Sprintf("Protocol %d spoken by default, understands %d to %d%s",
		config.Protocol, consul.ProtocolVersionMin, consul.ProtocolVersionMax, supplement))

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints the Consul version"
}
