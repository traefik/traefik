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
	"regexp"
	"strings"
)

// String provides methods to inspect attached string value
// (Go representation of JSON string).
type String struct {
	chain chain
	value string
}

// Raw returns underlying value attached to String.
// This is the value originally passed to NewString.
func (s *String) Raw() string {
	return s.value
}

// Empty succeeds if string is empty.
func (s *String) Empty() *String {
	return s.Equal("")
}

// NotEmpty succeeds if string is non-empty.
func (s *String) NotEmpty() *String {
	return s.NotEqual("")
}

// Equal succeeds if string is equal to given Go string.
func (s *String) Equal(value string) *String {
	if !(s.value == value) {
		s.chain.fail("\nexpected string equal to:\n %q\n\nbut got:\n %q", value, s.value)
	}
	return s
}

// NotEqual succeeds if string is not equal to given Go string.
func (s *String) NotEqual(value string) *String {
	if !(s.value != value) {
		s.chain.fail("\nexpected string not equal to:\n %q", value)
	}
	return s
}

// Contains succeeds if string contains given Go string as a substring.
func (s *String) Contains(value string) *String {
	if !strings.Contains(s.value, value) {
		s.chain.fail(
			"\nexpected string containing substring:\n %q\n\nbut got:\n %q",
			value, s.value)
	}
	return s
}

// NotContains succeeds if string doesn't contain Go string as a substring.
func (s *String) NotContains(value string) *String {
	if strings.Contains(s.value, value) {
		s.chain.fail("\nexpected string not containing substring:\n %q\n\nbut got:\n %q", value, s.value)
	}
	return s
}

// ContainsFold succeeds if string contains given Go string as a substring after
// applying Unicode case-folding (so it's a case-insensitive match).
func (s *String) ContainsFold(value string) *String {
	if !strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail("\nexpected string containing substring (case-insensitive):\n %q"+"\n\nbut got:\n %q", value, s.value)
	}
	return s
}

// NotContainsFold succeeds if string doesn't contain given Go string as a substring
// after applying Unicode case-folding (so it's a case-insensitive match).
//
// Example:
//
//	str := NewString(t, "Hello")
//	str.NotContainsFold("BYE")
func (s *String) NotContainsFold(value string) *String {
	if strings.Contains(strings.ToLower(s.value), strings.ToLower(value)) {
		s.chain.fail("\nexpected string not containing substring (case-insensitive):\n %q"+"\n\nbut got:\n %q", value, s.value)
	}
	return s
}

// Match matches the string with given regexp and returns a new Match object
// with found submatches.
func (s *String) Match(re string) *Match {
	r, err := regexp.Compile(re)
	if err != nil {
		s.chain.fail(err.Error())
		return makeMatch(s.chain, nil, nil)
	}

	m := r.FindStringSubmatch(s.value)
	if m == nil {
		s.chain.fail("\nexpected string matching regexp:\n `%s`\n\nbut got:\n %q", re, s.value)
		return makeMatch(s.chain, nil, nil)
	}

	return makeMatch(s.chain, m, r.SubexpNames())
}
