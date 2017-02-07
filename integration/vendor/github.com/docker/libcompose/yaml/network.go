package yaml

import (
	"fmt"
)

// Networks represents a list of service networks in compose file.
// It has several representation, hence this specific struct.
type Networks struct {
	Networks []*Network
}

// Network represents a  service network in compose file.
type Network struct {
	Name        string   `yaml:"-"`
	RealName    string   `yaml:"-"`
	Aliases     []string `yaml:"aliases,omitempty"`
	IPv4Address string   `yaml:"ipv4_address,omitempty"`
	IPv6Address string   `yaml:"ipv6_address,omitempty"`
}

// MarshalYAML implements the Marshaller interface.
func (n Networks) MarshalYAML() (tag string, value interface{}, err error) {
	m := map[string]*Network{}
	for _, network := range n.Networks {
		m[network.Name] = network
	}
	return "", m, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (n *Networks) UnmarshalYAML(tag string, value interface{}) error {
	switch v := value.(type) {
	case []interface{}:
		n.Networks = []*Network{}
		for _, network := range v {
			name, ok := network.(string)
			if !ok {
				return fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			n.Networks = append(n.Networks, &Network{
				Name: name,
			})
		}
	case map[interface{}]interface{}:
		n.Networks = []*Network{}
		for mapKey, mapValue := range v {
			name, ok := mapKey.(string)
			if !ok {
				return fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			network, err := handleNetwork(name, mapValue)
			if err != nil {
				return err
			}
			n.Networks = append(n.Networks, network)
		}
	default:
		return fmt.Errorf("Failed to unmarshal Networks: %#v", value)
	}
	return nil
}

func handleNetwork(name string, value interface{}) (*Network, error) {
	if value == nil {
		return &Network{
			Name: name,
		}, nil
	}
	switch v := value.(type) {
	case map[interface{}]interface{}:
		network := &Network{
			Name: name,
		}
		for mapKey, mapValue := range v {
			name, ok := mapKey.(string)
			if !ok {
				return &Network{}, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			switch name {
			case "aliases":
				aliases, ok := mapValue.([]interface{})
				if !ok {
					return &Network{}, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", aliases, aliases)
				}
				network.Aliases = []string{}
				for _, alias := range aliases {
					network.Aliases = append(network.Aliases, alias.(string))
				}
			case "ipv4_address":
				network.IPv4Address = mapValue.(string)
			case "ipv6_address":
				network.IPv6Address = mapValue.(string)
			default:
				// Ignorer unknown keys ?
				continue
			}
		}
		return network, nil
	default:
		return &Network{}, fmt.Errorf("Failed to unmarshal Network: %#v", value)
	}
}
