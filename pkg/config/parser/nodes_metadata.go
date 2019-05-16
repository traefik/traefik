package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// AddMetadata Adds metadata to a node.
// nodes + element -> nodes
func AddMetadata(structure interface{}, node *Node) error {
	if node == nil {
		return nil
	}

	if len(node.Children) == 0 {
		return fmt.Errorf("invalid node %s: no child", node.Name)
	}

	if structure == nil {
		return errors.New("nil structure")
	}

	rootType := reflect.TypeOf(structure)
	node.Kind = rootType.Kind()

	return browseChildren(rootType, node)
}

func browseChildren(fType reflect.Type, node *Node) error {
	for _, child := range node.Children {
		if err := addMetadata(fType, child); err != nil {
			return err
		}
	}
	return nil
}

func addMetadata(rootType reflect.Type, node *Node) error {
	rType := rootType
	if rootType.Kind() == reflect.Ptr {
		rType = rootType.Elem()
	}

	field, err := findTypedField(rType, node)
	if err != nil {
		return err
	}

	if err = isSupportedType(field); err != nil {
		return err
	}

	fType := field.Type
	node.Kind = fType.Kind()
	node.Tag = field.Tag

	if fType.Kind() == reflect.Struct || fType.Kind() == reflect.Ptr && fType.Elem().Kind() == reflect.Struct ||
		fType.Kind() == reflect.Map {
		if len(node.Children) == 0 && field.Tag.Get(TagLabel) != TagLabelAllowEmpty {
			return fmt.Errorf("%s cannot be a standalone element (type %s)", node.Name, fType)
		}

		node.Disabled = len(node.Value) > 0 && !strings.EqualFold(node.Value, "true") && field.Tag.Get(TagLabel) == TagLabelAllowEmpty
	}

	if len(node.Children) == 0 {
		return nil
	}

	if fType.Kind() == reflect.Struct || fType.Kind() == reflect.Ptr && fType.Elem().Kind() == reflect.Struct {
		return browseChildren(fType, node)
	}

	if fType.Kind() == reflect.Map {
		for _, child := range node.Children {
			// elem is a map entry value type
			elem := fType.Elem()
			child.Kind = elem.Kind()

			if elem.Kind() == reflect.Map || elem.Kind() == reflect.Struct ||
				(elem.Kind() == reflect.Ptr && elem.Elem().Kind() == reflect.Struct) {
				if err = browseChildren(elem, child); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if fType.Kind() == reflect.Slice {
		if field.Tag.Get(TagLabelSliceAsStruct) != "" {
			return browseChildren(fType.Elem(), node)
		}

		for _, ch := range node.Children {
			ch.Kind = fType.Elem().Kind()
			if err = browseChildren(fType.Elem(), ch); err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("invalid node %s: %v", node.Name, fType.Kind())
}

func findTypedField(rType reflect.Type, node *Node) (reflect.StructField, error) {
	for i := 0; i < rType.NumField(); i++ {
		cField := rType.Field(i)

		fieldName := cField.Tag.Get(TagLabelSliceAsStruct)
		if len(fieldName) == 0 {
			fieldName = cField.Name
		}

		if IsExported(cField) {
			if cField.Anonymous {
				if cField.Type.Kind() == reflect.Struct {
					structField, err := findTypedField(cField.Type, node)
					if err != nil {
						continue
					}
					return structField, nil
				}
			}

			if strings.EqualFold(fieldName, node.Name) {
				node.FieldName = cField.Name
				return cField, nil
			}
		}

	}

	return reflect.StructField{}, fmt.Errorf("field not found, node: %s", node.Name)
}

// IsExported return true is a struct field is exported, else false
// https://golang.org/pkg/reflect/#StructField
func IsExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

func isSupportedType(field reflect.StructField) error {
	fType := field.Type

	if fType.Kind() == reflect.Slice {
		switch fType.Elem().Kind() {
		case reflect.String,
			reflect.Bool,
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64,
			reflect.Struct,
			reflect.Ptr:
			return nil
		default:
			return fmt.Errorf("unsupported slice type: %v", fType)
		}
	}

	if fType.Kind() == reflect.Map && fType.Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type: %v", fType.Key())
	}

	if fType.Kind() == reflect.Func {
		return fmt.Errorf("unsupported type: %v", fType)
	}

	return nil
}
