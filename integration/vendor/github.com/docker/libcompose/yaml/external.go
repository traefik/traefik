package yaml

import (
	"fmt"
)

// External represents an external network entry in compose file.
// It can be a boolean (true|false) or have a name
type External struct {
	External bool
	Name     string
}

// MarshalYAML implements the Marshaller interface.
func (n External) MarshalYAML() (tag string, value interface{}, err error) {
	if n.Name == "" {
		return "", n.External, nil
	}
	return "", map[string]interface{}{
		"name": n.Name,
	}, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (n *External) UnmarshalYAML(tag string, value interface{}) error {
	switch v := value.(type) {
	case bool:
		n.External = v
	case map[interface{}]interface{}:
		for mapKey, mapValue := range v {
			switch mapKey {
			case "name":
				n.Name = mapValue.(string)
			default:
				// Ignore unknown keys
				continue
			}
		}
		n.External = true
	default:
		return fmt.Errorf("Failed to unmarshal External: %#v", value)
	}
	return nil
}
