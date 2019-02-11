// Package jsonhooks adds hooks that are automatically called before JSON marshaling (PreMarshalJSON) and
// after JSON unmarshaling (PostUnmarshalJSON). It does not do so recursively.
package jsonhooks

import (
	"encoding/json"
	"reflect"
)

// Marshal wraps encoding/json.Marshal, calls v.PreMarshalJSON() if it exists
func Marshal(v interface{}) ([]byte, error) {
	if ImplementsPreJSONMarshaler(v) {
		err := v.(PreJSONMarshaler).PreMarshalJSON()
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(v)
}

// Unmarshal wraps encoding/json.Unmarshal, calls v.PostUnmarshalJSON() if it exists
func Unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	if ImplementsPostJSONUnmarshaler(v) {
		err := v.(PostJSONUnmarshaler).PostUnmarshalJSON()
		if err != nil {
			return err
		}
	}

	return nil
}

// PreJSONMarshaler infers support for the PreMarshalJSON pre-hook
type PreJSONMarshaler interface {
	PreMarshalJSON() error
}

// ImplementsPreJSONMarshaler checks for support for the PreMarshalJSON pre-hook
func ImplementsPreJSONMarshaler(v interface{}) bool {
	value := reflect.ValueOf(v)
	if !value.IsValid() {
		return false
	}

	_, ok := value.Interface().(PreJSONMarshaler)
	return ok
}

// PostJSONUnmarshaler infers support for the PostUnmarshalJSON post-hook
type PostJSONUnmarshaler interface {
	PostUnmarshalJSON() error
}

// ImplementsPostJSONUnmarshaler checks for support for the PostUnmarshalJSON post-hook
func ImplementsPostJSONUnmarshaler(v interface{}) bool {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return false
	}

	_, ok := value.Interface().(PostJSONUnmarshaler)
	return ok
}
