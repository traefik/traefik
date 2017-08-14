package yaml

// External represents an external network entry in compose file.
// It can be a boolean (true|false) or have a name
type External struct {
	External bool
	Name     string
}

// MarshalYAML implements the Marshaller interface.
func (n External) MarshalYAML() (interface{}, error) {
	if n.Name == "" {
		return n.External, nil
	}
	return map[string]interface{}{
		"name": n.Name,
	}, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (n *External) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&n.External); err == nil {
		return nil
	}
	var dummyExternal struct {
		Name string
	}

	err := unmarshal(&dummyExternal)
	if err != nil {
		return err
	}
	n.Name = dummyExternal.Name
	n.External = true

	return nil
}
