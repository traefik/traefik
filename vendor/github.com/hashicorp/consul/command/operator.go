package command

import (
	"strings"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

// OperatorCommand is used to provide various low-level tools for Consul
// operators.
type OperatorCommand struct {
	base.Command
}

func (c *OperatorCommand) Help() string {
	helpText := `
Usage: consul operator <subcommand> [options]

  Provides cluster-level tools for Consul operators, such as interacting with
  the Raft subsystem. NOTE: Use this command with extreme caution, as improper
  use could lead to a Consul outage and even loss of data.

  If ACLs are enabled then a token with operator privileges may be required in
  order to use this command. Requests are forwarded internally to the leader
  if required, so this can be run from any Consul node in a cluster.

  Run consul operator <subcommand> with no arguments for help on that
  subcommand.
`
	return strings.TrimSpace(helpText)
}

func (c *OperatorCommand) Run(args []string) int {
	return cli.RunResultHelp
}

// Synopsis returns a one-line description of this command.
func (c *OperatorCommand) Synopsis() string {
	return "Provides cluster-level tools for Consul operators"
}
