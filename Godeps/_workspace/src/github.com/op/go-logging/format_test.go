// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"bytes"
	"testing"
)

func TestFormat(t *testing.T) {
	backend := InitForTesting(DEBUG)

	f, err := NewStringFormatter("%{shortfile} %{time:2006-01-02T15:04:05} %{level:.1s} %{id:04d} %{module} %{message}")
	if err != nil {
		t.Fatalf("failed to set format: %s", err)
	}
	SetFormatter(f)

	log := MustGetLogger("module")
	log.Debug("hello")

	line := MemoryRecordN(backend, 0).Formatted(0)
	if "format_test.go:24 1970-01-01T00:00:00 D 0001 module hello" != line {
		t.Errorf("Unexpected format: %s", line)
	}
}

func logAndGetLine(backend *MemoryBackend) string {
	MustGetLogger("foo").Debug("hello")
	return MemoryRecordN(backend, 0).Formatted(1)
}

func getLastLine(backend *MemoryBackend) string {
	return MemoryRecordN(backend, 0).Formatted(1)
}

func realFunc(backend *MemoryBackend) string {
	return logAndGetLine(backend)
}

type structFunc struct{}

func (structFunc) Log(backend *MemoryBackend) string {
	return logAndGetLine(backend)
}

func TestRealFuncFormat(t *testing.T) {
	backend := InitForTesting(DEBUG)
	SetFormatter(MustStringFormatter("%{shortfunc}"))

	line := realFunc(backend)
	if "realFunc" != line {
		t.Errorf("Unexpected format: %s", line)
	}
}

func TestStructFuncFormat(t *testing.T) {
	backend := InitForTesting(DEBUG)
	SetFormatter(MustStringFormatter("%{longfunc}"))

	var x structFunc
	line := x.Log(backend)
	if "structFunc.Log" != line {
		t.Errorf("Unexpected format: %s", line)
	}
}

func TestVarFuncFormat(t *testing.T) {
	backend := InitForTesting(DEBUG)
	SetFormatter(MustStringFormatter("%{shortfunc}"))

	var varFunc = func() string {
		return logAndGetLine(backend)
	}

	line := varFunc()
	if "???" == line || "TestVarFuncFormat" == line || "varFunc" == line {
		t.Errorf("Unexpected format: %s", line)
	}
}

func TestFormatFuncName(t *testing.T) {
	var tests = []struct {
		filename  string
		longpkg   string
		shortpkg  string
		longfunc  string
		shortfunc string
	}{
		{"",
			"???",
			"???",
			"???",
			"???"},
		{"main",
			"???",
			"???",
			"???",
			"???"},
		{"main.",
			"main",
			"main",
			"",
			""},
		{"main.main",
			"main",
			"main",
			"main",
			"main"},
		{"github.com/op/go-logging.func·001",
			"github.com/op/go-logging",
			"go-logging",
			"func·001",
			"func·001"},
		{"github.com/op/go-logging.stringFormatter.Format",
			"github.com/op/go-logging",
			"go-logging",
			"stringFormatter.Format",
			"Format"},
	}

	var v string
	for _, test := range tests {
		v = formatFuncName(fmtVerbLongpkg, test.filename)
		if test.longpkg != v {
			t.Errorf("%s != %s", test.longpkg, v)
		}
		v = formatFuncName(fmtVerbShortpkg, test.filename)
		if test.shortpkg != v {
			t.Errorf("%s != %s", test.shortpkg, v)
		}
		v = formatFuncName(fmtVerbLongfunc, test.filename)
		if test.longfunc != v {
			t.Errorf("%s != %s", test.longfunc, v)
		}
		v = formatFuncName(fmtVerbShortfunc, test.filename)
		if test.shortfunc != v {
			t.Errorf("%s != %s", test.shortfunc, v)
		}
	}
}

func TestBackendFormatter(t *testing.T) {
	InitForTesting(DEBUG)

	// Create two backends and wrap one of the with a backend formatter
	b1 := NewMemoryBackend(1)
	b2 := NewMemoryBackend(1)

	f := MustStringFormatter("%{level} %{message}")
	bf := NewBackendFormatter(b2, f)

	SetBackend(b1, bf)

	log := MustGetLogger("module")
	log.Info("foo")
	if "foo" != getLastLine(b1) {
		t.Errorf("Unexpected line: %s", getLastLine(b1))
	}
	if "INFO foo" != getLastLine(b2) {
		t.Errorf("Unexpected line: %s", getLastLine(b2))
	}
}

func BenchmarkStringFormatter(b *testing.B) {
	fmt := "%{time:2006-01-02T15:04:05} %{level:.1s} %{id:04d} %{module} %{message}"
	f := MustStringFormatter(fmt)

	backend := InitForTesting(DEBUG)
	buf := &bytes.Buffer{}
	log := MustGetLogger("module")
	log.Debug("")
	record := MemoryRecordN(backend, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := f.Format(1, record, buf); err != nil {
			b.Fatal(err)
			buf.Truncate(0)
		}
	}
}
