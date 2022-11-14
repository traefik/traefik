package safe

import (
	"sync"
	"sync/atomic"
)

// Sync synchronizes a thread-safe value, meaning, once Set() has returned, subsequent
// calls to Get() will retrieve the updated value.
type Sync[T any] struct {
	value T
	lock  sync.RWMutex
}

// New create a new Sync instance given a value.
func NewSync[T any](value T) *Sync[T] {
	return &Sync[T]{value: value, lock: sync.RWMutex{}}
}

// Get returns the value.
func (s *Sync[T]) Get() T {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.value
}

// Set sets a new value.
func (s *Sync[T]) Set(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.value = value
}

// Atomic contains a thread-safe value however not synchronized. This means that
// calls to Get() can retrieve the old value for a short amount of time after Set()
// had been called. This trade-off allows much more rapid Get() than with a Sync
// which uses a sync.RWMutex instead of an atomic.Pointer.
type Atomic[T any] struct {
	atomic.Pointer[T]
}

// New create a new Atomic instance given a value.
func NewAtomic[T any](value T) *Atomic[T] {
	p := Atomic[T]{atomic.Pointer[T]{}}
	p.Store(&value)
	return &p
}

// Get returns the value.
func (s *Atomic[T]) Get() T {
	return *s.Load()
}

// Set sets a new value.
func (s *Atomic[T]) Set(value T) {
	s.Store(&value)
}
