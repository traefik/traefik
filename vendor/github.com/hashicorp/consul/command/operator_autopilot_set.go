package command

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
)

type OperatorAutopilotSetCommand struct {
	base.Command
}

func (c *OperatorAutopilotSetCommand) Help() string {
	helpText := `
Usage: consul operator autopilot set-config [options]

Modifies the current Autopilot configuration.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *OperatorAutopilotSetCommand) Synopsis() string {
	return "Modify the current Autopilot configuration"
}

func (c *OperatorAutopilotSetCommand) Run(args []string) int {
	var cleanupDeadServers base.BoolValue
	var maxTrailingLogs base.UintValue
	var lastContactThreshold base.DurationValue
	var serverStabilizationTime base.DurationValue
	var redundancyZoneTag base.StringValue
	var disableUpgradeMigration base.BoolValue

	f := c.Command.NewFlagSet(c)

	f.Var(&cleanupDeadServers, "cleanup-dead-servers",
		"Controls whether Consul will automatically remove dead servers "+
			"when new ones are successfully added. Must be one of `true|false`.")
	f.Var(&maxTrailingLogs, "max-trailing-logs",
		"Controls the maximum number of log entries that a server can trail the "+
			"leader by before being considered unhealthy.")
	f.Var(&lastContactThreshold, "last-contact-threshold",
		"Controls the maximum amount of time a server can go without contact "+
			"from the leader before being considered unhealthy. Must be a duration value "+
			"such as `200ms`.")
	f.Var(&serverStabilizationTime, "server-stabilization-time",
		"Controls the minimum amount of time a server must be stable in the "+
			"'healthy' state before being added to the cluster. Only takes effect if all "+
			"servers are running Raft protocol version 3 or higher. Must be a duration "+
			"value such as `10s`.")
	f.Var(&redundancyZoneTag, "redundancy-zone-tag",
		"(Enterprise-only) Controls the node_meta tag name used for separating servers into "+
			"different redundancy zones.")
	f.Var(&disableUpgradeMigration, "disable-upgrade-migration",
		"(Enterprise-only) Controls whether Consul will avoid promoting new servers until "+
			"it can perform a migration. Must be one of `true|false`.")

	if err := c.Command.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		c.Ui.Error(fmt.Sprintf("Failed to parse args: %v", err))
		return 1
	}

	// Set up a client.
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing client: %s", err))
		return 1
	}

	// Fetch the current configuration.
	operator := client.Operator()
	conf, err := operator.AutopilotGetConfiguration(nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Autopilot configuration: %s", err))
		return 1
	}

	// Update the config values based on the set flags.
	cleanupDeadServers.Merge(&conf.CleanupDeadServers)
	redundancyZoneTag.Merge(&conf.RedundancyZoneTag)
	disableUpgradeMigration.Merge(&conf.DisableUpgradeMigration)

	trailing := uint(conf.MaxTrailingLogs)
	maxTrailingLogs.Merge(&trailing)
	conf.MaxTrailingLogs = uint64(trailing)

	last := time.Duration(*conf.LastContactThreshold)
	lastContactThreshold.Merge(&last)
	conf.LastContactThreshold = api.NewReadableDuration(last)

	stablization := time.Duration(*conf.ServerStabilizationTime)
	serverStabilizationTime.Merge(&stablization)
	conf.ServerStabilizationTime = api.NewReadableDuration(stablization)

	// Check-and-set the new configuration.
	result, err := operator.AutopilotCASConfiguration(conf, nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error setting Autopilot configuration: %s", err))
		return 1
	}
	if result {
		c.Ui.Output("Configuration updated!")
		return 0
	} else {
		c.Ui.Output("Configuration could not be atomically updated, please try again")
		return 1
	}
}
