// Copyright 2016, Google Inc.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package gax

import "testing"

func TestPathTemplateMatchRender(t *testing.T) {
	testCases := []struct {
		message  string
		template string
		path     string
		values   map[string]string
	}{
		{
			"base",
			"buckets/*/*/objects/*",
			"buckets/f/o/objects/bar",
			map[string]string{"$0": "f", "$1": "o", "$2": "bar"},
		},
		{
			"path wildcards",
			"bar/**/foo/*",
			"bar/foo/foo/foo/bar",
			map[string]string{"$0": "foo/foo", "$1": "bar"},
		},
		{
			"named binding",
			"buckets/{foo}/objects/*",
			"buckets/foo/objects/bar",
			map[string]string{"$0": "bar", "foo": "foo"},
		},
		{
			"named binding with colon",
			"buckets/{foo}/objects/*",
			"buckets/foo:boo/objects/bar",
			map[string]string{"$0": "bar", "foo": "foo:boo"},
		},
		{
			"named binding with complex patterns",
			"buckets/{foo=x/*/y/**}/objects/*",
			"buckets/x/foo/y/bar/baz/objects/quox",
			map[string]string{"$0": "quox", "foo": "x/foo/y/bar/baz"},
		},
		{
			"starts with slash",
			"/foo/*",
			"/foo/bar",
			map[string]string{"$0": "bar"},
		},
	}
	for _, testCase := range testCases {
		pt, err := NewPathTemplate(testCase.template)
		if err != nil {
			t.Errorf("[%s] Failed to parse template %s: %v", testCase.message, testCase.template, err)
			continue
		}
		values, err := pt.Match(testCase.path)
		if err != nil {
			t.Errorf("[%s] PathTemplate '%s' failed to match with '%s', %v", testCase.message, testCase.template, testCase.path, err)
			continue
		}
		for key, expected := range testCase.values {
			actual, ok := values[key]
			if !ok {
				t.Errorf("[%s] The matched data misses the value for %s", testCase.message, key)
				continue
			}
			delete(values, key)
			if actual != expected {
				t.Errorf("[%s] Failed to match: value for '%s' is expected '%s' but is actually '%s'", testCase.message, key, expected, actual)
			}
		}
		if len(values) != 0 {
			t.Errorf("[%s] The matched data has unexpected keys: %v", testCase.message, values)
		}
		built, err := pt.Render(testCase.values)
		if err != nil || built != testCase.path {
			t.Errorf("[%s] Built path '%s' is different from the expected '%s', %v", testCase.message, built, testCase.path, err)
		}
	}
}

func TestPathTemplateMatchFailure(t *testing.T) {
	testCases := []struct {
		message  string
		template string
		path     string
	}{
		{
			"too many paths",
			"buckets/*/*/objects/*",
			"buckets/f/o/o/objects/bar",
		},
		{
			"missing last path",
			"buckets/*/*/objects/*",
			"buckets/f/o/objects",
		},
		{
			"too many paths at end",
			"buckets/*/*/objects/*",
			"buckets/f/o/objects/too/long",
		},
	}
	for _, testCase := range testCases {
		pt, err := NewPathTemplate(testCase.template)
		if err != nil {
			t.Errorf("[%s] Failed to parse path %s: %v", testCase.message, testCase.template, err)
			continue
		}
		if values, err := pt.Match(testCase.path); err == nil {
			t.Errorf("[%s] PathTemplate %s doesn't expect to match %s, but succeeded somehow. Match result: %v", testCase.message, testCase.template, testCase.path, values)

		}
	}
}

func TestPathTemplateRenderTooManyValues(t *testing.T) {
	// Test cases where Render() succeeds but Match() doesn't return the same map.
	testCases := []struct {
		message  string
		template string
		values   map[string]string
		expected string
	}{
		{
			"too many",
			"bar/*/foo/*",
			map[string]string{"$0": "_1", "$1": "_2", "$2": "_3"},
			"bar/_1/foo/_2",
		},
	}
	for _, testCase := range testCases {
		pt, err := NewPathTemplate(testCase.template)
		if err != nil {
			t.Errorf("[%s] Failed to parse template %s (error %v)", testCase.message, testCase.template, err)
			continue
		}
		if result, err := pt.Render(testCase.values); err != nil || result != testCase.expected {
			t.Errorf("[%s] Failed to build the path (expected '%s' but returned '%s'", testCase.message, testCase.expected, result)
		}
	}
}

func TestPathTemplateParseErrors(t *testing.T) {
	testCases := []struct {
		message  string
		template string
	}{
		{
			"multiple path wildcards",
			"foo/**/bar/**",
		},
		{
			"recursive named bindings",
			"foo/{foo=foo/{bar}/baz/*}/baz/*",
		},
		{
			"complicated multiple path wildcards patterns",
			"foo/{foo=foo/**/bar/*}/baz/**",
		},
		{
			"consective slashes",
			"foo//bar",
		},
		{
			"invalid variable pattern",
			"foo/{foo=foo/*/}bar",
		},
		{
			"same name multiple times",
			"foo/{foo}/bar/{foo}",
		},
		{
			"empty string after '='",
			"foo/{foo=}/bar",
		},
	}
	for _, testCase := range testCases {
		if pt, err := NewPathTemplate(testCase.template); err == nil {
			t.Errorf("[%s] Template '%s' should fail to be parsed, but succeeded and returned %+v", testCase.message, testCase.template, pt)
		}
	}
}
