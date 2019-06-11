// Package env implements encoding and decoding between environment variable and a configuration.
package env

import (
	"strings"

	"github.com/containous/traefik/pkg/config/parser"
)

// Decode decodes the given environment variables into the given element.
// The operation goes through three stages roughly summarized as:
// env vars -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as type (inferred from element)
// "typed" nodes -> typed element
func Decode(environ []string, element interface{}) error {
	vars := make(map[string]string)
	for _, evr := range environ {
		n := strings.SplitN(evr, "=", 2)
		if strings.HasPrefix(strings.ToUpper(n[0]), "TRAEFIK_") {
			key := strings.ReplaceAll(strings.ToLower(n[0]), "_", ".")
			vars[key] = n[1]
		}
	}

	return parser.Decode(vars, element)
}

// Encode encodes the configuration in element in the returned environment variables.
// The operation goes through three stages roughly summarized as:
// typed configuration in element -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as type (inferred from element)
// "typed" nodes -> environment variables with default values (determined by type)
func Encode(element interface{}) ([]parser.Flat, error) {
	if element == nil {
		return nil, nil
	}

	node, err := parser.EncodeToNode(element, false)
	if err != nil {
		return nil, err
	}

	err = parser.AddMetadata(element, node)
	if err != nil {
		return nil, err
	}

	return parser.EncodeToFlat(element, node, parser.FlatOpts{Case: "upper", Separator: "_"})
}
