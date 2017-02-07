package safe

import (
	"sync"
)

// Safe contains a thread-safe value
type Safe struct {
	value interface{}
	lock  sync.RWMutex
}

// New create a new Safe instance given a value
func New(value interface{}) *Safe {
	return &Safe{value: value, lock: sync.RWMutex{}}
}

// Get returns the value
func (s *Safe) Get() interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.value
}

// Set sets a new value
func (s *Safe) Set(value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.value = value
}
