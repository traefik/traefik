package internal

import (
	"fmt"
	"sort"
	"strings"
)

// DecodeToNode Converts the labels to a node.
// labels -> nodes
func DecodeToNode(labels map[string]string, filters ...string) (*Node, error) {
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

	labelRoot := "traefik"

	var node *Node
	for i, key := range sortedKeys {
		split := strings.Split(key, ".")

		if split[0] != labelRoot {
			// TODO (@ldez): error or continue
			return nil, fmt.Errorf("invalid label root %s", split[0])
		}

		labelRoot = split[0]

		if i == 0 {
			node = &Node{}
		}
		decodeToNode(node, split, labels[key])
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
		if name == n.Name {
			return n
		}
	}
	return nil
}
