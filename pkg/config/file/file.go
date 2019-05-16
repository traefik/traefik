package file

import (
	"github.com/containous/traefik/pkg/config/parser"
)

// Decode converts the file to an element.
// file -> [ node -> node + metadata (type) ] -> element (node)
func Decode(filePath string, element interface{}, excludes ...string) error {
	if element == nil {
		return nil
	}

	root, err := decodeFileToNode(filePath, excludes...)
	if err != nil {
		return err
	}

	err = parser.AddMetadata(element, root)
	if err != nil {
		return err
	}

	return parser.Fill(element, root)
}
