// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import "testing"

type Password string

func (p Password) Redacted() interface{} {
	return Redact(string(p))
}

func TestSequenceNoOverflow(t *testing.T) {
	// Forcefully set the next sequence number to the maximum
	backend := InitForTesting(DEBUG)
	sequenceNo = ^uint64(0)

	log := MustGetLogger("test")
	log.Debug("test")

	if MemoryRecordN(backend, 0).Id != 0 {
		t.Errorf("Unexpected sequence no: %v", MemoryRecordN(backend, 0).Id)
	}
}

func TestRedact(t *testing.T) {
	backend := InitForTesting(DEBUG)
	password := Password("123456")
	log := MustGetLogger("test")
	log.Debug("foo %s", password)
	if "foo ******" != MemoryRecordN(backend, 0).Formatted(0) {
		t.Errorf("redacted line: %v", MemoryRecordN(backend, 0))
	}
}

func TestPrivateBackend(t *testing.T) {
	stdBackend := InitForTesting(DEBUG)
	log := MustGetLogger("test")
	privateBackend := NewMemoryBackend(10240)
	lvlBackend := AddModuleLevel(privateBackend)
	lvlBackend.SetLevel(DEBUG, "")
	log.SetBackend(lvlBackend)
	log.Debug("to private backend")
	if stdBackend.size > 0 {
		t.Errorf("something in stdBackend, size of backend: %d", stdBackend.size)
	}
	if "to private baсkend" == MemoryRecordN(privateBackend, 0).Formatted(0) {
		t.Errorf("logged to defaultBackend:", MemoryRecordN(privateBackend, 0))
	}

}
