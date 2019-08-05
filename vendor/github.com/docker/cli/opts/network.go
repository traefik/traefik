package opts

import (
	"encoding/csv"
	"fmt"
	"regexp"
	"strings"
)

const (
	networkOptName  = "name"
	networkOptAlias = "alias"
	driverOpt       = "driver-opt"
)

// NetworkAttachmentOpts represents the network options for endpoint creation
type NetworkAttachmentOpts struct {
	Target       string
	Aliases      []string
	DriverOpts   map[string]string
	Links        []string // TODO add support for links in the csv notation of `--network`
	IPv4Address  string   // TODO add support for IPv4-address in the csv notation of `--network`
	IPv6Address  string   // TODO add support for IPv6-address in the csv notation of `--network`
	LinkLocalIPs []string // TODO add support for LinkLocalIPs in the csv notation of `--network` ?
}

// NetworkOpt represents a network config in swarm mode.
type NetworkOpt struct {
	options []NetworkAttachmentOpts
}

// Set networkopts value
func (n *NetworkOpt) Set(value string) error {
	longSyntax, err := regexp.MatchString(`\w+=\w+(,\w+=\w+)*`, value)
	if err != nil {
		return err
	}

	var netOpt NetworkAttachmentOpts
	if longSyntax {
		csvReader := csv.NewReader(strings.NewReader(value))
		fields, err := csvReader.Read()
		if err != nil {
			return err
		}

		netOpt.Aliases = []string{}
		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)

			if len(parts) < 2 {
				return fmt.Errorf("invalid field %s", field)
			}

			key := strings.TrimSpace(strings.ToLower(parts[0]))
			value := strings.TrimSpace(strings.ToLower(parts[1]))

			switch key {
			case networkOptName:
				netOpt.Target = value
			case networkOptAlias:
				netOpt.Aliases = append(netOpt.Aliases, value)
			case driverOpt:
				key, value, err = parseDriverOpt(value)
				if err == nil {
					if netOpt.DriverOpts == nil {
						netOpt.DriverOpts = make(map[string]string)
					}
					netOpt.DriverOpts[key] = value
				} else {
					return err
				}
			default:
				return fmt.Errorf("invalid field key %s", key)
			}
		}
		if len(netOpt.Target) == 0 {
			return fmt.Errorf("network name/id is not specified")
		}
	} else {
		netOpt.Target = value
	}
	n.options = append(n.options, netOpt)
	return nil
}

// Type returns the type of this option
func (n *NetworkOpt) Type() string {
	return "network"
}

// Value returns the networkopts
func (n *NetworkOpt) Value() []NetworkAttachmentOpts {
	return n.options
}

// String returns the network opts as a string
func (n *NetworkOpt) String() string {
	return ""
}

// NetworkMode return the network mode for the network option
func (n *NetworkOpt) NetworkMode() string {
	networkIDOrName := "default"
	netOptVal := n.Value()
	if len(netOptVal) > 0 {
		networkIDOrName = netOptVal[0].Target
	}
	return networkIDOrName
}

func parseDriverOpt(driverOpt string) (string, string, error) {
	parts := strings.SplitN(driverOpt, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid key value pair format in driver options")
	}
	key := strings.TrimSpace(strings.ToLower(parts[0]))
	value := strings.TrimSpace(strings.ToLower(parts[1]))
	return key, value, nil
}
