// Package file implements decoding between configuration in a file and a typed Configuration.
package file

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/v2/pkg/config/parser"
	"gopkg.in/yaml.v2"
)

// Decode decodes the given configuration file into the given element.
// The operation goes through three stages roughly summarized as:
// file contents -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> typed element.
func Decode(filePath string, element interface{}) error {
	if element == nil {
		return nil
	}

	filters := getRootFieldNames(element)

	root, err := decodeFileToNode(filePath, filters...)
	if err != nil {
		return err
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagFile, AllowSliceAsStruct: false}
	err = parser.AddMetadata(element, root, metaOpts)
	if err != nil {
		return err
	}

	return parser.Fill(element, root, parser.FillerOpts{AllowSliceAsStruct: false})
}

// DecodeContent decodes the given configuration file content into the given element.
// The operation goes through three stages roughly summarized as:
// file contents -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> typed element.
func DecodeContent(content, extension string, element interface{}) error {
	data := make(map[string]interface{})

	switch extension {
	case ".toml":
		_, err := toml.Decode(content, &data)
		if err != nil {
			return err
		}

	case ".yml", ".yaml":
		var err error
		err = yaml.Unmarshal([]byte(content), &data)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported file extension: %s", extension)
	}

	filters := getRootFieldNames(element)

	node, err := decodeRawToNode(data, parser.DefaultRootName, filters...)
	if err != nil {
		return err
	}

	if len(node.Children) == 0 {
		return nil
	}

	metaOpts := parser.MetadataOpts{TagName: parser.TagFile, AllowSliceAsStruct: false}
	err = parser.AddMetadata(element, node, metaOpts)
	if err != nil {
		return err
	}

	return parser.Fill(element, node, parser.FillerOpts{AllowSliceAsStruct: false})
}
