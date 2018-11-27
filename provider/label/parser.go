package label

import (
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/provider/label/internal"
)

// Decode Converts the labels to a configuration.
// labels -> [ node -> node + metadata (type) ] -> element (node)
func Decode(labels map[string]string) (*config.Configuration, error) {
	node, err := internal.DecodeToNode(labels)
	if err != nil {
		return nil, err
	}

	conf := &config.Configuration{}
	err = internal.AddMetadata(conf, node)
	if err != nil {
		return nil, err
	}

	err = internal.Fill(conf, node)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// Encode Converts a configuration to labels.
// element -> node (value) -> label (node)
func Encode(conf *config.Configuration) (map[string]string, error) {
	node, err := internal.EncodeToNode(conf)
	if err != nil {
		return nil, err
	}

	return internal.EncodeNode(node), nil
}
