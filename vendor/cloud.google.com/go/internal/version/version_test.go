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

package version

import (
	"regexp"
	"testing"
)

func TestGo(t *testing.T) {
	got := Go()
	want := `^\S+$`
	match, err := regexp.MatchString(want, got)
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Errorf("got %q, want match of regexp %q", got, want)
	}
}

func TestRemoveWhitespace(t *testing.T) {
	for _, test := range []struct {
		in, want string
	}{
		{"", ""},
		{"go1.7", "go1.7"},
		{" a b c ", "_a_b_c_"},
		{"a\tb\t c\n", "a_b__c_"},
	} {
		if got := removeWhitespace(test.in); got != test.want {
			t.Errorf("removeWhitespace(%q) = %q, want %q", test.in, got, test.want)
		}
	}
}
