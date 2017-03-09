package yaml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/engine-api/types/strslice"
)

// Stringorslice represents a string or an array of strings.
// Using engine-api Strslice and augment it with YAML marshalling stuff.
type Stringorslice strslice.StrSlice

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		*s = parts
	case string:
		*s = []string{value}
	default:
		return fmt.Errorf("Failed to unmarshal Stringorslice: %#v", value)
	}
	return nil
}

// SliceorMap represents a slice or a map of strings.
type SliceorMap map[string]string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *SliceorMap) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case map[interface{}]interface{}:
		parts := map[string]string{}
		for k, v := range value {
			if sk, ok := k.(string); ok {
				if sv, ok := v.(string); ok {
					parts[sk] = sv
				} else {
					return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
				}
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
			}
		}
		*s = parts
	case []interface{}:
		parts := map[string]string{}
		for _, s := range value {
			if str, ok := s.(string); ok {
				str := strings.TrimSpace(str)
				keyValueSlice := strings.SplitN(str, "=", 2)

				key := keyValueSlice[0]
				val := ""
				if len(keyValueSlice) == 2 {
					val = keyValueSlice[1]
				}
				parts[key] = val
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", s, s)
			}
		}
		*s = parts
	default:
		return fmt.Errorf("Failed to unmarshal SliceorMap: %#v", value)
	}
	return nil
}

// MaporEqualSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key=value' string.
type MaporEqualSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporEqualSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, "=")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

// ToMap returns the list of string as a map splitting using = the key=value
func (s *MaporEqualSlice) ToMap() map[string]string {
	return toMap(*s, "=")
}

// MaporColonSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key:value' string.
type MaporColonSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporColonSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, ":")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

// ToMap returns the list of string as a map splitting using = the key=value
func (s *MaporColonSlice) ToMap() map[string]string {
	return toMap(*s, ":")
}

// MaporSpaceSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key value' string.
type MaporSpaceSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporSpaceSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, " ")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

// ToMap returns the list of string as a map splitting using = the key=value
func (s *MaporSpaceSlice) ToMap() map[string]string {
	return toMap(*s, " ")
}

func unmarshalToStringOrSepMapParts(value interface{}, key string) ([]string, error) {
	switch value := value.(type) {
	case []interface{}:
		return toStrings(value)
	case map[interface{}]interface{}:
		return toSepMapParts(value, key)
	default:
		return nil, fmt.Errorf("Failed to unmarshal Map or Slice: %#v", value)
	}
}

func toSepMapParts(value map[interface{}]interface{}, sep string) ([]string, error) {
	if len(value) == 0 {
		return nil, nil
	}
	parts := make([]string, 0, len(value))
	for k, v := range value {
		if sk, ok := k.(string); ok {
			if sv, ok := v.(string); ok {
				parts = append(parts, sk+sep+sv)
			} else if sv, ok := v.(int64); ok {
				parts = append(parts, sk+sep+strconv.FormatInt(sv, 10))
			} else if v == nil {
				parts = append(parts, sk)
			} else {
				return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
			}
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
		}
	}
	return parts, nil
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}

func toMap(s []string, sep string) map[string]string {
	m := map[string]string{}
	for _, v := range s {
		values := strings.Split(v, sep)
		m[values[0]] = values[1]
	}
	return m
}
