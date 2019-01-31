package internal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/containous/flaeg/parse"
)

type initializer interface {
	SetDefaults()
}

// Fill the fields of the element.
// nodes -> element
func Fill(element interface{}, node *Node) error {
	if element == nil || node == nil {
		return nil
	}

	if node.Kind == 0 {
		return fmt.Errorf("missing node type: %s", node.Name)
	}

	elem := reflect.ValueOf(element)
	if elem.Kind() == reflect.Struct {
		return fmt.Errorf("struct are not supported, use pointer instead")
	}

	return fill(elem.Elem(), node)
}

func fill(field reflect.Value, node *Node) error {
	// related to allow-empty tag
	if node.Disabled {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(node.Value)
		return nil
	case reflect.Bool:
		val, err := strconv.ParseBool(node.Value)
		if err != nil {
			return err
		}
		field.SetBool(val)
		return nil
	case reflect.Int8:
		return setInt(field, node.Value, 8)
	case reflect.Int16:
		return setInt(field, node.Value, 16)
	case reflect.Int32:
		return setInt(field, node.Value, 32)
	case reflect.Int64, reflect.Int:
		return setInt(field, node.Value, 64)
	case reflect.Uint8:
		return setUint(field, node.Value, 8)
	case reflect.Uint16:
		return setUint(field, node.Value, 16)
	case reflect.Uint32:
		return setUint(field, node.Value, 32)
	case reflect.Uint64, reflect.Uint:
		return setUint(field, node.Value, 64)
	case reflect.Float32:
		return setFloat(field, node.Value, 32)
	case reflect.Float64:
		return setFloat(field, node.Value, 64)
	case reflect.Struct:
		return setStruct(field, node)
	case reflect.Ptr:
		return setPtr(field, node)
	case reflect.Map:
		return setMap(field, node)
	case reflect.Slice:
		return setSlice(field, node)
	default:
		return nil
	}
}

func setPtr(field reflect.Value, node *Node) error {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))

		if field.Type().Implements(reflect.TypeOf((*initializer)(nil)).Elem()) {
			method := field.MethodByName("SetDefaults")
			if method.IsValid() {
				method.Call([]reflect.Value{})
			}
		}
	}

	return fill(field.Elem(), node)
}

func setStruct(field reflect.Value, node *Node) error {
	for _, child := range node.Children {
		fd := field.FieldByName(child.FieldName)

		zeroValue := reflect.Value{}
		if fd == zeroValue {
			return fmt.Errorf("field not found, node: %s (%s)", child.Name, child.FieldName)
		}

		err := fill(fd, child)
		if err != nil {
			return err
		}
	}

	return nil
}

func setSlice(field reflect.Value, node *Node) error {
	if field.Type().Elem().Kind() == reflect.Struct {
		return setSliceAsStruct(field, node)
	}

	if len(node.Value) == 0 {
		return nil
	}

	values := strings.Split(node.Value, ",")

	slice := reflect.MakeSlice(field.Type(), len(values), len(values))
	field.Set(slice)

	for i := 0; i < len(values); i++ {
		value := strings.TrimSpace(values[i])

		switch field.Type().Elem().Kind() {
		case reflect.String:
			field.Index(i).Set(reflect.ValueOf(value))
		case reflect.Int:
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.Index(i).SetInt(val)
		case reflect.Int8:
			err := setInt(field.Index(i), value, 8)
			if err != nil {
				return err
			}
		case reflect.Int16:
			err := setInt(field.Index(i), value, 16)
			if err != nil {
				return err
			}
		case reflect.Int32:
			err := setInt(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Int64:
			err := setInt(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Uint:
			val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			field.Index(i).SetUint(val)
		case reflect.Uint8:
			err := setUint(field.Index(i), value, 8)
			if err != nil {
				return err
			}
		case reflect.Uint16:
			err := setUint(field.Index(i), value, 16)
			if err != nil {
				return err
			}
		case reflect.Uint32:
			err := setUint(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Uint64:
			err := setUint(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Float32:
			err := setFloat(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Float64:
			err := setFloat(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Bool:
			val, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.Index(i).SetBool(val)
		default:
			return fmt.Errorf("unsupported type: %s", field.Type().Elem())
		}
	}
	return nil
}

func setSliceAsStruct(field reflect.Value, node *Node) error {
	if len(node.Children) == 0 {
		return fmt.Errorf("invalid slice: node %s", node.Name)
	}

	// use Ptr to allow "SetDefaults"
	value := reflect.New(reflect.PtrTo(field.Type().Elem()))
	err := setPtr(value, node)
	if err != nil {
		return err
	}

	elem := value.Elem().Elem()

	field.Set(reflect.MakeSlice(field.Type(), 1, 1))
	field.Index(0).Set(elem)

	return nil
}

func setMap(field reflect.Value, node *Node) error {
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	for _, child := range node.Children {
		ptrValue := reflect.New(reflect.PtrTo(field.Type().Elem()))

		err := fill(ptrValue, child)
		if err != nil {
			return err
		}

		value := ptrValue.Elem().Elem()

		key := reflect.ValueOf(child.Name)
		field.SetMapIndex(key, value)
	}
	return nil
}

func setInt(field reflect.Value, value string, bitSize int) error {
	switch field.Type() {
	case reflect.TypeOf(parse.Duration(0)):
		return setDuration(field, value, bitSize, time.Second)
	case reflect.TypeOf(time.Duration(0)):
		return setDuration(field, value, bitSize, time.Nanosecond)
	default:
		val, err := strconv.ParseInt(value, 10, bitSize)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(val).Convert(field.Type()))
		return nil
	}
}

func setDuration(field reflect.Value, value string, bitSize int, defaultUnit time.Duration) error {
	val, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.Set(reflect.ValueOf(time.Duration(val) * defaultUnit).Convert(field.Type()))
		return nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(duration).Convert(field.Type()))
	return nil
}

func setUint(field reflect.Value, value string, bitSize int) error {
	val, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(val).Convert(field.Type()))
	return nil
}

func setFloat(field reflect.Value, value string, bitSize int) error {
	val, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(val).Convert(field.Type()))
	return nil
}
