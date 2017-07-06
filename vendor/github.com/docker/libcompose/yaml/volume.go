package yaml

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Volumes represents a list of service volumes in compose file.
// It has several representation, hence this specific struct.
type Volumes struct {
	Volumes []*Volume
}

// Volume represent a service volume
type Volume struct {
	Source      string `yaml:"-"`
	Destination string `yaml:"-"`
	AccessMode  string `yaml:"-"`
}

// Generate a hash string to detect service volume config changes
func (v *Volumes) HashString() string {
	if v == nil {
		return ""
	}
	result := []string{}
	for _, vol := range v.Volumes {
		result = append(result, vol.String())
	}
	sort.Strings(result)
	return strings.Join(result, ",")
}

// String implements the Stringer interface.
func (v *Volume) String() string {
	var paths []string
	if v.Source != "" {
		paths = []string{v.Source, v.Destination}
	} else {
		paths = []string{v.Destination}
	}
	if v.AccessMode != "" {
		paths = append(paths, v.AccessMode)
	}
	return strings.Join(paths, ":")
}

// MarshalYAML implements the Marshaller interface.
func (v Volumes) MarshalYAML() (interface{}, error) {
	vs := []string{}
	for _, volume := range v.Volumes {
		vs = append(vs, volume.String())
	}
	return vs, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (v *Volumes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []interface{}
	if err := unmarshal(&sliceType); err == nil {
		v.Volumes = []*Volume{}
		for _, volume := range sliceType {
			name, ok := volume.(string)
			if !ok {
				return fmt.Errorf("Cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			elts := strings.SplitN(name, ":", 3)
			var vol *Volume
			switch {
			case len(elts) == 1:
				vol = &Volume{
					Destination: elts[0],
				}
			case len(elts) == 2:
				vol = &Volume{
					Source:      elts[0],
					Destination: elts[1],
				}
			case len(elts) == 3:
				vol = &Volume{
					Source:      elts[0],
					Destination: elts[1],
					AccessMode:  elts[2],
				}
			default:
				// FIXME
				return fmt.Errorf("")
			}
			v.Volumes = append(v.Volumes, vol)
		}
		return nil
	}

	return errors.New("Failed to unmarshal Volumes")
}
