package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
)

// KVPutCommand is a Command implementation that is used to write data to the
// key-value store.
type KVPutCommand struct {
	base.Command

	// testStdin is the input for testing.
	testStdin io.Reader
}

func (c *KVPutCommand) Help() string {
	helpText := `
Usage: consul kv put [options] KEY [DATA]

  Writes the data to the given path in the key-value store. The data can be of
  any type.

      $ consul kv put config/redis/maxconns 5

  The data can also be consumed from a file on disk by prefixing with the "@"
  symbol. For example:

      $ consul kv put config/program/license @license.lic

  Or it can be read from stdin using the "-" symbol:

      $ echo "abcd1234" | consul kv put config/program/license -

  The DATA argument itself is optional. If omitted, this will create an empty
  key-value pair at the specified path:

      $ consul kv put webapp/beta/active

  If the -base64 flag is specified, the data will be treated as base 64
  encoded.

  To perform a Check-And-Set operation, specify the -cas flag with the
  appropriate -modify-index flag corresponding to the key you want to perform
  the CAS operation on:

      $ consul kv put -cas -modify-index=844 config/redis/maxconns 5

  Additional flags and more advanced use cases are detailed below.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *KVPutCommand) Run(args []string) int {
	f := c.Command.NewFlagSet(c)
	cas := f.Bool("cas", false,
		"Perform a Check-And-Set operation. Specifying this value also "+
			"requires the -modify-index flag to be set. The default value "+
			"is false.")
	flags := f.Uint64("flags", 0,
		"Unsigned integer value to assign to this key-value pair. This "+
			"value is not read by Consul, so clients can use this value however "+
			"makes sense for their use case. The default value is 0 (no flags).")
	base64encoded := f.Bool("base64", false,
		"Treat the data as base 64 encoded. The default value is false.")
	modifyIndex := f.Uint64("modify-index", 0,
		"Unsigned integer representing the ModifyIndex of the key. This is "+
			"used in combination with the -cas flag.")
	session := f.String("session", "",
		"User-defined identifer for this session as a string. This is commonly "+
			"used with the -acquire and -release operations to build robust locking, "+
			"but it can be set on any key. The default value is empty (no session).")
	acquire := f.Bool("acquire", false,
		"Obtain a lock on the key. If the key does not exist, this operation "+
			"will create the key and obtain the lock. The session must already "+
			"exist and be specified via the -session flag. The default value is false.")
	release := f.Bool("release", false,
		"Forfeit the lock on the key at the given path. This requires the "+
			"-session flag to be set. The key must be held by the session in order to "+
			"be unlocked. The default value is false.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	// Check for arg validation
	args = f.Args()
	key, data, err := c.dataFromArgs(args)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error! %s", err))
		return 1
	}

	dataBytes := []byte(data)
	if *base64encoded {
		dataBytes, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error! Cannot base 64 decode data: %s", err))
		}
	}

	// Session is reauired for release or acquire
	if (*release || *acquire) && *session == "" {
		c.Ui.Error("Error! Missing -session (required with -acquire and -release)")
		return 1
	}

	// ModifyIndex is required for CAS
	if *cas && *modifyIndex == 0 {
		c.Ui.Error("Must specify -modify-index with -cas!")
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	pair := &api.KVPair{
		Key:         key,
		ModifyIndex: *modifyIndex,
		Flags:       *flags,
		Value:       dataBytes,
		Session:     *session,
	}

	switch {
	case *cas:
		ok, _, err := client.KV().CAS(pair, nil)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error! Did not write to %s: %s", key, err))
			return 1
		}
		if !ok {
			c.Ui.Error(fmt.Sprintf("Error! Did not write to %s: CAS failed", key))
			return 1
		}

		c.Ui.Info(fmt.Sprintf("Success! Data written to: %s", key))
		return 0
	case *acquire:
		ok, _, err := client.KV().Acquire(pair, nil)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error! Failed writing data: %s", err))
			return 1
		}
		if !ok {
			c.Ui.Error("Error! Did not acquire lock")
			return 1
		}

		c.Ui.Info(fmt.Sprintf("Success! Lock acquired on: %s", key))
		return 0
	case *release:
		ok, _, err := client.KV().Release(pair, nil)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error! Failed writing data: %s", key))
			return 1
		}
		if !ok {
			c.Ui.Error("Error! Did not release lock")
			return 1
		}

		c.Ui.Info(fmt.Sprintf("Success! Lock released on: %s", key))
		return 0
	default:
		if _, err := client.KV().Put(pair, nil); err != nil {
			c.Ui.Error(fmt.Sprintf("Error! Failed writing data: %s", err))
			return 1
		}

		c.Ui.Info(fmt.Sprintf("Success! Data written to: %s", key))
		return 0
	}
}

func (c *KVPutCommand) Synopsis() string {
	return "Sets or updates data in the KV store"
}

func (c *KVPutCommand) dataFromArgs(args []string) (string, string, error) {
	var stdin io.Reader = os.Stdin
	if c.testStdin != nil {
		stdin = c.testStdin
	}

	switch len(args) {
	case 0:
		return "", "", fmt.Errorf("Missing KEY argument")
	case 1:
		return args[0], "", nil
	case 2:
	default:
		return "", "", fmt.Errorf("Too many arguments (expected 1 or 2, got %d)", len(args))
	}

	key := args[0]
	data := args[1]

	// Handle empty quoted shell parameters
	if len(data) == 0 {
		return key, "", nil
	}

	switch data[0] {
	case '@':
		data, err := ioutil.ReadFile(data[1:])
		if err != nil {
			return "", "", fmt.Errorf("Failed to read file: %s", err)
		}
		return key, string(data), nil
	case '-':
		if len(data) > 1 {
			return key, data, nil
		} else {
			var b bytes.Buffer
			if _, err := io.Copy(&b, stdin); err != nil {
				return "", "", fmt.Errorf("Failed to read stdin: %s", err)
			}
			return key, b.String(), nil
		}
	default:
		return key, data, nil
	}
}
