package log

import (
	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
)

// NewFluentHook return a fluentd hook for logrus loggin
func NewFluentHook(level logrus.Level) (*logrus_fluent.FluentHook, error) {
	// add fluentd hooks if specified in configuration
	// TODO: dynamic fluentd configuration
	hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Host: "localhost",
		Port: 24224,
	})
	if err != nil {
		return nil, err
	}

	// TODO: check for tags and stuff
	hook.SetLevels([]logrus.Level{level})
	hook.SetTag("traefik.tag")

	return hook, err
}
