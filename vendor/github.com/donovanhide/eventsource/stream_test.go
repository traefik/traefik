package eventsource

import (
	"io"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const (
	eventChannelName   = "Test"
	timeToWaitForEvent = 100 * time.Millisecond
)

func TestStreamSubscribeEventsChan(t *testing.T) {
	server := NewServer()
	httpServer := httptest.NewServer(server.Handler(eventChannelName))
	// The server has to be closed before the httpServer is closed.
	// Otherwise the httpServer has still an open connection and it can not close.
	defer httpServer.Close()
	defer server.Close()

	stream := mustSubscribe(t, httpServer.URL, "")

	publishedEvent := &publication{id: "123"}
	server.Publish([]string{eventChannelName}, publishedEvent)

	select {
	case receivedEvent := <-stream.Events:
		if !reflect.DeepEqual(receivedEvent, publishedEvent) {
			t.Errorf("got event %+v, want %+v", receivedEvent, publishedEvent)
		}
	case <-time.After(timeToWaitForEvent):
		t.Error("Timed out waiting for event")
	}
}

func TestStreamSubscribeErrorsChan(t *testing.T) {
	server := NewServer()
	httpServer := httptest.NewServer(server.Handler(eventChannelName))

	defer httpServer.Close()

	stream := mustSubscribe(t, httpServer.URL, "")
	server.Close()

	select {
	case err := <-stream.Errors:
		if err != io.EOF {
			t.Errorf("got error %+v, want %+v", err, io.EOF)
		}
	case <-time.After(timeToWaitForEvent):
		t.Error("Timed out waiting for error event")
	}
}

func TestStreamClose(t *testing.T) {
	server := NewServer()
	httpServer := httptest.NewServer(server.Handler(eventChannelName))
	// The server has to be closed before the httpServer is closed.
	// Otherwise the httpServer has still an open connection and it can not close.
	defer httpServer.Close()
	defer server.Close()

	stream := mustSubscribe(t, httpServer.URL, "")
	stream.Close()
	// its safe to Close the stream multiple times
	stream.Close()

	select {
	case _, ok := <-stream.Events:
		if ok {
			t.Error("Expected stream.Events channel to be closed. Is still open.")
		}
	case <-time.After(timeToWaitForEvent):
		t.Error("Timed out waiting for stream.Events channel to close")
	}

	select {
	case _, ok := <-stream.Errors:
		if ok {
			t.Error("Expected stream.Errors channel to be closed. Is still open.")
		}
	case <-time.After(timeToWaitForEvent):
		t.Error("Timed out waiting for stream.Errors channel to close")
	}
}

func mustSubscribe(t *testing.T, url, lastEventId string) *Stream {
	stream, err := Subscribe(url, lastEventId)
	if err != nil {
		t.Fatalf("Failed to subscribe: %s", err)
	}
	return stream
}
