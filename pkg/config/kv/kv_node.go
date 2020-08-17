package kv

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/abronan/valkeyrie/store"
	"github.com/traefik/paerser/parser"
)

// DecodeToNode converts the labels to a tree of nodes.
// If any filters are present, labels which do not match the filters are skipped.
func DecodeToNode(pairs []*store.KVPair, rootName string, filters ...string) (*parser.Node, error) {
	sortedPairs := filterPairs(pairs, filters)

	exp := regexp.MustCompile(`^\d+$`)

	var node *parser.Node

	for i, pair := range sortedPairs {
		split := strings.FieldsFunc(pair.Key, func(c rune) bool { return c == '/' })

		if split[0] != rootName {
			return nil, fmt.Errorf("invalid label root %s", split[0])
		}

		var parts []string
		for _, fragment := range split {
			if exp.MatchString(fragment) {
				parts = append(parts, "["+fragment+"]")
			} else {
				parts = append(parts, fragment)
			}
		}

		if i == 0 {
			node = &parser.Node{}
		}
		decodeToNode(node, parts, string(pair.Value))
	}

	return node, nil
}

func decodeToNode(root *parser.Node, path []string, value string) {
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
			child := &parser.Node{Name: path[1]}
			decodeToNode(child, path[1:], value)
			root.Children = append(root.Children, child)
		}
	} else {
		root.Value = value
	}
}

func containsNode(nodes []*parser.Node, name string) *parser.Node {
	for _, n := range nodes {
		if strings.EqualFold(name, n.Name) {
			return n
		}
	}
	return nil
}

func filterPairs(pairs []*store.KVPair, filters []string) []*store.KVPair {
	exp := regexp.MustCompile(`^(.+)/\d+$`)

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})

	simplePairs := map[string]*store.KVPair{}
	slicePairs := map[string][]string{}

	for _, pair := range pairs {
		if len(filters) == 0 {
			// Slice of simple type
			if exp.MatchString(pair.Key) {
				sanitizedKey := exp.FindStringSubmatch(pair.Key)[1]
				slicePairs[sanitizedKey] = append(slicePairs[sanitizedKey], string(pair.Value))
			} else {
				simplePairs[pair.Key] = pair
			}
			continue
		}

		for _, filter := range filters {
			if len(pair.Key) >= len(filter) && strings.EqualFold(pair.Key[:len(filter)], filter) {
				// Slice of simple type
				if exp.MatchString(pair.Key) {
					sanitizedKey := exp.FindStringSubmatch(pair.Key)[1]
					slicePairs[sanitizedKey] = append(slicePairs[sanitizedKey], string(pair.Value))
				} else {
					simplePairs[pair.Key] = pair
				}
				continue
			}
		}
	}

	var sortedPairs []*store.KVPair
	for k, v := range slicePairs {
		delete(simplePairs, k)
		sortedPairs = append(sortedPairs, &store.KVPair{Key: k, Value: []byte(strings.Join(v, ","))})
	}

	for _, v := range simplePairs {
		sortedPairs = append(sortedPairs, v)
	}

	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[i].Key < sortedPairs[j].Key
	})

	return sortedPairs
}
