package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/traefik/paerser/parser"
)

func encodeNode(labels map[string]string, root string, node *parser.Node) {
	for _, child := range node.Children {
		if child.Disabled {
			continue
		}

		var sep string
		if child.Name[0] != '[' {
			sep = "."
		}

		childName := root + sep + child.Name

		if child.RawValue != nil {
			encodeRawValue(labels, childName, child.RawValue)
			continue
		}

		if strings.Contains(child.Tag.Get(parser.TagLabel), parser.TagLabelAllowEmpty) {
			labels[childName] = "true"
		}

		if len(child.Children) > 0 {
			encodeNode(labels, childName, child)
		} else if len(child.Name) > 0 {
			labels[childName] = child.Value
		}
	}
}

func encodeRawValue(labels map[string]string, root string, rawValue interface{}) {
	if rawValue == nil {
		return
	}

	tValue := reflect.TypeOf(rawValue)

	if tValue.Kind() == reflect.Map && tValue.Elem().Kind() == reflect.Interface {
		r := reflect.ValueOf(rawValue).
			Convert(reflect.TypeOf((map[string]interface{})(nil))).
			Interface().(map[string]interface{})

		for k, v := range r {
			switch tv := v.(type) {
			case string:
				labels[root+"."+k] = tv
			case []interface{}:
				for i, e := range tv {
					encodeRawValue(labels, fmt.Sprintf("%s.%s[%d]", root, k, i), e)
				}
			default:
				encodeRawValue(labels, root+"."+k, v)
			}
		}
	}
}
