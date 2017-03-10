// Package eventsource implements a client and server to allow streaming data one-way over a HTTP connection
// using the Server-Sent Events API http://dev.w3.org/html5/eventsource/
//
// The client and server respect the Last-Event-ID header.
// If the Repository interface is implemented on the server, events can be replayed in case of a network disconnection.
package eventsource

// Any event received by the client or sent by the server will implement this interface
type Event interface {
	// Id is an identifier that can be used to allow a client to replay
	// missed Events by returning the Last-Event-Id header.
	// Return empty string if not required.
	Id() string
	// The name of the event. Return empty string if not required.
	Event() string
	// The payload of the event.
	Data() string
}

// If history is required, this interface will allow clients to reply previous events through the server.
// Both methods can be called from different goroutines concurrently, so you must make sure they are go-routine safe.
type Repository interface {
	// Gets the Events which should follow on from the specified channel and event id.
	Replay(channel, id string) chan Event
}
