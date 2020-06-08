// Package flag implements encoding and decoding between flag arguments and a typed Configuration.
package flag

import (
	"github.com/containous/traefik/v2/pkg/config/parser"
)

// Decode decodes the given flag arguments into the given element.
// The operation goes through four stages roughly summarized as:
// flag arguments -> parsed map of flags
// map -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> typed element.
func Decode(args []string, element interface{}) error {
	ref, err := Parse(args, element)
	if err != nil {
		return err
	}

	return parser.Decode(ref, element, parser.DefaultRootName)
}

// Encode encodes the configuration in element into the flags represented in the returned Flats.
// The operation goes through three stages roughly summarized as:
// typed configuration in element -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> flags with default values (determined by type/kind).
func Encode(element interface{}) ([]parser.Flat, error) {
	if element == nil {
		return nil, nil
	}

	etnOpts := parser.EncoderToNodeOpts{OmitEmpty: false, TagName: parser.TagLabel, AllowSliceAsStruct: true}
	node, err := parser.EncodeToNode(element, parser.DefaultRootName, etnOpts)
	if err != nil {
		return nil, err
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagLabel, AllowSliceAsStruct: true}
	err = parser.AddMetadata(element, node, metaOpts)
	if err != nil {
		return nil, err
	}

	flatOpts := parser.FlatOpts{Separator: ".", SkipRoot: true, TagName: parser.TagLabel}
	return parser.EncodeToFlat(element, node, flatOpts)
}
