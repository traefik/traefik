package yaml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-units"
)

// StringorInt represents a string or an integer.
type StringorInt int64

// UnmarshalYAML implements the Unmarshaller interface.
func (s *StringorInt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = StringorInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := strconv.ParseInt(stringType, 10, 64)

		if err != nil {
			return err
		}
		*s = StringorInt(intType)
		return nil
	}

	return errors.New("Failed to unmarshal StringorInt")
}

// MemStringorInt represents a string or an integer
// the String supports notations like 10m for then Megabyte of memory
type MemStringorInt int64

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MemStringorInt) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = MemStringorInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := units.RAMInBytes(stringType)

		if err != nil {
			return err
		}
		*s = MemStringorInt(intType)
		return nil
	}

	return errors.New("Failed to unmarshal MemStringorInt")
}

// Stringorslice represents
// Using engine-api Strslice and augment it with YAML marshalling stuff. a string or an array of strings.
type Stringorslice strslice.StrSlice

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		*s = []string{stringType}
		return nil
	}

	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		parts, err := toStrings(sliceType)
		if err != nil {
			return err
		}
		*s = parts
		return nil
	}

	return errors.New("Failed to unmarshal Stringorslice")
}

// SliceorMap represents a slice or a map of strings.
type SliceorMap map[string]string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *SliceorMap) UnmarshalYAML(unmarshal func(interface{}) error) error {

	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		parts := map[string]string{}
		for _, s := range sliceType {
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
		return nil
	}

	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		parts := map[string]string{}
		for k, v := range mapType {
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
		return nil
	}

	return errors.New("Failed to unmarshal SliceorMap")
}

// MaporEqualSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key=value' string.
type MaporEqualSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parts, err := unmarshalToStringOrSepMapParts(unmarshal, "=")
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
func (s *MaporColonSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parts, err := unmarshalToStringOrSepMapParts(unmarshal, ":")
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
func (s *MaporSpaceSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parts, err := unmarshalToStringOrSepMapParts(unmarshal, " ")
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

func unmarshalToStringOrSepMapParts(unmarshal func(interface{}) error, key string) ([]string, error) {
	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		return toStrings(sliceType)
	}
	var mapType map[interface{}]interface{}
	if err := unmarshal(&mapType); err == nil {
		return toSepMapParts(mapType, key)
	}
	return nil, errors.New("Failed to unmarshal MaporSlice")
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
			} else if sv, ok := v.(int); ok {
				parts = append(parts, sk+sep+strconv.Itoa(sv))
			} else if sv, ok := v.(int64); ok {
				parts = append(parts, sk+sep+strconv.FormatInt(sv, 10))
			} else if sv, ok := v.(float64); ok {
				parts = append(parts, sk+sep+strconv.FormatFloat(sv, 'f', -1, 64))
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
		// Return everything past first sep
		values := strings.Split(v, sep)
		m[values[0]] = strings.SplitN(v, sep, 2)[1]
	}
	return m
}
