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

package atomiccache

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	var c Cache
	called := false
	get := func(k interface{}) interface{} {
		return c.Get(k, func() interface{} {
			called = true
			return fmt.Sprintf("v%d", k)
		})
	}
	got := get(1)
	if want := "v1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if !called {
		t.Error("getter not called, expected a call")
	}
	called = false
	got = get(1)
	if want := "v1"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if called {
		t.Error("getter unexpectedly called")
	}
}
