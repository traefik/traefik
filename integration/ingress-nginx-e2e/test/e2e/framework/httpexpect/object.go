/*
Copyright 2022 The Kubernetes Authors.

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

package httpexpect

import (
	"reflect"

	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

const unavailableMsg = " (unavailable)"

// Object provides methods to inspect attached map[string]interface{} object
// (Go representation of JSON object).
type Object struct {
	chain chain
	value map[string]interface{}
}

func (o *Object) ValueEqual(key string, value interface{}) *Object {
	if !o.containsKey(key) {
		o.chain.fail("\nexpected object containing key '%s', but got:\n%s",
			key, dumpValue(o.value))
		return o
	}
	expected, ok := canonValue(&o.chain, value)
	if !ok {
		return o
	}
	if !reflect.DeepEqual(expected, o.value[key]) {
		o.chain.fail("\nexpected value for key '%s' equal to:\n%s\n\nbut got:\n%s\n\ndiff:\n%s",
			key,
			dumpValue(expected),
			dumpValue(o.value[key]),
			diffValues(expected, o.value[key]))
	}
	return o
}

func (o *Object) ContainsKey(key string) *Object {
	if !o.containsKey(key) {
		o.chain.fail("\nexpected object containing key '%s', but got:\n%s",
			key,
			dumpValue(o.value))
	}
	return o
}

func (o *Object) NotContainsKey(key string) *Object {
	if o.containsKey(key) {
		o.chain.fail("\nexpected object not containing key '%s', but got:\n%s",
			key, dumpValue(o.value))
	}
	return o
}

func (o *Object) containsKey(key string) bool {
	for k := range o.value {
		if k == key {
			return true
		}
	}
	return false
}

func diffValues(expected, actual interface{}) string {
	differ := gojsondiff.New()

	var diff gojsondiff.Diff

	switch ve := expected.(type) {
	case map[string]interface{}:
		if va, ok := actual.(map[string]interface{}); ok {
			diff = differ.CompareObjects(ve, va)
		} else {
			return unavailableMsg
		}
	case []interface{}:
		if va, ok := actual.([]interface{}); ok {
			diff = differ.CompareArrays(ve, va)
		} else {
			return unavailableMsg
		}
	default:
		return unavailableMsg
	}

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
	}
	f := formatter.NewAsciiFormatter(expected, config)

	str, err := f.Format(diff)
	if err != nil {
		return unavailableMsg
	}

	return "--- expected\n+++ actual\n" + str
}
