package yaml

import (
	"fmt"
	"sort"
)

// Ulimits represents a list of Ulimit.
// It is, however, represented in yaml as keys (and thus map in Go)
type Ulimits struct {
	Elements []Ulimit
}

// MarshalYAML implements the Marshaller interface.
func (u Ulimits) MarshalYAML() (tag string, value interface{}, err error) {
	ulimitMap := make(map[string]Ulimit)
	for _, ulimit := range u.Elements {
		ulimitMap[ulimit.Name] = ulimit
	}
	return "", ulimitMap, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (u *Ulimits) UnmarshalYAML(tag string, value interface{}) error {
	ulimits := make(map[string]Ulimit)
	switch v := value.(type) {
	case map[interface{}]interface{}:
		for mapKey, mapValue := range v {
			name, ok := mapKey.(string)
			if !ok {
				return fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			var soft, hard int64
			switch mv := mapValue.(type) {
			case int64:
				soft = mv
				hard = mv
			case map[interface{}]interface{}:
				if len(mv) != 2 {
					return fmt.Errorf("Failed to unmarshal Ulimit: %#v", mapValue)
				}
				for mkey, mvalue := range mv {
					switch mkey {
					case "soft":
						soft = mvalue.(int64)
					case "hard":
						hard = mvalue.(int64)
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
	default:
		return fmt.Errorf("Failed to unmarshal Ulimit: %#v", value)
	}
	return nil
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
func (u Ulimit) MarshalYAML() (tag string, value interface{}, err error) {
	if u.Soft == u.Hard {
		return "", u.Soft, nil
	}
	return "", u.ulimitValues, err
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
