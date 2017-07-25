package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/command/base"
)

// MaintCommand is a Command implementation that enables or disables
// node or service maintenance mode.
type MaintCommand struct {
	base.Command
}

func (c *MaintCommand) Help() string {
	helpText := `
Usage: consul maint [options]

  Places a node or service into maintenance mode. During maintenance mode,
  the node or service will be excluded from all queries through the DNS
  or API interfaces, effectively taking it out of the pool of available
  nodes. This is done by registering an additional critical health check.

  When enabling maintenance mode for a node or service, you may optionally
  specify a reason string. This string will appear in the "Notes" field
  of the critical health check which is registered against the node or
  service. If no reason is provided, a default value will be used.

  Maintenance mode is persistent, and will be restored in the event of an
  agent restart. It is therefore required to disable maintenance mode on
  a given node or service before it will be placed back into the pool.

  By default, we operate on the node as a whole. By specifying the
  "-service" argument, this behavior can be changed to enable or disable
  only a specific service.

  If no arguments are given, the agent's maintenance status will be shown.
  This will return blank if nothing is currently under maintenance.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *MaintCommand) Run(args []string) int {
	var enable bool
	var disable bool
	var reason string
	var serviceID string

	f := c.Command.NewFlagSet(c)

	f.BoolVar(&enable, "enable", false, "Enable maintenance mode.")
	f.BoolVar(&disable, "disable", false, "Disable maintenance mode.")
	f.StringVar(&reason, "reason", "", "Text describing the maintenance reason.")
	f.StringVar(&serviceID, "service", "", "Control maintenance mode for a specific service ID.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	// Ensure we don't have conflicting args
	if enable && disable {
		c.Ui.Error("Only one of -enable or -disable may be provided")
		return 1
	}
	if !enable && reason != "" {
		c.Ui.Error("Reason may only be provided with -enable")
		return 1
	}
	if !enable && !disable && serviceID != "" {
		c.Ui.Error("Service requires either -enable or -disable")
		return 1
	}

	// Create and test the HTTP client
	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}
	a := client.Agent()
	nodeName, err := a.NodeName()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error querying Consul agent: %s", err))
		return 1
	}

	if !enable && !disable {
		// List mode - list nodes/services in maintenance mode
		checks, err := a.Checks()
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error getting checks: %s", err))
			return 1
		}

		for _, check := range checks {
			if check.CheckID == "_node_maintenance" {
				c.Ui.Output("Node:")
				c.Ui.Output("  Name:   " + nodeName)
				c.Ui.Output("  Reason: " + check.Notes)
				c.Ui.Output("")
			} else if strings.HasPrefix(string(check.CheckID), "_service_maintenance:") {
				c.Ui.Output("Service:")
				c.Ui.Output("  ID:     " + check.ServiceID)
				c.Ui.Output("  Reason: " + check.Notes)
				c.Ui.Output("")
			}
		}

		return 0
	}

	if enable {
		// Enable node maintenance
		if serviceID == "" {
			if err := a.EnableNodeMaintenance(reason); err != nil {
				c.Ui.Error(fmt.Sprintf("Error enabling node maintenance: %s", err))
				return 1
			}
			c.Ui.Output("Node maintenance is now enabled")
			return 0
		}

		// Enable service maintenance
		if err := a.EnableServiceMaintenance(serviceID, reason); err != nil {
			c.Ui.Error(fmt.Sprintf("Error enabling service maintenance: %s", err))
			return 1
		}
		c.Ui.Output(fmt.Sprintf("Service maintenance is now enabled for %q", serviceID))
		return 0
	}

	if disable {
		// Disable node maintenance
		if serviceID == "" {
			if err := a.DisableNodeMaintenance(); err != nil {
				c.Ui.Error(fmt.Sprintf("Error disabling node maintenance: %s", err))
				return 1
			}
			c.Ui.Output("Node maintenance is now disabled")
			return 0
		}

		// Disable service maintenance
		if err := a.DisableServiceMaintenance(serviceID); err != nil {
			c.Ui.Error(fmt.Sprintf("Error disabling service maintenance: %s", err))
			return 1
		}
		c.Ui.Output(fmt.Sprintf("Service maintenance is now disabled for %q", serviceID))
		return 0
	}

	return 0
}

func (c *MaintCommand) Synopsis() string {
	return "Controls node or service maintenance mode"
}
