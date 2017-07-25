package command

import (
	"fmt"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
	"github.com/mitchellh/cli"
)

// KeyringCommand is a Command implementation that handles querying, installing,
// and removing gossip encryption keys from a keyring.
type KeyringCommand struct {
	base.Command
}

func (c *KeyringCommand) Run(args []string) int {
	var installKey, useKey, removeKey string
	var listKeys bool
	var relay int

	f := c.Command.NewFlagSet(c)

	f.StringVar(&installKey, "install", "",
		"Install a new encryption key. This will broadcast the new key to "+
			"all members in the cluster.")
	f.StringVar(&useKey, "use", "",
		"Change the primary encryption key, which is used to encrypt "+
			"messages. The key must already be installed before this operation "+
			"can succeed.")
	f.StringVar(&removeKey, "remove", "",
		"Remove the given key from the cluster. This operation may only be "+
			"performed on keys which are not currently the primary key.")
	f.BoolVar(&listKeys, "list", false,
		"List all keys currently in use within the cluster.")
	f.IntVar(&relay, "relay-factor", 0,
		"Setting this to a non-zero value will cause nodes to relay their response "+
			"to the operation through this many randomly-chosen other nodes in the "+
			"cluster. The maximum allowed value is 5.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "",
		InfoPrefix:   "==> ",
		ErrorPrefix:  "",
		Ui:           c.Ui,
	}

	// Only accept a single argument
	found := listKeys
	for _, arg := range []string{installKey, useKey, removeKey} {
		if found && len(arg) > 0 {
			c.Ui.Error("Only a single action is allowed")
			return 1
		}
		found = found || len(arg) > 0
	}

	// Fail fast if no actionable args were passed
	if !found {
		c.Ui.Error(c.Help())
		return 1
	}

	// Validate the relay factor
	relayFactor, err := agent.ParseRelayFactor(relay)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing relay factor: %s", err))
		return 1
	}

	// All other operations will require a client connection
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	if listKeys {
		c.Ui.Info("Gathering installed encryption keys...")
		responses, err := client.Operator().KeyringList(&consulapi.QueryOptions{RelayFactor: relayFactor})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		c.handleList(responses)
		return 0
	}

	opts := &consulapi.WriteOptions{RelayFactor: relayFactor}
	if installKey != "" {
		c.Ui.Info("Installing new gossip encryption key...")
		err := client.Operator().KeyringInstall(installKey, opts)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	if useKey != "" {
		c.Ui.Info("Changing primary gossip encryption key...")
		err := client.Operator().KeyringUse(useKey, opts)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	if removeKey != "" {
		c.Ui.Info("Removing gossip encryption key...")
		err := client.Operator().KeyringRemove(removeKey, opts)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("error: %s", err))
			return 1
		}
		return 0
	}

	// Should never make it here
	return 0
}

func (c *KeyringCommand) handleList(responses []*consulapi.KeyringResponse) {
	for _, response := range responses {
		pool := response.Datacenter + " (LAN)"
		if response.WAN {
			pool = "WAN"
		}

		c.Ui.Output("")
		c.Ui.Output(pool + ":")
		for key, num := range response.Keys {
			c.Ui.Output(fmt.Sprintf("  %s [%d/%d]", key, num, response.NumNodes))
		}
	}
}

func (c *KeyringCommand) Help() string {
	helpText := `
Usage: consul keyring [options]

  Manages encryption keys used for gossip messages. Gossip encryption is
  optional. When enabled, this command may be used to examine active encryption
  keys in the cluster, add new keys, and remove old ones. When combined, this
  functionality provides the ability to perform key rotation cluster-wide,
  without disrupting the cluster.

  All operations performed by this command can only be run against server nodes,
  and affect both the LAN and WAN keyrings in lock-step.

  All variations of the keyring command return 0 if all nodes reply and there
  are no errors. If any node fails to reply or reports failure, the exit code
  will be 1.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *KeyringCommand) Synopsis() string {
	return "Manages gossip layer encryption keys"
}
