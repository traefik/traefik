// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !appengine

package logging

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// TODO pick one of the memory backends and stick with it or share interface.

// InitForTesting is a convenient method when using logging in a test. Once
// called, the time will be frozen to January 1, 1970 UTC.
func InitForTesting(level Level) *MemoryBackend {
	Reset()

	memoryBackend := NewMemoryBackend(10240)

	leveledBackend := AddModuleLevel(memoryBackend)
	leveledBackend.SetLevel(level, "")
	SetBackend(leveledBackend)

	timeNow = func() time.Time {
		return time.Unix(0, 0).UTC()
	}
	return memoryBackend
}

// Node is a record node pointing to an optional next node.
type node struct {
	next   *node
	Record *Record
}

// Next returns the next record node. If there's no node available, it will
// return nil.
func (n *node) Next() *node {
	return n.next
}

// MemoryBackend is a simple memory based logging backend that will not produce
// any output but merly keep records, up to the given size, in memory.
type MemoryBackend struct {
	size       int32
	maxSize    int32
	head, tail unsafe.Pointer
}

// NewMemoryBackend creates a simple in-memory logging backend.
func NewMemoryBackend(size int) *MemoryBackend {
	return &MemoryBackend{maxSize: int32(size)}
}

// Log implements the Log method required by Backend.
func (b *MemoryBackend) Log(level Level, calldepth int, rec *Record) error {
	var size int32

	n := &node{Record: rec}
	np := unsafe.Pointer(n)

	// Add the record to the tail. If there's no records available, tail and
	// head will both be nil. When we successfully set the tail and the previous
	// value was nil, it's safe to set the head to the current value too.
	for {
		tailp := b.tail
		swapped := atomic.CompareAndSwapPointer(
			&b.tail,
			tailp,
			np,
		)
		if swapped == true {
			if tailp == nil {
				b.head = np
			} else {
				(*node)(tailp).next = n
			}
			size = atomic.AddInt32(&b.size, 1)
			break
		}
	}

	// Since one record was added, we might have overflowed the list. Remove
	// a record if that is the case. The size will fluctate a bit, but
	// eventual consistent.
	if b.maxSize > 0 && size > b.maxSize {
		for {
			headp := b.head
			head := (*node)(b.head)
			if head.next == nil {
				break
			}
			swapped := atomic.CompareAndSwapPointer(
				&b.head,
				headp,
				unsafe.Pointer(head.next),
			)
			if swapped == true {
				atomic.AddInt32(&b.size, -1)
				break
			}
		}
	}
	return nil
}

// Head returns the oldest record node kept in memory. It can be used to
// iterate over records, one by one, up to the last record.
//
// Note: new records can get added while iterating. Hence the number of records
// iterated over might be larger than the maximum size.
func (b *MemoryBackend) Head() *node {
	return (*node)(b.head)
}

type event int

const (
	eventFlush event = iota
	eventStop
)

// ChannelMemoryBackend is very similar to the MemoryBackend, except that it
// internally utilizes a channel.
type ChannelMemoryBackend struct {
	maxSize    int
	size       int
	incoming   chan *Record
	events     chan event
	mu         sync.Mutex
	running    bool
	flushWg    sync.WaitGroup
	stopWg     sync.WaitGroup
	head, tail *node
}

// NewChannelMemoryBackend creates a simple in-memory logging backend which
// utilizes a go channel for communication.
//
// Start will automatically be called by this function.
func NewChannelMemoryBackend(size int) *ChannelMemoryBackend {
	backend := &ChannelMemoryBackend{
		maxSize:  size,
		incoming: make(chan *Record, 1024),
		events:   make(chan event),
	}
	backend.Start()
	return backend
}

// Start launches the internal goroutine which starts processing data from the
// input channel.
func (b *ChannelMemoryBackend) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Launch the goroutine unless it's already running.
	if b.running != true {
		b.running = true
		b.stopWg.Add(1)
		go b.process()
	}
}

func (b *ChannelMemoryBackend) process() {
	defer b.stopWg.Done()
	for {
		select {
		case rec := <-b.incoming:
			b.insertRecord(rec)
		case e := <-b.events:
			switch e {
			case eventStop:
				return
			case eventFlush:
				for len(b.incoming) > 0 {
					b.insertRecord(<-b.incoming)
				}
				b.flushWg.Done()
			}
		}
	}
}

func (b *ChannelMemoryBackend) insertRecord(rec *Record) {
	prev := b.tail
	b.tail = &node{Record: rec}
	if prev == nil {
		b.head = b.tail
	} else {
		prev.next = b.tail
	}

	if b.maxSize > 0 && b.size >= b.maxSize {
		b.head = b.head.next
	} else {
		b.size += 1
	}
}

// Flush waits until all records in the buffered channel have been processed.
func (b *ChannelMemoryBackend) Flush() {
	b.flushWg.Add(1)
	b.events <- eventFlush
	b.flushWg.Wait()
}

// Stop signals the internal goroutine to exit and waits until it have.
func (b *ChannelMemoryBackend) Stop() {
	b.mu.Lock()
	if b.running == true {
		b.running = false
		b.events <- eventStop
	}
	b.mu.Unlock()
	b.stopWg.Wait()
}

// Log implements the Log method required by Backend.
func (b *ChannelMemoryBackend) Log(level Level, calldepth int, rec *Record) error {
	b.incoming <- rec
	return nil
}

// Head returns the oldest record node kept in memory. It can be used to
// iterate over records, one by one, up to the last record.
//
// Note: new records can get added while iterating. Hence the number of records
// iterated over might be larger than the maximum size.
func (b *ChannelMemoryBackend) Head() *node {
	return b.head
}
