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
	CacheFrom  []*string
	Labels     map[string]*string
	// TODO: ShmSize (can be a string or int?) for v3.5
	Target string
	// Note: as of Sep 2018 this is undocumented but supported by docker-compose
	Network string
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
	if len(b.CacheFrom) > 0 {
		m["cache_from"] = b.CacheFrom
	}
	if len(b.Labels) > 0 {
		m["labels"] = b.Labels
	}
	if b.Target != "" {
		m["target"] = b.Target
	}
	if b.Network != "" {
		m["network"] = b.Network
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
			case "cache_from":
				cacheFrom, err := handleBuildCacheFrom(mapValue)
				if err != nil {
					return err
				}
				b.CacheFrom = cacheFrom
			case "labels":
				labels, err := handleBuildLabels(mapValue)
				if err != nil {
					return err
				}
				b.Labels = labels
			case "target":
				b.Target = mapValue.(string)
			case "network":
				b.Network = mapValue.(string)
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
		return handleBuildOptionMap(v)
	case []interface{}:
		return handleBuildArgsSlice(v)
	default:
		return args, fmt.Errorf("Failed to unmarshal Build args: %#v", value)
	}
}

func handleBuildCacheFrom(value interface{}) ([]*string, error) {
	var cacheFrom []*string
	switch v := value.(type) {
	case []interface{}:
		return handleBuildCacheFromSlice(v)
	default:
		return cacheFrom, fmt.Errorf("Failed to unmarshal Build cache_from: %#v", value)
	}
}

func handleBuildLabels(value interface{}) (map[string]*string, error) {
	var labels map[string]*string
	switch v := value.(type) {
	case map[interface{}]interface{}:
		return handleBuildOptionMap(v)
	default:
		return labels, fmt.Errorf("Failed to unmarshal Build labels: %#v", value)
	}
}

func handleBuildCacheFromSlice(s []interface{}) ([]*string, error) {
	var args = []*string{}
	for _, arg := range s {
		strArg := arg.(string)
		args = append(args, &strArg)
	}
	return args, nil
}

func handleBuildArgsSlice(s []interface{}) (map[string]*string, error) {
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

// Used for args and labels
func handleBuildOptionMap(m map[interface{}]interface{}) (map[string]*string, error) {
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
