package safe

import (
	"sync/atomic"
)

// Safe contains a thread-safe value.
type Safe struct {
	atomic.Value
}

// New create a new Safe instance given a value.
func New(value interface{}) *Safe {
	return &Safe{atomic.Value{}}
}

// Get returns the value.
func (s *Safe) Get() interface{} {
	return s.Load()
}

// Set sets a new value.
func (s *Safe) Set(value interface{}) {
	s.Store(value)
}
