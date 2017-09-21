package yaml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Build represents a build element in compose file.
// It can take multiple form in the compose file, hence this special type
type Build struct {
	Context    string
	Dockerfile string
	Args       map[string]*string
}

// MarshalYAML implements the Marshaller interface.
func (b Build) MarshalYAML() (interface{}, error) {
	m := map[string]interface{}{}
	if b.Context != "" {
		m["context"] = b.Context
	}
	if b.Dockerfile != "" {
		m["dockerfile"] = b.Dockerfile
	}
	if len(b.Args) > 0 {
		m["args"] = b.Args
	}
	return m, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (b *Build) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		b.Context = stringType
		return nil
	}

	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		for mapKey, mapValue := range mapType {
			switch mapKey {
			case "context":
				b.Context = mapValue.(string)
			case "dockerfile":
				b.Dockerfile = mapValue.(string)
			case "args":
				args, err := handleBuildArgs(mapValue)
				if err != nil {
					return err
				}
				b.Args = args
			default:
				// Ignore unknown keys
				continue
			}
		}
		return nil
	}

	return errors.New("Failed to unmarshal Build")
}

func handleBuildArgs(value interface{}) (map[string]*string, error) {
	var args map[string]*string
	switch v := value.(type) {
	case map[interface{}]interface{}:
		return handleBuildArgMap(v)
	case []interface{}:
		return handleBuildArgSlice(v)
	default:
		return args, fmt.Errorf("Failed to unmarshal Build args: %#v", value)
	}
}

func handleBuildArgSlice(s []interface{}) (map[string]*string, error) {
	var args = map[string]*string{}
	for _, arg := range s {
		// check if a value is provided
		switch v := strings.SplitN(arg.(string), "=", 2); len(v) {
		case 1:
			// if we have not specified a a value for this build arg, we assign it an ascii null value and query the environment
			// later when we build the service
			str := "\x00"
			args[v[0]] = &str
		case 2:
			// if we do have a value provided, we use it
			args[v[0]] = &v[1]
		}
	}
	return args, nil
}

func handleBuildArgMap(m map[interface{}]interface{}) (map[string]*string, error) {
	args := map[string]*string{}
	for mapKey, mapValue := range m {
		var argValue string
		name, ok := mapKey.(string)
		if !ok {
			return args, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
		}
		switch a := mapValue.(type) {
		case string:
			argValue = a
		case int:
			argValue = strconv.Itoa(a)
		case int64:
			argValue = strconv.Itoa(int(a))
		default:
			return args, fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", mapValue, mapValue)
		}
		args[name] = &argValue
	}
	return args, nil
}
