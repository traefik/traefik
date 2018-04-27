package docker

import (
	"context"
	"time"

	"github.com/containous/traefik/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/docker/client"
)

const (
	// SwarmDefaultWatchTime is the duration of the interval when polling docker
	SwarmDefaultWatchTime = 15 * time.Second
)

// NewEvent creates a new event listener depending on what the user's Docker daemon supports.
func NewEvent(dockerClient client.APIClient, dockerEventsOptions types.EventsOptions, stopChan chan bool, errChan chan error, callbackFunc func(events.Message)) (Event, error) {
	serverVersion, err := dockerClient.ServerVersion(context.Background())
	if err != nil {
		return nil, err
	}

	// https://docs.docker.com/engine/api/v1.29/#tag/Network (Docker 17.06)
	if versions.GreaterThanOrEqualTo(serverVersion.APIVersion, "1.29") {
		eventsCtx := context.Background()
		eventsMsgChan, eventsErrChan := dockerClient.Events(
			eventsCtx,
			dockerEventsOptions,
		)

		e := &Streamer{
			EventsMsgChan: eventsMsgChan,
			EventsErrChan: eventsErrChan,
			CallbackFunc:  callbackFunc,
			StopChan:      stopChan,
			ErrChan:       errChan,
			EventsCtx:     eventsCtx,
		}

		return e, nil
	}

	// Fallback to the ticker.
	e := &Ticker{
		CallbackFunc:   callbackFunc,
		StopChan:       stopChan,
		TickerInterval: SwarmDefaultWatchTime,
	}

	return e, nil
}

// Event is an interface for event listeners to implement.
type Event interface {
	Start()
	Stop()
}

// Ticker is a fake event listener, that instead of listening for real events, the callback function is executed when the ticker ticks.
type Ticker struct {
	CallbackFunc   func(events.Message)
	StopChan       chan bool
	TickerInterval time.Duration
	ticker         *time.Ticker
}

// Streamer is a real time Docker event listener that executes the callback function with the event as a payload.
type Streamer struct {
	EventsMsgChan <-chan events.Message
	EventsErrChan <-chan error
	CallbackFunc  func(events.Message)
	StopChan      chan bool
	ErrChan       chan error
	EventsCtx     context.Context
}

// Start starts up the ticker.
func (e *Ticker) Start() {
	log.Debug("Docker events handler: Ticker started!")

	e.ticker = time.NewTicker(e.TickerInterval)

	for {
		select {
		case <-e.ticker.C:
			go e.CallbackFunc(events.Message{})
		case <-e.StopChan:
			e.Stop()

			return
		}
	}
}

// Stop stops the ticker.
func (e *Ticker) Stop() {
	e.ticker.Stop()
}

// Start starts up the real time Docker event listener.
func (e *Streamer) Start() {
	log.Debug("Docker events handler: Streamer started!")

	for {
		select {
		case evt := <-e.EventsMsgChan:
			go e.CallbackFunc(evt)
		case evtErr := <-e.EventsErrChan:
			log.Errorf("Docker events listener: Events error, %s", evtErr.Error())

			e.Stop()
			e.ErrChan <- evtErr

			return
		case <-e.StopChan:
			e.Stop()

			return
		}
	}
}

// Stop stops the real time Docker event listener.
func (e *Streamer) Stop() {
	e.EventsCtx.Done()
}
