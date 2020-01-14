package parser

import (
	"fmt"
	"sort"
	"strings"
)

// DecodeToNode converts the labels to a tree of nodes.
// If any filters are present, labels which do not match the filters are skipped.
func DecodeToNode(labels map[string]string, rootName string, filters ...string) (*Node, error) {
	sortedKeys := sortKeys(labels, filters)

	var node *Node
	for i, key := range sortedKeys {
		split := strings.Split(key, ".")

		if split[0] != rootName {
			return nil, fmt.Errorf("invalid label root %s", split[0])
		}

		var parts []string
		for _, v := range split {
			if v == "" {
				return nil, fmt.Errorf("invalid element: %s", key)
			}

			if v[0] == '[' {
				return nil, fmt.Errorf("invalid leading character '[' in field name (bracket is a slice delimiter): %s", v)
			}

			if strings.HasSuffix(v, "]") && v[0] != '[' {
				indexLeft := strings.Index(v, "[")
				parts = append(parts, v[:indexLeft], v[indexLeft:])
			} else {
				parts = append(parts, v)
			}
		}

		if i == 0 {
			node = &Node{}
		}
		decodeToNode(node, parts, labels[key])
	}

	return node, nil
}

func decodeToNode(root *Node, path []string, value string) {
	if len(root.Name) == 0 {
		root.Name = path[0]
	}

	// it's a leaf or not -> children
	if len(path) > 1 {
		if n := containsNode(root.Children, path[1]); n != nil {
			// the child already exists
			decodeToNode(n, path[1:], value)
		} else {
			// new child
			child := &Node{Name: path[1]}
			decodeToNode(child, path[1:], value)
			root.Children = append(root.Children, child)
		}
	} else {
		root.Value = value
	}
}

func containsNode(nodes []*Node, name string) *Node {
	for _, n := range nodes {
		if strings.EqualFold(name, n.Name) {
			return n
		}
	}
	return nil
}

func sortKeys(labels map[string]string, filters []string) []string {
	var sortedKeys []string
	for key := range labels {
		if len(filters) == 0 {
			sortedKeys = append(sortedKeys, key)
			continue
		}

		for _, filter := range filters {
			if len(key) >= len(filter) && strings.EqualFold(key[:len(filter)], filter) {
				sortedKeys = append(sortedKeys, key)
				continue
			}
		}
	}
	sort.Strings(sortedKeys)

	return sortedKeys
}
