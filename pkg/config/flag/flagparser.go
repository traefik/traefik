package flag

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/parser"
)

// Parse parses the command-line flag arguments into a map,
// using the type information in element to discriminate whether a flag is supposed to be a bool,
// and other such ambiguities.
func Parse(args []string, element interface{}) (map[string]string, error) {
	f := flagSet{
		flagTypes: getFlagTypes(element),
		args:      args,
		values:    make(map[string]string),
		keys:      make(map[string]string),
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
	keys      map[string]string
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

	flagType := f.getFlagType(name)
	if flagType == reflect.Bool || flagType == reflect.Ptr {
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

func (f *flagSet) setValue(name, value string) {
	srcKey := parser.DefaultRootName + "." + name
	neutralKey := strings.ToLower(srcKey)

	key, ok := f.keys[neutralKey]
	if !ok {
		f.keys[neutralKey] = srcKey
		key = srcKey
	}

	v, ok := f.values[key]
	if ok && f.getFlagType(name) == reflect.Slice {
		f.values[key] = v + "," + value
		return
	}

	f.values[key] = value
}

func (f *flagSet) getFlagType(name string) reflect.Kind {
	neutral := strings.ToLower(name)

	kind, ok := f.flagTypes[neutral]
	if ok {
		return kind
	}

	for n, k := range f.flagTypes {
		if strings.Contains(n, parser.MapNamePlaceholder) {
			p := strings.NewReplacer(".", `\.`, parser.MapNamePlaceholder, `([^.]+)`).Replace(n)
			if regexp.MustCompile(p).MatchString(neutral) {
				return k
			}
		}
	}

	return reflect.Invalid
}
