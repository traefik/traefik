package yaml

import (
	"fmt"

	"github.com/docker/engine-api/types/strslice"
	"github.com/flynn/go-shlex"
)

// Command represents a docker command, can be a string or an array of strings.
type Command strslice.StrSlice

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Command) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		*s = parts
	case string:
		parts, err := shlex.Split(value)
		if err != nil {
			return err
		}
		*s = parts
	default:
		return fmt.Errorf("Failed to unmarshal Command: %#v", value)
	}
	return nil
}
