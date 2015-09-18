// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"strconv"
	"testing"
)

// TODO share more code between these tests
func MemoryRecordN(b *MemoryBackend, n int) *Record {
	node := b.Head()
	for i := 0; i < n; i++ {
		if node == nil {
			break
		}
		node = node.Next()
	}
	if node == nil {
		return nil
	}
	return node.Record
}

func ChannelMemoryRecordN(b *ChannelMemoryBackend, n int) *Record {
	b.Flush()
	node := b.Head()
	for i := 0; i < n; i++ {
		if node == nil {
			break
		}
		node = node.Next()
	}
	if node == nil {
		return nil
	}
	return node.Record
}

func TestMemoryBackend(t *testing.T) {
	backend := NewMemoryBackend(8)
	SetBackend(backend)

	log := MustGetLogger("test")

	if nil != MemoryRecordN(backend, 0) || 0 != backend.size {
		t.Errorf("memory level: %d", backend.size)
	}

	// Run 13 times, the resulting vector should be [5..12]
	for i := 0; i < 13; i++ {
		log.Info("%d", i)
	}

	if 8 != backend.size {
		t.Errorf("record length: %d", backend.size)
	}
	record := MemoryRecordN(backend, 0)
	if "5" != record.Formatted(0) {
		t.Errorf("unexpected start: %s", record.Formatted(0))
	}
	for i := 0; i < 8; i++ {
		record = MemoryRecordN(backend, i)
		if strconv.Itoa(i+5) != record.Formatted(0) {
			t.Errorf("unexpected record: %v", record.Formatted(0))
		}
	}
	record = MemoryRecordN(backend, 7)
	if "12" != record.Formatted(0) {
		t.Errorf("unexpected end: %s", record.Formatted(0))
	}
	record = MemoryRecordN(backend, 8)
	if nil != record {
		t.Errorf("unexpected eof: %s", record.Formatted(0))
	}
}

func TestChannelMemoryBackend(t *testing.T) {
	backend := NewChannelMemoryBackend(8)
	SetBackend(backend)

	log := MustGetLogger("test")

	if nil != ChannelMemoryRecordN(backend, 0) || 0 != backend.size {
		t.Errorf("memory level: %d", backend.size)
	}

	// Run 13 times, the resulting vector should be [5..12]
	for i := 0; i < 13; i++ {
		log.Info("%d", i)
	}
	backend.Flush()

	if 8 != backend.size {
		t.Errorf("record length: %d", backend.size)
	}
	record := ChannelMemoryRecordN(backend, 0)
	if "5" != record.Formatted(0) {
		t.Errorf("unexpected start: %s", record.Formatted(0))
	}
	for i := 0; i < 8; i++ {
		record = ChannelMemoryRecordN(backend, i)
		if strconv.Itoa(i+5) != record.Formatted(0) {
			t.Errorf("unexpected record: %v", record.Formatted(0))
		}
	}
	record = ChannelMemoryRecordN(backend, 7)
	if "12" != record.Formatted(0) {
		t.Errorf("unexpected end: %s", record.Formatted(0))
	}
	record = ChannelMemoryRecordN(backend, 8)
	if nil != record {
		t.Errorf("unexpected eof: %s", record.Formatted(0))
	}
}
