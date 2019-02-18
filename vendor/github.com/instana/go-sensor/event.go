package instana

import (
	"time"
)

// EventData is the construct serialized for the host agent
type EventData struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	// Duration in milliseconds
	Duration int `json:"duration"`
	// Severity with value of -1, 5, 10 : see type severity
	Severity int    `json:"severity"`
	Plugin   string `json:"plugin,omitempty"`
	ID       string `json:"id,omitempty"`
	Host     string `json:"host"`
}

type severity int

//Severity values for events sent to the instana agent
const (
	SeverityChange   severity = -1
	SeverityWarning  severity = 5
	SeverityCritical severity = 10
)

// Defaults for the Event API
const (
	ServicePlugin = "com.instana.forge.connection.http.logical.LogicalWebApp"
	ServiceHost   = ""
)

// SendDefaultServiceEvent sends a default event which already contains the service and host
func SendDefaultServiceEvent(title string, text string, sev severity, duration time.Duration) {
	if sensor == nil {
		// Since no sensor was initialized, there is no default service (as
		// configured on the sensor) so we send blank.
		SendServiceEvent("", title, text, sev, duration)
	} else {
		SendServiceEvent(sensor.serviceName, title, text, sev, duration)
	}
}

// SendServiceEvent send an event on a specific service
func SendServiceEvent(service string, title string, text string, sev severity, duration time.Duration) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Severity: int(sev),
		Plugin:   ServicePlugin,
		ID:       service,
		Host:     ServiceHost,
		Duration: int(duration / time.Millisecond),
	})
}

// SendHostEvent send an event on the current host
func SendHostEvent(title string, text string, sev severity, duration time.Duration) {
	sendEvent(&EventData{
		Title:    title,
		Text:     text,
		Duration: int(duration / time.Millisecond),
		Severity: int(sev),
	})
}

func sendEvent(event *EventData) {
	if sensor == nil {
		// If the sensor hasn't initialized we do so here so that we properly
		// discover where the host agent may be as it varies between a
		// normal host, docker, kubernetes etc..
		InitSensor(&Options{})
	}
	//we do fire & forget here, because the whole pid dance isn't necessary to send events
	go sensor.agent.request(sensor.agent.makeURL(agentEventURL), "POST", event)
}
