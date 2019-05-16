package env

import (
	"strings"

	"github.com/containous/traefik/pkg/config/parser"
)

// Decode Converts the environment variables to an element.
// env vars -> [ node -> node + metadata (type) ] -> element (node)
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

// Encode Converts an element to environment variables.
// element -> node (value) -> env vars (node)
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
