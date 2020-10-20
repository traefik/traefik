package log

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
)

const fluentPort = 24224

// NewFluentHook return a fluentd hook for logrus logging
func NewFluentHook(level logrus.Level, endpoint string) (*logrus_fluent.FluentHook, error) {
	// parse url
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Can't parse fluent endpoint: %v", err)
	}

	// parse port, use default one if none provided
	port := fluentPort
	if url.Port() != "" {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			return nil, err

		}
	}

	hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
		Host: url.Hostname(),
		Port: port,
	})
	if err != nil {
		return nil, err
	}

	// set loglevel to level defined in config
	hook.SetLevels([]logrus.Level{level})
	hook.SetTag("traefik.tag")

	return hook, err
}
