// Package webhook provides the support for reading and parsing the events
// sent from DNSimple via webhook.
package webhook

import (
	"encoding/json"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

// Actor represents the entity that triggered the event. It can be either an user,
// a DNSimple support representative or the DNSimple system.
type Actor struct {
	ID     string `json:"id"`
	Entity string `json:"entity"`
	Pretty string `json:"pretty"`
}

// Actor represents the account that this event is attached to.
type Account struct {
	dnsimple.Account

	// Display is a string that can be used as a display label
	// and it is sent in a webhook payload.
	// It generally represent the Name of the account.
	Display string `json:"display,omitempty"`

	// Identifier is a human-readable string identifier
	// and it is sent in a webhook payload
	// It generally represent the StringID or email of the account.
	Identifier string `json:"identifier,omitempty"`
}

// Event is an event generated in the DNSimple application.
type Event interface {
	EventName() string
	EventHeader() *Event_Header
	Payload() []byte
	parse([]byte) error
}

type Event_Header struct {
	APIVersion string   `json:"api_version"`
	RequestID  string   `json:"request_identifier"`
	Actor      *Actor   `json:"actor"`
	Account    *Account `json:"account"`
	Name       string   `json:"name"`
	Auto       bool     `json:"auto"`
	payload    []byte
}

type eventName struct {
	Name string `json:"name"`
}

// Event returns the event name as defined in the name field of the payload.
func (e *Event_Header) EventHeader() *Event_Header {
	return e
}

// EventName returns the event name as defined in the name field of the payload.
func (e *Event_Header) EventName() string {
	return e.Name
}

// Payload returns the binary payload the event was deserialized from.
func (e *Event_Header) Payload() []byte {
	return e.payload
}

func (e *Event_Header) parse(payload []byte) error {
	e.payload = payload
	return unmashalEvent(payload, e)
}

// Parse takes a payload and attempts to deserialize the payload into an event type
// that matches the event action in the payload. If no direct match is found, then a DefaultEvent is returned.
//
// Parse returns type is an Event interface. Therefore, you must perform typecasting
// to access any event-specific field.
func Parse(payload []byte) (Event, error) {
	action, err := ParseName(payload)
	if err != nil {
		return nil, err
	}

	return switchEvent(action, payload)
}

func ParseName(data []byte) (string, error) {
	eventName := &eventName{}
	err := json.Unmarshal(data, eventName)
	return eventName.Name, err
}

func unmashalEvent(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
