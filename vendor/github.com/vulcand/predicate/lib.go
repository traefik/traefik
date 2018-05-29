/*
Copyright 2016 Vulcand Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package predicate

import (
	"reflect"
	"strings"

	"github.com/gravitational/trace"
)

// GetStringMapValue is a helper function that returns property
// from map[string]string or map[string][]string
// the function returns empty value in case if key not found
// In case if map is nil, returns empty value as well
func GetStringMapValue(mapVal, keyVal interface{}) (interface{}, error) {
	key, ok := keyVal.(string)
	if !ok {
		return nil, trace.BadParameter("only string keys are supported")
	}
	switch m := mapVal.(type) {
	case map[string][]string:
		if len(m) == 0 {
			// to return nil with a proper type
			var n []string
			return n, nil
		}
		return m[key], nil
	case map[string]string:
		if len(m) == 0 {
			return "", nil
		}
		return m[key], nil
	default:
		return nil, trace.BadParameter("type %T is not supported", m)
	}
}

// BoolPredicate is a function without arguments that returns
// boolean value when called
type BoolPredicate func() bool

// Equals can compare complex objects, e.g. arrays of strings
// and strings together
func Equals(a interface{}, b interface{}) BoolPredicate {
	return func() bool {
		switch aval := a.(type) {
		case string:
			bval, ok := b.(string)
			return ok && aval == bval
		case []string:
			bval, ok := b.([]string)
			if !ok {
				return false
			}
			if len(aval) != len(bval) {
				return false
			}
			for i := range aval {
				if aval[i] != bval[i] {
					return false
				}
			}
			return true
		default:
			return false
		}
	}
}

// Contains checks if string slice contains a string
// Contains([]string{"a", "b"}, "b") -> true
func Contains(a interface{}, b interface{}) BoolPredicate {
	return func() bool {
		aval, ok := a.([]string)
		if !ok {
			return false
		}
		bval, ok := b.(string)
		if !ok {
			return false
		}
		for _, v := range aval {
			if v == bval {
				return true
			}
		}
		return false
	}
}

// And is a boolean predicate that calls two boolean predicates
// and returns result of && operation on their return values
func And(a, b BoolPredicate) BoolPredicate {
	return func() bool {
		return a() && b()
	}
}

// Or is a boolean predicate that calls two boolean predicates
// and returns result of || operation on their return values
func Or(a, b BoolPredicate) BoolPredicate {
	return func() bool {
		return a() || b()
	}
}

// GetFieldByTag returns a field from the object based on the tag
func GetFieldByTag(ival interface{}, tagName string, fieldNames []string) (interface{}, error) {
	if len(fieldNames) == 0 {
		return nil, trace.BadParameter("missing field names")
	}
	val := reflect.ValueOf(ival)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, trace.NotFound("field name %v is not found", strings.Join(fieldNames, "."))
	}
	fieldName := fieldNames[0]
	rest := fieldNames[1:]

	valType := val.Type()
	for i := 0; i < valType.NumField(); i++ {
		tagValue := valType.Field(i).Tag.Get(tagName)
		parts := strings.Split(tagValue, ",")
		if parts[0] == fieldName {
			value := val.Field(i).Interface()
			if len(rest) == 0 {
				return value, nil
			}
			return GetFieldByTag(value, tagName, rest)
		}
	}
	return nil, trace.NotFound("field name %v is not found", strings.Join(fieldNames, "."))
}
