// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package strutil

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// OrderedMap is a map of strings to strings that preserves the
// insert order when calling "Keys()".
//
// Heavily based on the spread.Environment code (thanks for that!)
type OrderedMap struct {
	keys []string
	vals map[string]string
}

// NewOrderedMap creates a new ordered map initialized with the
// given pairs of strings.
func NewOrderedMap(pairs ...string) *OrderedMap {
	o := &OrderedMap{vals: make(map[string]string),
		keys: make([]string, len(pairs)/2),
	}
	for i := 0; i+1 < len(pairs); i += 2 {
		o.vals[pairs[i]] = pairs[i+1]
		o.keys[i/2] = pairs[i]
	}
	return o
}

// Keys returns a list of keys in the map sorted by insertion order
func (o *OrderedMap) Keys() []string {
	return append([]string(nil), o.keys...)
}

// Get returns the value for the given key
func (o *OrderedMap) Get(k string) string {
	return o.vals[k]
}

// Del removes the given key from the data structure
func (o *OrderedMap) Del(key string) {
	l := len(o.vals)
	delete(o.vals, key)
	if len(o.vals) != l {
		for i, k := range o.keys {
			if k == key {
				copy(o.keys[i:], o.keys[i+1:])
				o.keys = o.keys[:len(o.keys)-1]
			}
		}
	}
}

// Set adds the given key, value to the map. If the key already
// exists it is removed and the new value is put on the end.
func (o *OrderedMap) Set(k, v string) {
	o.Del(k)
	o.keys = append(o.keys, k)
	o.vals[k] = v
}

// Copy makes a copy of the map
func (o *OrderedMap) Copy() *OrderedMap {
	copy := &OrderedMap{}
	copy.keys = append([]string(nil), o.keys...)
	copy.vals = make(map[string]string)
	for k, v := range o.vals {
		copy.vals[k] = v
	}
	return copy
}

// UnmarshalYAML unmarshals a yaml string map and preserves the order
func (o *OrderedMap) UnmarshalYAML(u func(interface{}) error) error {
	var vals map[string]string
	if err := u(&vals); err != nil {
		return err
	}

	var seen = make(map[string]bool)
	var keys = make([]string, len(vals))
	var order yaml.MapSlice
	if err := u(&order); err != nil {
		return err
	}
	for i, item := range order {
		k, ok := item.Key.(string)
		_, good := vals[k]
		if !ok || !good {
			return fmt.Errorf("cannot read %q", item.Key)
		}
		if seen[k] {
			return fmt.Errorf("found duplicate key %q", k)
		}
		seen[k] = true
		keys[i] = k
	}
	o.keys = keys
	o.vals = vals
	return nil
}
