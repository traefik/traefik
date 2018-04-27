package event_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/containous/traefik/provider/docker/event"
	"github.com/docker/docker/api/types/events"
)

type callbackTest struct {
	t        *testing.T
	expected events.Message
}

func (c *callbackTest) Execute(msg events.Message) {
	if !reflect.DeepEqual(c.expected, msg) {
		c.t.Fatal("expected", c.expected, "got", msg)
	}
}

func TestTickerCallback(t *testing.T) {
	callback := &callbackTest{
		t:        t,
		expected: events.Message{},
	}

	stopChan := make(chan bool)
	e := &event.Ticker{
		Callback:       callback,
		StopChan:       stopChan,
		TickerInterval: 1 * time.Second,
	}

	go func(stopChan chan bool) {
		time.Sleep(2 * time.Second)

		stopChan <- true
	}(stopChan)

	e.Start()
}

func TestStreamerCallbackSuccessfulEvent(t *testing.T) {
	eventsMsgChan := make(chan events.Message)
	eventsErrChan := make(chan error)

	callback := &callbackTest{
		t: t,
		expected: events.Message{
			Action: "update",
		},
	}

	stopChan := make(chan bool)
	errChan := make(chan error)
	e := event.Streamer{
		EventsMsgChan: eventsMsgChan,
		EventsErrChan: eventsErrChan,
		Callback:      callback,
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

func TestStreamerCallbackErrorEvent(t *testing.T) {
	eventsMsgChan := make(chan events.Message)
	eventsErrChan := make(chan error)

	callback := &callbackTest{}

	stopChan := make(chan bool)
	errChan := make(chan error)
	e := event.Streamer{
		EventsMsgChan: eventsMsgChan,
		EventsErrChan: eventsErrChan,
		Callback:      callback,
		StopChan:      stopChan,
		ErrChan:       errChan,
		EventsCtx:     context.Background(),
	}

	go func(eventsErrChan chan error) {
		time.Sleep(1 * time.Second)

		eventsErrChan <- fmt.Errorf("oh my")
	}(eventsErrChan)

	go e.Start()

	expectedErrorStr := "oh my"
	eventsError := <-errChan

	if expectedErrorStr != eventsError.Error() {
		t.Fatal("expected", expectedErrorStr, "got", eventsError.Error())
	}
}
