package instana

import (
	l "log"
)

// Valid log levels
const (
	Error = 0
	Warn  = 1
	Info  = 2
	Debug = 3
)

type logS struct {
	sensor *sensorS
}

var log *logS

func (r *logS) makeV(prefix string, v ...interface{}) []interface{} {
	return append([]interface{}{prefix}, v...)
}

func (r *logS) debug(v ...interface{}) {
	if r.sensor.options.LogLevel >= Debug {
		l.Println(r.makeV("DEBUG: instana:", v...)...)
	}
}

func (r *logS) info(v ...interface{}) {
	if r.sensor.options.LogLevel >= Info {
		l.Println(r.makeV("INFO: instana:", v...)...)
	}
}

func (r *logS) warn(v ...interface{}) {
	if r.sensor.options.LogLevel >= Warn {
		l.Println(r.makeV("WARN: instana:", v...)...)
	}
}

func (r *logS) error(v ...interface{}) {
	if r.sensor.options.LogLevel >= Error {
		l.Println(r.makeV("ERROR: instana:", v...)...)
	}
}

func (r *sensorS) initLog() {
	log = new(logS)
	log.sensor = r
}
