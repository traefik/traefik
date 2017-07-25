// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate ./update_version.sh

// Package version contains version information for Google Cloud Client
// Libraries for Go, as reported in request headers.
package version

import (
	"bytes"
	"runtime"
	"unicode"
)

// Repo is the current version of the client libraries in this
// repo. It should be a date in YYYYMMDD format.
const Repo = "20170210"

// Go returns the Go runtime version. The returned string
// has no whitespace.
func Go() string {
	return goVersion
}

var goVersion = removeWhitespace(runtime.Version())

func removeWhitespace(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		if unicode.IsSpace(r) {
			buf.WriteByte('_')
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
