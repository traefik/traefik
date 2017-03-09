package events

import (
	"sync"

	eventtypes "github.com/docker/engine-api/types/events"
)

// NewHandler creates an event handler using the specified function to qualify the message
// and to route it to the correct handler.
func NewHandler(fun func(eventtypes.Message) string) *Handler {
	return &Handler{
		keyFunc:  fun,
		handlers: make(map[string]func(eventtypes.Message)),
	}
}

// ByType is a qualify function based on message type.
func ByType(e eventtypes.Message) string {
	return e.Type
}

// ByAction is a qualify function based on message action.
func ByAction(e eventtypes.Message) string {
	return e.Action
}

// Handler is a struct holding the handlers by keys, and the function to get the
// key from the message.
type Handler struct {
	keyFunc  func(eventtypes.Message) string
	handlers map[string]func(eventtypes.Message)
	mu       sync.Mutex
}

// Handle registers a function has handler for the specified key.
func (w *Handler) Handle(key string, h func(eventtypes.Message)) {
	w.mu.Lock()
	w.handlers[key] = h
	w.mu.Unlock()
}

// Watch ranges over the passed in event chan and processes the events based on the
// handlers created for a given action.
// To stop watching, close the event chan.
func (w *Handler) Watch(c <-chan eventtypes.Message) {
	for e := range c {
		w.mu.Lock()
		h, exists := w.handlers[w.keyFunc(e)]
		w.mu.Unlock()
		if !exists {
			continue
		}
		go h(e)
	}
}
