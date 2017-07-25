package command

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
)

// KVExportCommand is a Command implementation that is used to export
// a KV tree as JSON
type KVExportCommand struct {
	base.Command
}

func (c *KVExportCommand) Synopsis() string {
	return "Exports a tree from the KV store as JSON"
}

func (c *KVExportCommand) Help() string {
	helpText := `
Usage: consul kv export [KEY_OR_PREFIX]

  Retrieves key-value pairs for the given prefix from Consul's key-value store,
  and writes a JSON representation to stdout. This can be used with the command
  "consul kv import" to move entire trees between Consul clusters.

      $ consul kv export vault

  For a full list of options and examples, please see the Consul documentation.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *KVExportCommand) Run(args []string) int {
	f := c.Command.NewFlagSet(c)
	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	key := ""
	// Check for arg validation
	args = f.Args()
	switch len(args) {
	case 0:
		key = ""
	case 1:
		key = args[0]
	default:
		c.Ui.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}

	// This is just a "nice" thing to do. Since pairs cannot start with a /, but
	// users will likely put "/" or "/foo", lets go ahead and strip that for them
	// here.
	if len(key) > 0 && key[0] == '/' {
		key = key[1:]
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	pairs, _, err := client.KV().List(key, &api.QueryOptions{
		AllowStale: c.Command.HTTPStale(),
	})
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}

	exported := make([]*kvExportEntry, len(pairs))
	for i, pair := range pairs {
		exported[i] = toExportEntry(pair)
	}

	marshaled, err := json.MarshalIndent(exported, "", "\t")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error exporting KV data: %s", err))
		return 1
	}

	c.Ui.Info(string(marshaled))

	return 0
}

type kvExportEntry struct {
	Key   string `json:"key"`
	Flags uint64 `json:"flags"`
	Value string `json:"value"`
}

func toExportEntry(pair *api.KVPair) *kvExportEntry {
	return &kvExportEntry{
		Key:   pair.Key,
		Flags: pair.Flags,
		Value: base64.StdEncoding.EncodeToString(pair.Value),
	}
}
