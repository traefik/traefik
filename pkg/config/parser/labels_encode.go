package parser

// EncodeNode Converts a node to labels.
// nodes -> labels
func EncodeNode(node *Node) map[string]string {
	labels := make(map[string]string)
	encodeNode(labels, node.Name, node)
	return labels
}

func encodeNode(labels map[string]string, root string, node *Node) {
	for _, child := range node.Children {
		if child.Disabled {
			continue
		}

		var sep string
		if child.Name[0] != '[' {
			sep = "."
		}

		childName := root + sep + child.Name

		if len(child.Children) > 0 {
			encodeNode(labels, childName, child)
		} else if len(child.Name) > 0 {
			labels[childName] = child.Value
		}
	}
}
