// Package parser implements decoding and encoding between a flat map of labels and a typed Configuration.
package parser

// Decode decodes the given map of labels into the given element.
// If any filters are present, labels which do not match the filters are skipped.
// The operation goes through three stages roughly summarized as:
// labels -> tree of untyped nodes
// untyped nodes -> nodes augmented with metadata such as kind (inferred from element)
// "typed" nodes -> typed element.
func Decode(labels map[string]string, element interface{}, rootName string, filters ...string) error {
	node, err := DecodeToNode(labels, rootName, filters...)
	if err != nil {
		return err
	}

	metaOpts := MetadataOpts{TagName: TagLabel, AllowSliceAsStruct: true}
	err = AddMetadata(element, node, metaOpts)
	if err != nil {
		return err
	}

	return Fill(element, node, FillerOpts{AllowSliceAsStruct: true})
}

// Encode converts an element to labels.
// element -> node (value) -> label (node).
func Encode(element interface{}, rootName string) (map[string]string, error) {
	etnOpts := EncoderToNodeOpts{OmitEmpty: true, TagName: TagLabel, AllowSliceAsStruct: true}
	node, err := EncodeToNode(element, rootName, etnOpts)
	if err != nil {
		return nil, err
	}

	return EncodeNode(node), nil
}
