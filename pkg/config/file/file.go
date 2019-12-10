// Package file implements decoding between configuration in a file and a typed Configuration.
package file

import (
	"github.com/containous/traefik/v2/pkg/config/parser"
)

// Decode decodes the given configuration file into the given element.
// The operation goes through three stages roughly summarized as:
// file contents -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> typed element
func Decode(filePath string, element interface{}) error {
	if element == nil {
		return nil
	}

	filters := getRootFieldNames(element)

	root, err := decodeFileToNode(filePath, filters...)
	if err != nil {
		return err
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagLabel, AllowSliceAsStruct: true}
	err = parser.AddMetadata(element, root, metaOpts)
	if err != nil {
		return err
	}

	return parser.Fill(element, root, parser.FillerOpts{AllowSliceAsStruct: true})
}
