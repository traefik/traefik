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

package jsonutil

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/snapcore/snapd/strutil"
)

// DecodeWithNumber decodes input data using json.Decoder, ensuring numbers are preserved
// via json.Number data type. It errors out on invalid json or any excess input.
func DecodeWithNumber(r io.Reader, value interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	if err := dec.Decode(&value); err != nil {
		return err
	}
	if dec.More() {
		return fmt.Errorf("cannot parse json value")
	}
	return nil
}

// StructFields takes a pointer to a struct and a list of exceptions,
// and returns a list of the fields in the struct that are JSON-tagged
// and whose tag is not in the list of exceptions.
// The struct can be nil.
func StructFields(s interface{}, exceptions ...string) []string {
	st := reflect.TypeOf(s).Elem()
	num := st.NumField()
	fields := make([]string, 0, num)
	for i := 0; i < num; i++ {
		tag := st.Field(i).Tag.Get("json")
		idx := strings.IndexByte(tag, ',')
		if idx > -1 {
			tag = tag[:idx]
		}
		if tag != "" && !strutil.ListContains(exceptions, tag) {
			fields = append(fields, tag)
		}
	}

	return fields
}
