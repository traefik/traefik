package command

import (
	"strings"

	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

// KVCommand is a Command implementation that just shows help for
// the subcommands nested below it.
type KVCommand struct {
	base.Command
}

func (c *KVCommand) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *KVCommand) Help() string {
	helpText := `
Usage: consul kv <subcommand> [options] [args]

  This command has subcommands for interacting with Consul's key-value
  store. Here are some simple examples, and more detailed examples are
  available in the subcommands or the documentation.

  Create or update the key named "redis/config/connections" with the value "5":

      $ consul kv put redis/config/connections 5

  Read this value back:

      $ consul kv get redis/config/connections

  Or get detailed key information:

      $ consul kv get -detailed redis/config/connections

  Finally, delete the key:

      $ consul kv delete redis/config/connections

  For more examples, ask for subcommand help or view the documentation.

`
	return strings.TrimSpace(helpText)
}

func (c *KVCommand) Synopsis() string {
	return "Interact with the key-value store"
}

var apiOptsText = strings.TrimSpace(`
API Options:

  -http-addr=<addr>       Address of the Consul agent with the port. This can
                          be an IP address or DNS address, but it must include
                          the port. This can also be specified via the
                          CONSUL_HTTP_ADDR environment variable. The default
                          value is 127.0.0.1:8500.

  -datacenter=<name>      Name of the datacenter to query. If unspecified, the
                          query will default to the datacenter of the Consul
                          agent at the HTTP address.

  -token=<value>          ACL token to use in the request. This can also be
                          specified via the CONSUL_HTTP_TOKEN environment
                          variable. If unspecified, the query will default to
                          the token of the Consul agent at the HTTP address.

  -stale                  Permit any Consul server (non-leader) to respond to
                          this request. This allows for lower latency and higher
                          throughput, but can result in stale data. This option
                          has no effect on non-read operations. The default
                          value is false.
`)
