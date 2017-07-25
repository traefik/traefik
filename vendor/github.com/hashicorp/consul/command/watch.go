package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/consul/watch"
)

// WatchCommand is a Command implementation that is used to setup
// a "watch" which uses a sub-process
type WatchCommand struct {
	base.Command
	ShutdownCh <-chan struct{}
}

func (c *WatchCommand) Help() string {
	helpText := `
Usage: consul watch [options] [child...]

  Watches for changes in a given data view from Consul. If a child process
  is specified, it will be invoked with the latest results on changes. Otherwise,
  the latest values are dumped to stdout and the watch terminates.

  Providing the watch type is required, and other parameters may be required
  or supported depending on the watch type.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *WatchCommand) Run(args []string) int {
	var watchType, key, prefix, service, tag, passingOnly, state, name string

	f := c.Command.NewFlagSet(c)
	f.StringVar(&watchType, "type", "",
		"Specifies the watch type. One of key, keyprefix services, nodes, "+
			"service, checks, or event.")
	f.StringVar(&key, "key", "",
		"Specifies the key to watch. Only for 'key' type.")
	f.StringVar(&prefix, "prefix", "",
		"Specifies the key prefix to watch. Only for 'keyprefix' type.")
	f.StringVar(&service, "service", "",
		"Specifies the service to watch. Required for 'service' type, "+
			"optional for 'checks' type.")
	f.StringVar(&tag, "tag", "",
		"Specifies the service tag to filter on. Optional for 'service' type.")
	f.StringVar(&passingOnly, "passingonly", "",
		"Specifies if only hosts passing all checks are displayed. "+
			"Optional for 'service' type, must be one of `[true|false]`. Defaults false.")
	f.StringVar(&state, "state", "",
		"Specifies the states to watch. Optional for 'checks' type.")
	f.StringVar(&name, "name", "",
		"Specifies an event name to watch. Only for 'event' type.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	// Check for a type
	if watchType == "" {
		c.Ui.Error("Watch type must be specified")
		c.Ui.Error("")
		c.Ui.Error(c.Help())
		return 1
	}

	// Grab the script to execute if any
	script := strings.Join(f.Args(), " ")

	// Compile the watch parameters
	params := make(map[string]interface{})
	if watchType != "" {
		params["type"] = watchType
	}
	if c.Command.HTTPDatacenter() != "" {
		params["datacenter"] = c.Command.HTTPDatacenter()
	}
	if c.Command.HTTPToken() != "" {
		params["token"] = c.Command.HTTPToken()
	}
	if key != "" {
		params["key"] = key
	}
	if prefix != "" {
		params["prefix"] = prefix
	}
	if service != "" {
		params["service"] = service
	}
	if tag != "" {
		params["tag"] = tag
	}
	if c.Command.HTTPStale() {
		params["stale"] = c.Command.HTTPStale()
	}
	if state != "" {
		params["state"] = state
	}
	if name != "" {
		params["name"] = name
	}
	if passingOnly != "" {
		b, err := strconv.ParseBool(passingOnly)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Failed to parse passingonly flag: %s", err))
			return 1
		}
		params["passingonly"] = b
	}

	// Create the watch
	wp, err := watch.Parse(params)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%s", err))
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}
	_, err = client.Agent().NodeName()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}

	// Setup handler

	// errExit:
	//	0: false
	//	1: true
	errExit := 0
	if script == "" {
		wp.Handler = func(idx uint64, data interface{}) {
			defer wp.Stop()
			buf, err := json.MarshalIndent(data, "", "    ")
			if err != nil {
				c.Ui.Error(fmt.Sprintf("Error encoding output: %s", err))
				errExit = 1
			}
			c.Ui.Output(string(buf))
		}
	} else {
		wp.Handler = func(idx uint64, data interface{}) {
			// Create the command
			var buf bytes.Buffer
			var err error
			cmd, err := agent.ExecScript(script)
			if err != nil {
				c.Ui.Error(fmt.Sprintf("Error executing handler: %s", err))
				goto ERR
			}
			cmd.Env = append(os.Environ(),
				"CONSUL_INDEX="+strconv.FormatUint(idx, 10),
			)

			// Encode the input
			if err = json.NewEncoder(&buf).Encode(data); err != nil {
				c.Ui.Error(fmt.Sprintf("Error encoding output: %s", err))
				goto ERR
			}
			cmd.Stdin = &buf
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Run the handler
			if err := cmd.Run(); err != nil {
				c.Ui.Error(fmt.Sprintf("Error executing handler: %s", err))
				goto ERR
			}
			return
		ERR:
			wp.Stop()
			errExit = 1
		}
	}

	// Watch for a shutdown
	go func() {
		<-c.ShutdownCh
		wp.Stop()
		os.Exit(0)
	}()

	// Run the watch
	if err := wp.Run(c.Command.HTTPAddr()); err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}

	return errExit
}

func (c *WatchCommand) Synopsis() string {
	return "Watch for changes in Consul"
}
