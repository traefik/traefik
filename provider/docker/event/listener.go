package event

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

// NewListener creates a new event listener depending on what the user's Docker daemon supports.
func NewListener(dockerClient client.APIClient, dockerEventsOptions types.EventsOptions, stopChan chan bool, errChan chan error, callback Callback) (Listener, error) {
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
			Callback:      callback,
			StopChan:      stopChan,
			ErrChan:       errChan,
			EventsCtx:     eventsCtx,
		}

		return e, nil
	}

	// Fallback to the ticker.
	e := &Ticker{
		Callback:       callback,
		StopChan:       stopChan,
		TickerInterval: SwarmDefaultWatchTime,
	}

	return e, nil
}

// Listener is an interface for event listeners to implement.
type Listener interface {
	Start()
	Stop()
}

// Ticker is a fake event listener, that instead of listening for real events, the callback function is executed when the ticker ticks.
type Ticker struct {
	Callback       Callback
	StopChan       chan bool
	TickerInterval time.Duration
	ticker         *time.Ticker
}

// Streamer is a real time Docker event listener that executes the callback function with the event as a payload.
type Streamer struct {
	EventsMsgChan <-chan events.Message
	EventsErrChan <-chan error
	Callback      Callback
	StopChan      chan bool
	ErrChan       chan error
	EventsCtx     context.Context
}

// Start starts up the ticker.
func (e *Ticker) Start() {
	log.Debug("Docker events listener: Ticker started!")

	e.ticker = time.NewTicker(e.TickerInterval)

	for {
		select {
		case <-e.ticker.C:
			go e.Callback.Execute(events.Message{})
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
	log.Debug("Docker events listener: Streamer started!")

	for {
		select {
		case evt := <-e.EventsMsgChan:
			log.Debugf("Docker events listener, incoming event: %#v", evt)
			go e.Callback.Execute(evt)
		case evtErr := <-e.EventsErrChan:
			log.Errorf("Docker events listener: Events error, %s", evtErr.Error())

			e.ErrChan <- evtErr
			e.Stop()

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
