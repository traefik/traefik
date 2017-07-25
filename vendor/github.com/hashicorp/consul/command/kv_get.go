package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
)

// KVGetCommand is a Command implementation that is used to fetch the value of
// a key from the key-value store.
type KVGetCommand struct {
	base.Command
}

func (c *KVGetCommand) Help() string {
	helpText := `
Usage: consul kv get [options] [KEY_OR_PREFIX]

  Retrieves the value from Consul's key-value store at the given key name. If no
  key exists with that name, an error is returned. If a key exists with that
  name but has no data, nothing is returned. If the name or prefix is omitted,
  it defaults to "" which is the root of the key-value store.

  To retrieve the value for the key named "foo" in the key-value store:

      $ consul kv get foo

  This will return the original, raw value stored in Consul. To view detailed
  information about the key, specify the "-detailed" flag. This will output all
  known metadata about the key including ModifyIndex and any user-supplied
  flags:

      $ consul kv get -detailed foo

  To treat the path as a prefix and list all keys which start with the given
  prefix, specify the "-recurse" flag:

      $ consul kv get -recurse foo

  This will return all key-vlaue pairs. To just list the keys which start with
  the specified prefix, use the "-keys" option instead:

      $ consul kv get -keys foo

  For a full list of options and examples, please see the Consul documentation.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *KVGetCommand) Run(args []string) int {
	f := c.Command.NewFlagSet(c)
	base64encode := f.Bool("base64", false,
		"Base64 encode the value. The default value is false.")
	detailed := f.Bool("detailed", false,
		"Provide additional metadata about the key in addition to the value such "+
			"as the ModifyIndex and any flags that may have been set on the key. "+
			"The default value is false.")
	keys := f.Bool("keys", false,
		"List keys which start with the given prefix, but not their values. "+
			"This is especially useful if you only need the key names themselves. "+
			"This option is commonly combined with the -separator option. The default "+
			"value is false.")
	recurse := f.Bool("recurse", false,
		"Recursively look at all keys prefixed with the given path. The default "+
			"value is false.")
	separator := f.String("separator", "/",
		"String to use as a separator between keys. The default value is \"/\", "+
			"but this option is only taken into account when paired with the -keys flag.")

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

	// If the key is empty and we are not doing a recursive or key-based lookup,
	// this is an error.
	if key == "" && !(*recurse || *keys) {
		c.Ui.Error("Error! Missing KEY argument")
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	switch {
	case *keys:
		keys, _, err := client.KV().Keys(key, *separator, &api.QueryOptions{
			AllowStale: c.Command.HTTPStale(),
		})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
			return 1
		}

		for _, k := range keys {
			c.Ui.Info(string(k))
		}

		return 0
	case *recurse:
		pairs, _, err := client.KV().List(key, &api.QueryOptions{
			AllowStale: c.Command.HTTPStale(),
		})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
			return 1
		}

		for i, pair := range pairs {
			if *detailed {
				var b bytes.Buffer
				if err := prettyKVPair(&b, pair, *base64encode); err != nil {
					c.Ui.Error(fmt.Sprintf("Error rendering KV pair: %s", err))
					return 1
				}

				c.Ui.Info(b.String())

				if i < len(pairs)-1 {
					c.Ui.Info("")
				}
			} else {
				if *base64encode {
					c.Ui.Info(fmt.Sprintf("%s:%s", pair.Key, base64.StdEncoding.EncodeToString(pair.Value)))
				} else {
					c.Ui.Info(fmt.Sprintf("%s:%s", pair.Key, pair.Value))
				}
			}
		}

		return 0
	default:
		pair, _, err := client.KV().Get(key, &api.QueryOptions{
			AllowStale: c.Command.HTTPStale(),
		})
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
			return 1
		}

		if pair == nil {
			c.Ui.Error(fmt.Sprintf("Error! No key exists at: %s", key))
			return 1
		}

		if *detailed {
			var b bytes.Buffer
			if err := prettyKVPair(&b, pair, *base64encode); err != nil {
				c.Ui.Error(fmt.Sprintf("Error rendering KV pair: %s", err))
				return 1
			}

			c.Ui.Info(b.String())
			return 0
		} else {
			c.Ui.Info(string(pair.Value))
			return 0
		}
	}
}

func (c *KVGetCommand) Synopsis() string {
	return "Retrieves or lists data from the KV store"
}

func prettyKVPair(w io.Writer, pair *api.KVPair, base64EncodeValue bool) error {
	tw := tabwriter.NewWriter(w, 0, 2, 6, ' ', 0)
	fmt.Fprintf(tw, "CreateIndex\t%d\n", pair.CreateIndex)
	fmt.Fprintf(tw, "Flags\t%d\n", pair.Flags)
	fmt.Fprintf(tw, "Key\t%s\n", pair.Key)
	fmt.Fprintf(tw, "LockIndex\t%d\n", pair.LockIndex)
	fmt.Fprintf(tw, "ModifyIndex\t%d\n", pair.ModifyIndex)
	if pair.Session == "" {
		fmt.Fprint(tw, "Session\t-\n")
	} else {
		fmt.Fprintf(tw, "Session\t%s\n", pair.Session)
	}
	if base64EncodeValue {
		fmt.Fprintf(tw, "Value\t%s", base64.StdEncoding.EncodeToString(pair.Value))
	} else {
		fmt.Fprintf(tw, "Value\t%s", pair.Value)
	}
	return tw.Flush()
}
