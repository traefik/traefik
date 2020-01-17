package parser

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/containous/traefik/v2/pkg/types"
)

const defaultPtrValue = "false"

// FlatOpts holds options used when encoding to Flat.
type FlatOpts struct {
	Case      string // "lower" or "upper", defaults to "lower".
	Separator string
	SkipRoot  bool
	TagName   string
}

// Flat is a configuration item representation.
type Flat struct {
	Name        string
	Description string
	Default     string
}

// EncodeToFlat encodes a node to a Flat representation.
// Even though the given node argument should have already been augmented with metadata such as kind,
// the element (and its type information) is still needed to treat remaining edge cases.
func EncodeToFlat(element interface{}, node *Node, opts FlatOpts) ([]Flat, error) {
	if element == nil || node == nil {
		return nil, nil
	}

	if node.Kind == 0 {
		return nil, fmt.Errorf("missing node type: %s", node.Name)
	}

	elem := reflect.ValueOf(element)
	if elem.Kind() == reflect.Struct {
		return nil, fmt.Errorf("structs are not supported, use pointer instead")
	}

	encoder := encoderToFlat{FlatOpts: opts}

	var entries []Flat
	if encoder.SkipRoot {
		for _, child := range node.Children {
			field := encoder.getField(elem.Elem(), child)
			entries = append(entries, encoder.createFlat(field, child.Name, child)...)
		}
	} else {
		entries = encoder.createFlat(elem, strings.ToLower(node.Name), node)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	return entries, nil
}

type encoderToFlat struct {
	FlatOpts
}

func (e encoderToFlat) createFlat(field reflect.Value, name string, node *Node) []Flat {
	var entries []Flat
	if node.Kind != reflect.Map && node.Description != "-" {
		if !(node.Kind == reflect.Ptr && len(node.Children) > 0) ||
			(node.Kind == reflect.Ptr && node.Tag.Get(e.TagName) == TagLabelAllowEmpty) {
			if node.Name[0] != '[' {
				entries = append(entries, Flat{
					Name:        e.getName(name),
					Description: node.Description,
					Default:     e.getNodeValue(e.getField(field, node), node),
				})
			}
		}
	}

	for _, child := range node.Children {
		if node.Kind == reflect.Map {
			fChild := e.getField(field, child)

			var v string
			if child.Kind == reflect.Struct {
				v = defaultPtrValue
			} else {
				v = e.getNodeValue(fChild, child)
			}

			if node.Description != "-" {
				entries = append(entries, Flat{
					Name:        e.getName(name, child.Name),
					Description: node.Description,
					Default:     v,
				})
			}

			if child.Kind == reflect.Struct || child.Kind == reflect.Ptr {
				for _, ch := range child.Children {
					f := e.getField(fChild, ch)
					n := e.getName(name, child.Name, ch.Name)
					entries = append(entries, e.createFlat(f, n, ch)...)
				}
			}
		} else {
			f := e.getField(field, child)
			n := e.getName(name, child.Name)
			entries = append(entries, e.createFlat(f, n, child)...)
		}
	}

	return entries
}

func (e encoderToFlat) getField(field reflect.Value, node *Node) reflect.Value {
	switch field.Kind() {
	case reflect.Struct:
		return field.FieldByName(node.FieldName)
	case reflect.Ptr:
		if field.Elem().Kind() == reflect.Struct {
			return field.Elem().FieldByName(node.FieldName)
		}
		return field.Elem()
	case reflect.Map:
		return field.MapIndex(reflect.ValueOf(node.FieldName))
	default:
		return field
	}
}

func (e encoderToFlat) getNodeValue(field reflect.Value, node *Node) string {
	if node.Kind == reflect.Ptr && len(node.Children) > 0 {
		return defaultPtrValue
	}

	if field.Kind() == reflect.Int64 {
		i, _ := strconv.ParseInt(node.Value, 10, 64)

		switch field.Type() {
		case reflect.TypeOf(types.Duration(time.Second)):
			return strconv.Itoa(int(i) / int(time.Second))
		case reflect.TypeOf(time.Second):
			return time.Duration(i).String()
		}
	}

	return node.Value
}

func (e encoderToFlat) getName(names ...string) string {
	var name string
	if names[len(names)-1][0] == '[' {
		name = strings.Join(names, "")
	} else {
		name = strings.Join(names, e.Separator)
	}

	if strings.EqualFold(e.Case, "upper") {
		return strings.ToUpper(name)
	}
	return strings.ToLower(name)
}
