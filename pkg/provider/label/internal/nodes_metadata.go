package internal

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

	if fType.Kind() == reflect.Struct || fType.Kind() == reflect.Ptr && fType.Elem().Kind() == reflect.Struct ||
		fType.Kind() == reflect.Map {
		if len(node.Children) == 0 && field.Tag.Get(TagLabel) != "allowEmpty" {
			return fmt.Errorf("node %s (type %s) must have children", node.Name, fType)
		}

		node.Disabled = len(node.Value) > 0 && node.Value != "true" && field.Tag.Get(TagLabel) == "allowEmpty"
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

	// only for struct/Ptr with label-slice-as-struct tag
	if fType.Kind() == reflect.Slice {
		return browseChildren(fType.Elem(), node)
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

		if isExported(cField) && strings.EqualFold(fieldName, node.Name) {
			node.FieldName = cField.Name
			return cField, nil
		}
	}

	return reflect.StructField{}, fmt.Errorf("field not found, node: %s", node.Name)
}

func browseChildren(fType reflect.Type, node *Node) error {
	for _, child := range node.Children {
		if err := addMetadata(fType, child); err != nil {
			return err
		}
	}
	return nil
}

// isExported return true is a struct field is exported, else false
// https://golang.org/pkg/reflect/#StructField
func isExported(f reflect.StructField) bool {
	if f.PkgPath != "" && !f.Anonymous {
		return false
	}
	return true
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
			reflect.Float64:
			return nil
		default:
			if len(field.Tag.Get(TagLabelSliceAsStruct)) > 0 {
				return nil
			}
			return fmt.Errorf("unsupported slice type: %v", fType)
		}
	}

	if fType.Kind() == reflect.Ptr && fType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("unsupported pointer type: %v", fType.Elem())
	}

	if fType.Kind() == reflect.Map && fType.Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type: %v", fType.Key())
	}

	if fType.Kind() == reflect.Func {
		return fmt.Errorf("unsupported type: %v", fType)
	}

	return nil
}
