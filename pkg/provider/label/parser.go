package label

import (
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/provider/label/internal"
)

// DecodeConfiguration Converts the labels to a configuration.
func DecodeConfiguration(labels map[string]string) (*config.Configuration, error) {
	conf := &config.Configuration{
		HTTP: &config.HTTPConfiguration{},
		TCP:  &config.TCPConfiguration{},
	}

	err := Decode(labels, conf, "traefik.http", "traefik.tcp")
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// EncodeConfiguration Converts a configuration to labels.
func EncodeConfiguration(conf *config.Configuration) (map[string]string, error) {
	return Encode(conf)
}

// Decode Converts the labels to an element.
// labels -> [ node -> node + metadata (type) ] -> element (node)
func Decode(labels map[string]string, element interface{}, filters ...string) error {
	node, err := internal.DecodeToNode(labels, filters...)
	if err != nil {
		return err
	}

	err = internal.AddMetadata(element, node)
	if err != nil {
		return err
	}

	err = internal.Fill(element, node)
	if err != nil {
		return err
	}

	return nil
}

// Encode Converts an element to labels.
// element -> node (value) -> label (node)
func Encode(element interface{}) (map[string]string, error) {
	node, err := internal.EncodeToNode(element)
	if err != nil {
		return nil, err
	}

	return internal.EncodeNode(node), nil
}
