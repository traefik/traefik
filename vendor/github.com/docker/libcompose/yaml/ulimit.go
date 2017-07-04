package yaml

import (
	"errors"
	"fmt"
	"sort"
)

// Ulimits represents a list of Ulimit.
// It is, however, represented in yaml as keys (and thus map in Go)
type Ulimits struct {
	Elements []Ulimit
}

// MarshalYAML implements the Marshaller interface.
func (u Ulimits) MarshalYAML() (interface{}, error) {
	ulimitMap := make(map[string]Ulimit)
	for _, ulimit := range u.Elements {
		ulimitMap[ulimit.Name] = ulimit
	}
	return ulimitMap, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (u *Ulimits) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ulimits := make(map[string]Ulimit)

	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		for mapKey, mapValue := range mapType {
			name, ok := mapKey.(string)
			if !ok {
				return fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			var soft, hard int64
			switch mv := mapValue.(type) {
			case int:
				soft = int64(mv)
				hard = int64(mv)
			case map[interface{}]interface{}:
				if len(mv) != 2 {
					return fmt.Errorf("Failed to unmarshal Ulimit: %#v", mapValue)
				}
				for mkey, mvalue := range mv {
					switch mkey {
					case "soft":
						soft = int64(mvalue.(int))
					case "hard":
						hard = int64(mvalue.(int))
					default:
						// FIXME(vdemeester) Should we ignore or fail ?
						continue
					}
				}
			default:
				return fmt.Errorf("Failed to unmarshal Ulimit: %v, %T", mapValue, mapValue)
			}
			ulimits[name] = Ulimit{
				Name: name,
				ulimitValues: ulimitValues{
					Soft: soft,
					Hard: hard,
				},
			}
		}
		keys := make([]string, 0, len(ulimits))
		for key := range ulimits {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			u.Elements = append(u.Elements, ulimits[key])
		}
		return nil
	}

	return errors.New("Failed to unmarshal Ulimit")
}

// Ulimit represents ulimit information.
type Ulimit struct {
	ulimitValues
	Name string
}

type ulimitValues struct {
	Soft int64 `yaml:"soft"`
	Hard int64 `yaml:"hard"`
}

// MarshalYAML implements the Marshaller interface.
func (u Ulimit) MarshalYAML() (interface{}, error) {
	if u.Soft == u.Hard {
		return u.Soft, nil
	}
	return u.ulimitValues, nil
}

// NewUlimit creates a Ulimit based on the specified parts.
func NewUlimit(name string, soft int64, hard int64) Ulimit {
	return Ulimit{
		Name: name,
		ulimitValues: ulimitValues{
			Soft: soft,
			Hard: hard,
		},
	}
}
