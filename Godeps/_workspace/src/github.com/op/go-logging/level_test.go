// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import "testing"

func TestLevelString(t *testing.T) {
	// Make sure all levels can be converted from string -> constant -> string
	for _, name := range levelNames {
		level, err := LogLevel(name)
		if err != nil {
			t.Errorf("failed to get level: %v", err)
			continue
		}

		if level.String() != name {
			t.Errorf("invalid level conversion: %v != %v", level, name)
		}
	}
}

func TestLevelLogLevel(t *testing.T) {
	tests := []struct {
		expected Level
		level    string
	}{
		{-1, "bla"},
		{INFO, "iNfO"},
		{ERROR, "error"},
		{WARNING, "warninG"},
	}

	for _, test := range tests {
		level, err := LogLevel(test.level)
		if err != nil {
			if test.expected == -1 {
				continue
			} else {
				t.Errorf("failed to convert %s: %s", test.level, err)
			}
		}
		if test.expected != level {
			t.Errorf("failed to convert %s to level: %s != %s", test.level, test.expected, level)
		}
	}
}

func TestLevelModuleLevel(t *testing.T) {
	backend := NewMemoryBackend(128)

	leveled := AddModuleLevel(backend)
	leveled.SetLevel(NOTICE, "")
	leveled.SetLevel(ERROR, "foo")
	leveled.SetLevel(INFO, "foo.bar")
	leveled.SetLevel(WARNING, "bar")

	expected := []struct {
		level  Level
		module string
	}{
		{NOTICE, ""},
		{NOTICE, "something"},
		{ERROR, "foo"},
		{INFO, "foo.bar"},
		{WARNING, "bar"},
	}

	for _, e := range expected {
		actual := leveled.GetLevel(e.module)
		if e.level != actual {
			t.Errorf("unexpected level in %s: %s != %s", e.module, e.level, actual)
		}
	}
}
