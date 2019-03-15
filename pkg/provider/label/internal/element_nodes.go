package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EncodeToNode Converts an element to a node.
// element -> nodes
func EncodeToNode(element interface{}) (*Node, error) {
	rValue := reflect.ValueOf(element)
	node := &Node{Name: "traefik"}

	err := setNodeValue(node, rValue)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func setNodeValue(node *Node, rValue reflect.Value) error {
	switch rValue.Kind() {
	case reflect.String:
		node.Value = rValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		node.Value = strconv.FormatInt(rValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		node.Value = strconv.FormatUint(rValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		node.Value = strconv.FormatFloat(rValue.Float(), 'f', 6, 64)
	case reflect.Bool:
		node.Value = strconv.FormatBool(rValue.Bool())
	case reflect.Struct:
		return setStructValue(node, rValue)
	case reflect.Ptr:
		return setNodeValue(node, rValue.Elem())
	case reflect.Map:
		return setMapValue(node, rValue)
	case reflect.Slice:
		return setSliceValue(node, rValue)
	default:
		// noop
	}

	return nil
}

func setStructValue(node *Node, rValue reflect.Value) error {
	rType := rValue.Type()

	for i := 0; i < rValue.NumField(); i++ {
		field := rType.Field(i)
		fieldValue := rValue.Field(i)

		if !isExported(field) {
			continue
		}

		if field.Tag.Get(TagLabel) == "-" {
			continue
		}

		if err := isSupportedType(field); err != nil {
			return err
		}

		if isSkippedField(field, fieldValue) {
			continue
		}

		nodeName := field.Name
		if field.Type.Kind() == reflect.Slice && len(field.Tag.Get(TagLabelSliceAsStruct)) != 0 {
			nodeName = field.Tag.Get(TagLabelSliceAsStruct)
		}

		child := &Node{Name: nodeName, FieldName: field.Name}

		if err := setNodeValue(child, fieldValue); err != nil {
			return err
		}

		if field.Type.Kind() == reflect.Ptr && len(child.Children) == 0 {
			if field.Tag.Get(TagLabel) != "allowEmpty" {
				continue
			}

			child.Value = "true"
		}

		node.Children = append(node.Children, child)
	}

	return nil
}

func setMapValue(node *Node, rValue reflect.Value) error {
	for _, key := range rValue.MapKeys() {
		child := &Node{Name: key.String(), FieldName: key.String()}
		node.Children = append(node.Children, child)

		if err := setNodeValue(child, rValue.MapIndex(key)); err != nil {
			return err
		}
	}
	return nil
}

func setSliceValue(node *Node, rValue reflect.Value) error {
	// label-slice-as-struct
	if rValue.Type().Elem().Kind() == reflect.Struct && !strings.EqualFold(node.Name, node.FieldName) {
		if rValue.Len() > 1 {
			return fmt.Errorf("node %s has too many slice entries: %d", node.Name, rValue.Len())
		}

		if err := setNodeValue(node, rValue.Index(0)); err != nil {
			return err
		}
	}

	var values []string

	for i := 0; i < rValue.Len(); i++ {
		eValue := rValue.Index(i)

		switch eValue.Kind() {
		case reflect.String:
			values = append(values, eValue.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values = append(values, strconv.FormatInt(eValue.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			values = append(values, strconv.FormatUint(eValue.Uint(), 10))
		case reflect.Float32, reflect.Float64:
			values = append(values, strconv.FormatFloat(eValue.Float(), 'f', 6, 64))
		case reflect.Bool:
			values = append(values, strconv.FormatBool(eValue.Bool()))
		default:
			// noop
		}
	}

	node.Value = strings.Join(values, ", ")
	return nil
}

func isSkippedField(field reflect.StructField, fieldValue reflect.Value) bool {
	if field.Type.Kind() == reflect.String && fieldValue.Len() == 0 {
		return true
	}

	if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && fieldValue.IsNil() {
		return true
	}

	if (field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map) &&
		(fieldValue.IsNil() || fieldValue.Len() == 0) {
		return true
	}

	return false
}
