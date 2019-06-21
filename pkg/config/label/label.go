// Package label implements the decoding and encoding between flat labels and a typed Configuration.
package label

import (
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/parser"
)

// DecodeConfiguration converts the labels to a configuration.
func DecodeConfiguration(labels map[string]string) (*config.Configuration, error) {
	conf := &config.Configuration{
		HTTP: &config.HTTPConfiguration{},
		TCP:  &config.TCPConfiguration{},
	}

	err := parser.Decode(labels, conf, parser.DefaultRootName, "traefik.http", "traefik.tcp")
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// EncodeConfiguration converts a configuration to labels.
func EncodeConfiguration(conf *config.Configuration) (map[string]string, error) {
	return parser.Encode(conf, parser.DefaultRootName)
}

// Decode converts the labels to an element.
// labels -> [ node -> node + metadata (type) ] -> element (node)
func Decode(labels map[string]string, element interface{}, filters ...string) error {
	return parser.Decode(labels, element, parser.DefaultRootName, filters...)
}
