// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import "testing"

func TestMultiLogger(t *testing.T) {
	log1 := NewMemoryBackend(8)
	log2 := NewMemoryBackend(8)
	SetBackend(MultiLogger(log1, log2))

	log := MustGetLogger("test")
	log.Debug("log")

	if "log" != MemoryRecordN(log1, 0).Formatted(0) {
		t.Errorf("log1: %v", MemoryRecordN(log1, 0).Formatted(0))
	}
	if "log" != MemoryRecordN(log2, 0).Formatted(0) {
		t.Errorf("log2: %v", MemoryRecordN(log2, 0).Formatted(0))
	}
}

func TestMultiLoggerLevel(t *testing.T) {
	log1 := NewMemoryBackend(8)
	log2 := NewMemoryBackend(8)

	leveled1 := AddModuleLevel(log1)
	leveled2 := AddModuleLevel(log2)

	multi := MultiLogger(leveled1, leveled2)
	multi.SetLevel(ERROR, "test")
	SetBackend(multi)

	log := MustGetLogger("test")
	log.Notice("log")

	if nil != MemoryRecordN(log1, 0) || nil != MemoryRecordN(log2, 0) {
		t.Errorf("unexpected log record")
	}

	leveled1.SetLevel(DEBUG, "test")
	log.Notice("log")
	if "log" != MemoryRecordN(log1, 0).Formatted(0) {
		t.Errorf("log1 not received")
	}
	if nil != MemoryRecordN(log2, 0) {
		t.Errorf("log2 received")
	}
}
