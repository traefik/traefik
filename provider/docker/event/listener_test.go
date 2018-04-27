package event_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/containous/traefik/provider/docker/event"
	"github.com/docker/docker/api/types/events"
)

func TestTickerCallback(t *testing.T) {
	callbackFunc := func(msg events.Message) {
		expected := events.Message{}
		if !reflect.DeepEqual(expected, msg) {
			t.Fatal("expected", expected, "got", msg)
		}
	}

	stopChan := make(chan bool)
	e := &event.Ticker{
		CallbackFunc:   callbackFunc,
		StopChan:       stopChan,
		TickerInterval: 1 * time.Second,
	}

	go func(stopChan chan bool) {
		time.Sleep(2 * time.Second)

		stopChan <- true
	}(stopChan)

	e.Start()
}

func TestStreamerCallback(t *testing.T) {
	eventsMsgChan := make(chan events.Message)
	eventsErrChan := make(chan error)

	callbackFunc := func(msg events.Message) {
		expected := events.Message{
			Action: "update",
		}

		if !reflect.DeepEqual(expected, msg) {
			t.Fatal("expected", expected, "got", msg)
		}
	}

	stopChan := make(chan bool)
	errChan := make(chan error)
	e := event.Streamer{
		EventsMsgChan: eventsMsgChan,
		EventsErrChan: eventsErrChan,
		CallbackFunc:  callbackFunc,
		StopChan:      stopChan,
		ErrChan:       errChan,
		EventsCtx:     context.Background(),
	}

	go func(eventsMsgChan chan events.Message, stopChan chan bool) {
		time.Sleep(1 * time.Second)
		msg := events.Message{
			Action: "update",
		}
		eventsMsgChan <- msg

		time.Sleep(1 * time.Second)

		stopChan <- true
	}(eventsMsgChan, stopChan)

	e.Start()
}
