package flag

import (
	"github.com/containous/traefik/pkg/config/parser"
)

// Decode Converts the flags to an element.
// flags -> [ node -> node + metadata (type) ] -> element (node)
func Decode(args []string, element interface{}) error {
	ref, err := Parse(args, element)
	if err != nil {
		return err
	}

	return parser.Decode(ref, element)
}

// Encode Converts an element to Flat.
// element -> node (value) -> Flat
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

	return parser.EncodeToFlat(element, node, parser.FlatOpts{Separator: ".", SkipRoot: true})
}
