package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// EncoderToNodeOpts Options for the encoderToNode.
type EncoderToNodeOpts struct {
	OmitEmpty          bool
	TagName            string
	AllowSliceAsStruct bool
}

// EncodeToNode converts an element to a node.
// element -> nodes.
func EncodeToNode(element interface{}, rootName string, opts EncoderToNodeOpts) (*Node, error) {
	rValue := reflect.ValueOf(element)
	node := &Node{Name: rootName}

	encoder := encoderToNode{EncoderToNodeOpts: opts}

	err := encoder.setNodeValue(node, rValue)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type encoderToNode struct {
	EncoderToNodeOpts
}

func (e encoderToNode) setNodeValue(node *Node, rValue reflect.Value) error {
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
		return e.setStructValue(node, rValue)
	case reflect.Ptr:
		return e.setNodeValue(node, rValue.Elem())
	case reflect.Map:
		return e.setMapValue(node, rValue)
	case reflect.Slice:
		return e.setSliceValue(node, rValue)
	default:
		// noop
	}

	return nil
}

func (e encoderToNode) setStructValue(node *Node, rValue reflect.Value) error {
	rType := rValue.Type()

	for i := 0; i < rValue.NumField(); i++ {
		field := rType.Field(i)
		fieldValue := rValue.Field(i)

		if !IsExported(field) {
			continue
		}

		if field.Tag.Get(e.TagName) == "-" {
			continue
		}

		if err := isSupportedType(field); err != nil {
			return err
		}

		if e.isSkippedField(field, fieldValue) {
			continue
		}

		nodeName := field.Name
		if e.AllowSliceAsStruct && field.Type.Kind() == reflect.Slice && len(field.Tag.Get(TagLabelSliceAsStruct)) != 0 {
			nodeName = field.Tag.Get(TagLabelSliceAsStruct)
		}

		if field.Anonymous {
			if err := e.setNodeValue(node, fieldValue); err != nil {
				return err
			}
			continue
		}

		child := &Node{Name: nodeName, FieldName: field.Name, Description: field.Tag.Get(TagDescription)}

		if err := e.setNodeValue(child, fieldValue); err != nil {
			return err
		}

		if field.Type.Kind() == reflect.Ptr {
			if field.Type.Elem().Kind() != reflect.Struct && fieldValue.IsNil() {
				continue
			}

			if field.Type.Elem().Kind() == reflect.Struct && len(child.Children) == 0 {
				if field.Tag.Get(e.TagName) != TagLabelAllowEmpty {
					continue
				}

				child.Value = "true"
			}
		}

		node.Children = append(node.Children, child)
	}

	return nil
}

func (e encoderToNode) setMapValue(node *Node, rValue reflect.Value) error {
	if rValue.Type().Elem().Kind() == reflect.Interface {
		node.RawValue = rValue.Interface()
		return nil
	}

	for _, key := range rValue.MapKeys() {
		child := &Node{Name: key.String(), FieldName: key.String()}
		node.Children = append(node.Children, child)

		if err := e.setNodeValue(child, rValue.MapIndex(key)); err != nil {
			return err
		}
	}
	return nil
}

func (e encoderToNode) setSliceValue(node *Node, rValue reflect.Value) error {
	// label-slice-as-struct
	if rValue.Type().Elem().Kind() == reflect.Struct && !strings.EqualFold(node.Name, node.FieldName) {
		if rValue.Len() > 1 {
			return fmt.Errorf("node %s has too many slice entries: %d", node.Name, rValue.Len())
		}

		return e.setNodeValue(node, rValue.Index(0))
	}

	if rValue.Type().Elem().Kind() == reflect.Struct ||
		rValue.Type().Elem().Kind() == reflect.Ptr && rValue.Type().Elem().Elem().Kind() == reflect.Struct {
		for i := 0; i < rValue.Len(); i++ {
			child := &Node{Name: "[" + strconv.Itoa(i) + "]"}

			eValue := rValue.Index(i)

			err := e.setNodeValue(child, eValue)
			if err != nil {
				return err
			}

			node.Children = append(node.Children, child)
		}

		return nil
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

func (e encoderToNode) isSkippedField(field reflect.StructField, fieldValue reflect.Value) bool {
	if e.OmitEmpty && field.Type.Kind() == reflect.String && fieldValue.Len() == 0 {
		return true
	}

	if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && fieldValue.IsNil() {
		return true
	}

	if e.OmitEmpty && (field.Type.Kind() == reflect.Slice) &&
		(fieldValue.IsNil() || fieldValue.Len() == 0) {
		return true
	}

	if (field.Type.Kind() == reflect.Map) &&
		(fieldValue.IsNil() || fieldValue.Len() == 0) {
		return true
	}

	return false
}
