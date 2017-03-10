// Package events holds event structures, methods and functions.
package events

import (
	"fmt"
	"time"
)

// Notifier defines the methods an event notifier should have.
type Notifier interface {
	Notify(eventType EventType, serviceName string, data map[string]string)
}

// Emitter defines the methods an event emitter should have.
type Emitter interface {
	AddListener(c chan<- Event)
}

// Event holds project-wide event informations.
type Event struct {
	EventType   EventType
	ServiceName string
	Data        map[string]string
}

// ContainerEvent holds attributes of container events.
type ContainerEvent struct {
	Service    string            `json:"service"`
	Event      string            `json:"event"`
	ID         string            `json:"id"`
	Time       time.Time         `json:"time"`
	Attributes map[string]string `json:"attributes"`
	Type       string            `json:"type"`
}

// EventType defines a type of libcompose event.
type EventType int

// Definitions of libcompose events
const (
	NoEvent = EventType(iota)

	ContainerCreated = EventType(iota)
	ContainerStarted = EventType(iota)

	ServiceAdd          = EventType(iota)
	ServiceUpStart      = EventType(iota)
	ServiceUpIgnored    = EventType(iota)
	ServiceUp           = EventType(iota)
	ServiceCreateStart  = EventType(iota)
	ServiceCreate       = EventType(iota)
	ServiceDeleteStart  = EventType(iota)
	ServiceDelete       = EventType(iota)
	ServiceDownStart    = EventType(iota)
	ServiceDown         = EventType(iota)
	ServiceRestartStart = EventType(iota)
	ServiceRestart      = EventType(iota)
	ServicePullStart    = EventType(iota)
	ServicePull         = EventType(iota)
	ServiceKillStart    = EventType(iota)
	ServiceKill         = EventType(iota)
	ServiceStartStart   = EventType(iota)
	ServiceStart        = EventType(iota)
	ServiceBuildStart   = EventType(iota)
	ServiceBuild        = EventType(iota)
	ServicePauseStart   = EventType(iota)
	ServicePause        = EventType(iota)
	ServiceUnpauseStart = EventType(iota)
	ServiceUnpause      = EventType(iota)
	ServiceStopStart    = EventType(iota)
	ServiceStop         = EventType(iota)
	ServiceRunStart     = EventType(iota)
	ServiceRun          = EventType(iota)

	VolumeAdd  = EventType(iota)
	NetworkAdd = EventType(iota)

	ProjectDownStart     = EventType(iota)
	ProjectDownDone      = EventType(iota)
	ProjectCreateStart   = EventType(iota)
	ProjectCreateDone    = EventType(iota)
	ProjectUpStart       = EventType(iota)
	ProjectUpDone        = EventType(iota)
	ProjectDeleteStart   = EventType(iota)
	ProjectDeleteDone    = EventType(iota)
	ProjectRestartStart  = EventType(iota)
	ProjectRestartDone   = EventType(iota)
	ProjectReload        = EventType(iota)
	ProjectReloadTrigger = EventType(iota)
	ProjectKillStart     = EventType(iota)
	ProjectKillDone      = EventType(iota)
	ProjectStartStart    = EventType(iota)
	ProjectStartDone     = EventType(iota)
	ProjectBuildStart    = EventType(iota)
	ProjectBuildDone     = EventType(iota)
	ProjectPauseStart    = EventType(iota)
	ProjectPauseDone     = EventType(iota)
	ProjectUnpauseStart  = EventType(iota)
	ProjectUnpauseDone   = EventType(iota)
	ProjectStopStart     = EventType(iota)
	ProjectStopDone      = EventType(iota)
)

func (e EventType) String() string {
	var m string
	switch e {
	case ContainerCreated:
		m = "Created container"
	case ContainerStarted:
		m = "Started container"

	case ServiceAdd:
		m = "Adding"
	case ServiceUpStart:
		m = "Starting"
	case ServiceUpIgnored:
		m = "Ignoring"
	case ServiceUp:
		m = "Started"
	case ServiceCreateStart:
		m = "Creating"
	case ServiceCreate:
		m = "Created"
	case ServiceDeleteStart:
		m = "Deleting"
	case ServiceDelete:
		m = "Deleted"
	case ServiceStopStart:
		m = "Stopping"
	case ServiceStop:
		m = "Stopped"
	case ServiceDownStart:
		m = "Stopping"
	case ServiceDown:
		m = "Stopped"
	case ServiceRestartStart:
		m = "Restarting"
	case ServiceRestart:
		m = "Restarted"
	case ServicePullStart:
		m = "Pulling"
	case ServicePull:
		m = "Pulled"
	case ServiceKillStart:
		m = "Killing"
	case ServiceKill:
		m = "Killed"
	case ServiceStartStart:
		m = "Starting"
	case ServiceStart:
		m = "Started"
	case ServiceBuildStart:
		m = "Building"
	case ServiceBuild:
		m = "Built"
	case ServiceRunStart:
		m = "Executing"
	case ServiceRun:
		m = "Executed"
	case ServicePauseStart:
		m = "Pausing"
	case ServicePause:
		m = "Paused"
	case ServiceUnpauseStart:
		m = "Unpausing"
	case ServiceUnpause:
		m = "Unpaused"

	case ProjectDownStart:
		m = "Stopping project"
	case ProjectDownDone:
		m = "Project stopped"
	case ProjectStopStart:
		m = "Stopping project"
	case ProjectStopDone:
		m = "Project stopped"
	case ProjectCreateStart:
		m = "Creating project"
	case ProjectCreateDone:
		m = "Project created"
	case ProjectUpStart:
		m = "Starting project"
	case ProjectUpDone:
		m = "Project started"
	case ProjectDeleteStart:
		m = "Deleting project"
	case ProjectDeleteDone:
		m = "Project deleted"
	case ProjectRestartStart:
		m = "Restarting project"
	case ProjectRestartDone:
		m = "Project restarted"
	case ProjectReload:
		m = "Reloading project"
	case ProjectReloadTrigger:
		m = "Triggering project reload"
	case ProjectKillStart:
		m = "Killing project"
	case ProjectKillDone:
		m = "Project killed"
	case ProjectStartStart:
		m = "Starting project"
	case ProjectStartDone:
		m = "Project started"
	case ProjectBuildStart:
		m = "Building project"
	case ProjectBuildDone:
		m = "Project built"
	case ProjectPauseStart:
		m = "Pausing project"
	case ProjectPauseDone:
		m = "Project paused"
	case ProjectUnpauseStart:
		m = "Unpausing project"
	case ProjectUnpauseDone:
		m = "Project unpaused"
	}

	if m == "" {
		m = fmt.Sprintf("EventType: %d", int(e))
	}

	return m
}
