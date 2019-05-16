package flag

import (
	"fmt"
	"reflect"
	"strings"
)

// Parse Parses CLI flags/args to a map.
func Parse(args []string, element interface{}) (map[string]string, error) {
	f := flagSet{
		flagTypes: getFlagTypes(element),
		args:      args,
		values:    make(map[string]string),
	}

	for {
		seen, err := f.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		return nil, err
	}
	return f.values, nil
}

type flagSet struct {
	flagTypes map[string]reflect.Kind
	args      []string
	values    map[string]string
}

func (f *flagSet) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}

	s := f.args[0]
	if len(s) < 2 || s[0] != '-' {
		return false, nil
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 { // "--" terminates the flags
			f.args = f.args[1:]
			return false, nil
		}
	}

	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return false, fmt.Errorf("bad flag syntax: %s", s)
	}

	// it's a flag. does it have an argument?
	f.args = f.args[1:]
	hasValue := false
	value := ""
	for i := 1; i < len(name); i++ { // equals cannot be first
		if name[i] == '=' {
			value = name[i+1:]
			hasValue = true
			name = name[0:i]
			break
		}
	}

	if hasValue {
		f.setValue(name, value)
		return true, nil
	}

	if f.flagTypes[name] == reflect.Bool || f.flagTypes[name] == reflect.Ptr {
		f.setValue(name, "true")
		return true, nil
	}

	if len(f.args) > 0 {
		// value is the next arg
		hasValue = true
		value, f.args = f.args[0], f.args[1:]
	}

	if !hasValue {
		return false, fmt.Errorf("flag needs an argument: -%s", name)
	}

	f.setValue(name, value)
	return true, nil
}

func (f *flagSet) setValue(name string, value string) {
	n := strings.ToLower("traefik." + name)
	v, ok := f.values[n]

	if ok && f.flagTypes[name] == reflect.Slice {
		f.values[n] = v + "," + value
		return
	}

	f.values[n] = value
}
