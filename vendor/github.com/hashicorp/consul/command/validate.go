package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/command/agent"
	"github.com/hashicorp/consul/command/base"
)

// ValidateCommand is a Command implementation that is used to
// verify config files
type ValidateCommand struct {
	base.Command
}

func (c *ValidateCommand) Help() string {
	helpText := `
Usage: consul validate [options] FILE_OR_DIRECTORY...

  Performs a basic sanity test on Consul configuration files. For each file
  or directory given, the validate command will attempt to parse the
  contents just as the "consul agent" command would, and catch any errors.
  This is useful to do a test of the configuration only, without actually
  starting the agent.

  Returns 0 if the configuration is valid, or 1 if there are problems.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *ValidateCommand) Run(args []string) int {
	var configFiles []string
	var quiet bool

	f := c.Command.NewFlagSet(c)
	f.Var((*agent.AppendSliceValue)(&configFiles), "config-file",
		"Path to a JSON file to read configuration from. This can be specified multiple times.")
	f.Var((*agent.AppendSliceValue)(&configFiles), "config-dir",
		"Path to a directory to read configuration files from. This will read every file ending in "+
			".json as configuration in this directory in alphabetical order.")
	f.BoolVar(&quiet, "quiet", false,
		"When given, a successful run will produce no output.")
	c.Command.HideFlags("config-file", "config-dir")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	if len(f.Args()) > 0 {
		configFiles = append(configFiles, f.Args()...)
	}

	if len(configFiles) < 1 {
		c.Ui.Error("Must specify at least one config file or directory")
		return 1
	}

	_, err := agent.ReadConfigPaths(configFiles)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Config validation failed: %v", err.Error()))
		return 1
	}

	if !quiet {
		c.Ui.Output("Configuration is valid!")
	}
	return 0
}

func (c *ValidateCommand) Synopsis() string {
	return "Validate config files/directories"
}
