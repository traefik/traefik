// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pretty

import (
	"fmt"
	"reflect"
	"sort"
)

func isZeroVal(val reflect.Value) bool {
	if !val.CanInterface() {
		return false
	}
	z := reflect.Zero(val.Type()).Interface()
	return reflect.DeepEqual(val.Interface(), z)
}

func (c *Config) val2node(val reflect.Value) node {
	// TODO(kevlar): pointer tracking?

	if c.PrintStringers && val.CanInterface() {
		stringer, ok := val.Interface().(fmt.Stringer)
		if ok {
			return stringVal(stringer.String())
		}
	}

	switch kind := val.Kind(); kind {
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return rawVal("nil")
		}
		return c.val2node(val.Elem())
	case reflect.String:
		return stringVal(val.String())
	case reflect.Slice, reflect.Array:
		n := list{}
		length := val.Len()
		for i := 0; i < length; i++ {
			n = append(n, c.val2node(val.Index(i)))
		}
		return n
	case reflect.Map:
		n := keyvals{}
		keys := val.MapKeys()
		for _, key := range keys {
			// TODO(kevlar): Support arbitrary type keys?
			n = append(n, keyval{compactString(c.val2node(key)), c.val2node(val.MapIndex(key))})
		}
		sort.Sort(n)
		return n
	case reflect.Struct:
		n := keyvals{}
		typ := val.Type()
		fields := typ.NumField()
		for i := 0; i < fields; i++ {
			sf := typ.Field(i)
			if !c.IncludeUnexported && sf.PkgPath != "" {
				continue
			}
			field := val.Field(i)
			if c.SkipZeroFields && isZeroVal(field) {
				continue
			}
			n = append(n, keyval{sf.Name, c.val2node(field)})
		}
		return n
	case reflect.Bool:
		if val.Bool() {
			return rawVal("true")
		}
		return rawVal("false")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rawVal(fmt.Sprintf("%d", val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rawVal(fmt.Sprintf("%d", val.Uint()))
	case reflect.Uintptr:
		return rawVal(fmt.Sprintf("0x%X", val.Uint()))
	case reflect.Float32, reflect.Float64:
		return rawVal(fmt.Sprintf("%v", val.Float()))
	case reflect.Complex64, reflect.Complex128:
		return rawVal(fmt.Sprintf("%v", val.Complex()))
	}

	// Fall back to the default %#v if we can
	if val.CanInterface() {
		return rawVal(fmt.Sprintf("%#v", val.Interface()))
	}

	return rawVal(val.String())
}
