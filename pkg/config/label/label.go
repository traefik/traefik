// Package label implements the decoding and encoding between flat labels and a typed Configuration.
package label

import (
	"github.com/traefik/paerser/parser"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// DecodeConfiguration converts the labels to a configuration.
func DecodeConfiguration(labels map[string]string) (*dynamic.Configuration, error) {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{},
		TCP:  &dynamic.TCPConfiguration{},
		UDP:  &dynamic.UDPConfiguration{},
		TLS:  &dynamic.TLSConfiguration{},
	}

	// When decoding the TLS configuration we are making sure that only the default TLS store can be configured.
	err := parser.Decode(labels, conf, parser.DefaultRootName, "traefik.http", "traefik.tcp", "traefik.udp", "traefik.tls.stores.default")
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// EncodeConfiguration converts a configuration to labels.
func EncodeConfiguration(conf *dynamic.Configuration) (map[string]string, error) {
	return parser.Encode(conf, parser.DefaultRootName)
}

// Decode converts the labels to an element.
// labels -> [ node -> node + metadata (type) ] -> element (node).
func Decode(labels map[string]string, element interface{}, filters ...string) error {
	return parser.Decode(labels, element, parser.DefaultRootName, filters...)
}
