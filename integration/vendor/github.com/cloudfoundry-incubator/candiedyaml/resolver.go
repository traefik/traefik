/*
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

package candiedyaml

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var byteSliceType = reflect.TypeOf([]byte(nil))

var binary_tags = [][]byte{[]byte("!binary"), []byte(yaml_BINARY_TAG)}
var bool_values map[string]bool
var null_values map[string]bool

var signs = []byte{'-', '+'}
var nulls = []byte{'~', 'n', 'N'}
var bools = []byte{'t', 'T', 'f', 'F', 'y', 'Y', 'n', 'N', 'o', 'O'}

var timestamp_regexp *regexp.Regexp
var ymd_regexp *regexp.Regexp

func init() {
	bool_values = make(map[string]bool)
	bool_values["y"] = true
	bool_values["yes"] = true
	bool_values["n"] = false
	bool_values["no"] = false
	bool_values["true"] = true
	bool_values["false"] = false
	bool_values["on"] = true
	bool_values["off"] = false

	null_values = make(map[string]bool)
	null_values["~"] = true
	null_values["null"] = true
	null_values["Null"] = true
	null_values["NULL"] = true

	timestamp_regexp = regexp.MustCompile("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:(?:[Tt]|[ \t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \t]*(?:Z|([-+][0-9][0-9]?)(?::([0-9][0-9])?)?))?)?$")
	ymd_regexp = regexp.MustCompile("^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)$")
}

func resolve(event yaml_event_t, v reflect.Value, useNumber bool) (string, error) {
	val := string(event.value)

	if null_values[val] {
		v.Set(reflect.Zero(v.Type()))
		return yaml_NULL_TAG, nil
	}

	switch v.Kind() {
	case reflect.String:
		if useNumber && v.Type() == numberType {
			tag, i := resolveInterface(event, useNumber)
			if n, ok := i.(Number); ok {
				v.Set(reflect.ValueOf(n))
				return tag, nil
			}
			return "", fmt.Errorf("Not a number: '%s' at %s", event.value, event.start_mark)
		}

		return resolve_string(val, v, event)
	case reflect.Bool:
		return resolve_bool(val, v, event)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return resolve_int(val, v, useNumber, event)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return resolve_uint(val, v, useNumber, event)
	case reflect.Float32, reflect.Float64:
		return resolve_float(val, v, useNumber, event)
	case reflect.Interface:
		_, i := resolveInterface(event, useNumber)
		if i != nil {
			v.Set(reflect.ValueOf(i))
		} else {
			v.Set(reflect.Zero(v.Type()))
		}

	case reflect.Struct:
		return resolve_time(val, v, event)
	case reflect.Slice:
		if v.Type() != byteSliceType {
			return "", fmt.Errorf("Cannot resolve %s into %s at %s", val, v.String(), event.start_mark)
		}
		b, err := decode_binary(event.value, event)
		if err != nil {
			return "", err
		}

		v.Set(reflect.ValueOf(b))
	default:
		return "", fmt.Errorf("Unknown resolution for '%s' using %s at %s", val, v.String(), event.start_mark)
	}

	return yaml_STR_TAG, nil
}

func hasBinaryTag(event yaml_event_t) bool {
	for _, tag := range binary_tags {
		if bytes.Equal(event.tag, tag) {
			return true
		}
	}
	return false
}

func decode_binary(value []byte, event yaml_event_t) ([]byte, error) {
	b := make([]byte, base64.StdEncoding.DecodedLen(len(value)))
	n, err := base64.StdEncoding.Decode(b, value)
	if err != nil {
		return nil, fmt.Errorf("Invalid base64 text: '%s' at %s", string(b), event.start_mark)
	}
	return b[:n], nil
}

func resolve_string(val string, v reflect.Value, event yaml_event_t) (string, error) {
	if len(event.tag) > 0 {
		if hasBinaryTag(event) {
			b, err := decode_binary(event.value, event)
			if err != nil {
				return "", err
			}
			val = string(b)
		}
	}
	v.SetString(val)
	return yaml_STR_TAG, nil
}

func resolve_bool(val string, v reflect.Value, event yaml_event_t) (string, error) {
	b, found := bool_values[strings.ToLower(val)]
	if !found {
		return "", fmt.Errorf("Invalid boolean: '%s' at %s", val, event.start_mark)
	}

	v.SetBool(b)
	return yaml_BOOL_TAG, nil
}

func resolve_int(val string, v reflect.Value, useNumber bool, event yaml_event_t) (string, error) {
	original := val
	val = strings.Replace(val, "_", "", -1)
	var value uint64

	isNumberValue := v.Type() == numberType

	sign := int64(1)
	if val[0] == '-' {
		sign = -1
		val = val[1:]
	} else if val[0] == '+' {
		val = val[1:]
	}

	base := 0
	if val == "0" {
		if isNumberValue {
			v.SetString("0")
		} else {
			v.Set(reflect.Zero(v.Type()))
		}

		return yaml_INT_TAG, nil
	}

	if strings.HasPrefix(val, "0o") {
		base = 8
		val = val[2:]
	}

	value, err := strconv.ParseUint(val, base, 64)
	if err != nil {
		return "", fmt.Errorf("Invalid integer: '%s' at %s", original, event.start_mark)
	}

	var val64 int64
	if value <= math.MaxInt64 {
		val64 = int64(value)
		if sign == -1 {
			val64 = -val64
		}
	} else if sign == -1 && value == uint64(math.MaxInt64)+1 {
		val64 = math.MinInt64
	} else {
		return "", fmt.Errorf("Invalid integer: '%s' at %s", original, event.start_mark)
	}

	if isNumberValue {
		v.SetString(strconv.FormatInt(val64, 10))
	} else {
		if v.OverflowInt(val64) {
			return "", fmt.Errorf("Invalid integer: '%s' at %s", original, event.start_mark)
		}
		v.SetInt(val64)
	}

	return yaml_INT_TAG, nil
}

func resolve_uint(val string, v reflect.Value, useNumber bool, event yaml_event_t) (string, error) {
	original := val
	val = strings.Replace(val, "_", "", -1)
	var value uint64

	isNumberValue := v.Type() == numberType

	if val[0] == '-' {
		return "", fmt.Errorf("Unsigned int with negative value: '%s' at %s", original, event.start_mark)
	}

	if val[0] == '+' {
		val = val[1:]
	}

	base := 0
	if val == "0" {
		if isNumberValue {
			v.SetString("0")
		} else {
			v.Set(reflect.Zero(v.Type()))
		}

		return yaml_INT_TAG, nil
	}

	if strings.HasPrefix(val, "0o") {
		base = 8
		val = val[2:]
	}

	value, err := strconv.ParseUint(val, base, 64)
	if err != nil {
		return "", fmt.Errorf("Invalid unsigned integer: '%s' at %s", val, event.start_mark)
	}

	if isNumberValue {
		v.SetString(strconv.FormatUint(value, 10))
	} else {
		if v.OverflowUint(value) {
			return "", fmt.Errorf("Invalid unsigned integer: '%s' at %s", val, event.start_mark)
		}

		v.SetUint(value)
	}

	return yaml_INT_TAG, nil
}

func resolve_float(val string, v reflect.Value, useNumber bool, event yaml_event_t) (string, error) {
	val = strings.Replace(val, "_", "", -1)
	var value float64

	isNumberValue := v.Type() == numberType
	typeBits := 64
	if !isNumberValue {
		typeBits = v.Type().Bits()
	}

	sign := 1
	if val[0] == '-' {
		sign = -1
		val = val[1:]
	} else if val[0] == '+' {
		val = val[1:]
	}

	valLower := strings.ToLower(val)
	if valLower == ".inf" {
		value = math.Inf(sign)
	} else if valLower == ".nan" {
		value = math.NaN()
	} else {
		var err error
		value, err = strconv.ParseFloat(val, typeBits)
		value *= float64(sign)

		if err != nil {
			return "", fmt.Errorf("Invalid float: '%s' at %s", val, event.start_mark)
		}
	}

	if isNumberValue {
		v.SetString(strconv.FormatFloat(value, 'g', -1, typeBits))
	} else {
		if v.OverflowFloat(value) {
			return "", fmt.Errorf("Invalid float: '%s' at %s", val, event.start_mark)
		}

		v.SetFloat(value)
	}

	return yaml_FLOAT_TAG, nil
}

func resolve_time(val string, v reflect.Value, event yaml_event_t) (string, error) {
	var parsedTime time.Time
	matches := ymd_regexp.FindStringSubmatch(val)
	if len(matches) > 0 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		parsedTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	} else {
		matches = timestamp_regexp.FindStringSubmatch(val)
		if len(matches) == 0 {
			return "", fmt.Errorf("Invalid timestamp: '%s' at %s", val, event.start_mark)
		}

		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		min, _ := strconv.Atoi(matches[5])
		sec, _ := strconv.Atoi(matches[6])

		nsec := 0
		if matches[7] != "" {
			millis, _ := strconv.Atoi(matches[7])
			nsec = int(time.Duration(millis) * time.Millisecond)
		}

		loc := time.UTC
		if matches[8] != "" {
			sign := matches[8][0]
			hr, _ := strconv.Atoi(matches[8][1:])
			min := 0
			if matches[9] != "" {
				min, _ = strconv.Atoi(matches[9])
			}

			zoneOffset := (hr*60 + min) * 60
			if sign == '-' {
				zoneOffset = -zoneOffset
			}

			loc = time.FixedZone("", zoneOffset)
		}
		parsedTime = time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)
	}

	v.Set(reflect.ValueOf(parsedTime))
	return "", nil
}

func resolveInterface(event yaml_event_t, useNumber bool) (string, interface{}) {
	val := string(event.value)
	if len(event.tag) == 0 && !event.implicit {
		return "", val
	}

	if len(val) == 0 {
		return yaml_NULL_TAG, nil
	}

	var result interface{}

	sign := false
	c := val[0]
	switch {
	case bytes.IndexByte(signs, c) != -1:
		sign = true
		fallthrough
	case c >= '0' && c <= '9':
		i := int64(0)
		result = &i
		if useNumber {
			var n Number
			result = &n
		}

		v := reflect.ValueOf(result).Elem()
		if _, err := resolve_int(val, v, useNumber, event); err == nil {
			return yaml_INT_TAG, v.Interface()
		}

		f := float64(0)
		result = &f
		if useNumber {
			var n Number
			result = &n
		}

		v = reflect.ValueOf(result).Elem()
		if _, err := resolve_float(val, v, useNumber, event); err == nil {
			return yaml_FLOAT_TAG, v.Interface()
		}

		if !sign {
			t := time.Time{}
			if _, err := resolve_time(val, reflect.ValueOf(&t).Elem(), event); err == nil {
				return "", t
			}
		}
	case bytes.IndexByte(nulls, c) != -1:
		if null_values[val] {
			return yaml_NULL_TAG, nil
		}
		b := false
		if _, err := resolve_bool(val, reflect.ValueOf(&b).Elem(), event); err == nil {
			return yaml_BOOL_TAG, b
		}
	case c == '.':
		f := float64(0)
		result = &f
		if useNumber {
			var n Number
			result = &n
		}

		v := reflect.ValueOf(result).Elem()
		if _, err := resolve_float(val, v, useNumber, event); err == nil {
			return yaml_FLOAT_TAG, v.Interface()
		}
	case bytes.IndexByte(bools, c) != -1:
		b := false
		if _, err := resolve_bool(val, reflect.ValueOf(&b).Elem(), event); err == nil {
			return yaml_BOOL_TAG, b
		}
	}

	if hasBinaryTag(event) {
		bytes, err := decode_binary(event.value, event)
		if err == nil {
			return yaml_BINARY_TAG, bytes
		}
	}

	return yaml_STR_TAG, val
}
