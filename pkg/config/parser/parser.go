package parser

// Decode Converts the labels to an element.
// labels -> [ node -> node + metadata (type) ] -> element (node)
func Decode(labels map[string]string, element interface{}, filters ...string) error {
	node, err := DecodeToNode(labels, filters...)
	if err != nil {
		return err
	}

	err = AddMetadata(element, node)
	if err != nil {
		return err
	}

	err = Fill(element, node)
	if err != nil {
		return err
	}

	return nil
}

// Encode Converts an element to labels.
// element -> node (value) -> label (node)
func Encode(element interface{}) (map[string]string, error) {
	node, err := EncodeToNode(element, true)
	if err != nil {
		return nil, err
	}

	return EncodeNode(node), nil
}
