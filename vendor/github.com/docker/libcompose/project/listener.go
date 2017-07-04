package project

import (
	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project/events"
)

var (
	infoEvents = map[events.EventType]bool{
		events.ServiceDeleteStart:  true,
		events.ServiceDelete:       true,
		events.ServiceDownStart:    true,
		events.ServiceDown:         true,
		events.ServiceStopStart:    true,
		events.ServiceStop:         true,
		events.ServiceKillStart:    true,
		events.ServiceKill:         true,
		events.ServiceCreateStart:  true,
		events.ServiceCreate:       true,
		events.ServiceStartStart:   true,
		events.ServiceStart:        true,
		events.ServiceRestartStart: true,
		events.ServiceRestart:      true,
		events.ServiceUpStart:      true,
		events.ServiceUp:           true,
		events.ServicePauseStart:   true,
		events.ServicePause:        true,
		events.ServiceUnpauseStart: true,
		events.ServiceUnpause:      true,
	}
)

type defaultListener struct {
	project    *Project
	listenChan chan events.Event
	upCount    int
}

// NewDefaultListener create a default listener for the specified project.
func NewDefaultListener(p *Project) chan<- events.Event {
	l := defaultListener{
		listenChan: make(chan events.Event),
		project:    p,
	}
	go l.start()
	return l.listenChan
}

func (d *defaultListener) start() {
	for event := range d.listenChan {
		buffer := bytes.NewBuffer(nil)
		if event.Data != nil {
			for k, v := range event.Data {
				if buffer.Len() > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(k)
				buffer.WriteString("=")
				buffer.WriteString(v)
			}
		}

		if event.EventType == events.ServiceUp {
			d.upCount++
		}

		logf := logrus.Debugf

		if infoEvents[event.EventType] {
			logf = logrus.Infof
		}

		if event.ServiceName == "" {
			logf("Project [%s]: %s %s", d.project.Name, event.EventType, buffer.Bytes())
		} else {
			logf("[%d/%d] [%s]: %s %s", d.upCount, d.project.ServiceConfigs.Len(), event.ServiceName, event.EventType, buffer.Bytes())
		}
	}
}
