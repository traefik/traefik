/*
Copyright
*/
package main

import (
	"errors"
	"reflect"
)

// Invoke calls the specified method with the specified arguments on the specified interface.
// It uses the go(lang) reflect package.
func invoke(any interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	method := reflect.ValueOf(any).MethodByName(name)
	if method.IsValid() {
		return method.Call(inputs), nil
	}
	return nil, errors.New("Method not found: " + name)
}
