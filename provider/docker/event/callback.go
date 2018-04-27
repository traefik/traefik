package event

import (
	"github.com/docker/docker/api/types/events"
)

// Callback defines the interface all event callbacks need to implement.
type Callback interface {
	Execute(events.Message)
}
