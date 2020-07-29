package file

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/parser"
)

func decodeRawToNode(data map[string]interface{}, rootName string, filters ...string) (*parser.Node, error) {
	root := &parser.Node{
		Name: rootName,
	}

	vData := reflect.ValueOf(data)
	err := decodeRaw(root, vData, filters...)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func decodeRaw(node *parser.Node, vData reflect.Value, filters ...string) error {
	sortedKeys := sortKeys(vData, filters)

	for _, key := range sortedKeys {
		if vData.MapIndex(key).IsNil() {
			continue
		}

		value := reflect.ValueOf(vData.MapIndex(key).Interface())

		child := &parser.Node{Name: key.String()}

		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fallthrough
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fallthrough
		case reflect.Float32, reflect.Float64:
			fallthrough
		case reflect.Bool:
			fallthrough
		case reflect.String:
			value, err := getSimpleValue(value)
			if err != nil {
				return err
			}
			child.Value = value
		case reflect.Slice:
			var values []string

			for i := 0; i < value.Len(); i++ {
				item := value.Index(i)
				switch item.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					fallthrough
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					fallthrough
				case reflect.Bool:
					fallthrough
				case reflect.String:
					fallthrough
				case reflect.Map:
					fallthrough
				case reflect.Interface:
					sValue := reflect.ValueOf(item.Interface())
					if sValue.Kind() == reflect.Map {
						ch := &parser.Node{
							Name: "[" + strconv.Itoa(i) + "]",
						}

						child.Children = append(child.Children, ch)
						err := decodeRaw(ch, sValue)
						if err != nil {
							return err
						}
					} else {
						val, err := getSimpleValue(sValue)
						if err != nil {
							return err
						}
						values = append(values, val)
					}
				default:
					return fmt.Errorf("field %s uses unsupported slice type: %s", child.Name, item.Kind().String())
				}
			}

			child.Value = strings.Join(values, ",")
		case reflect.Map:
			err := decodeRaw(child, value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("field %s uses unsupported type: %s", child.Name, value.Kind().String())
		}

		node.Children = append(node.Children, child)
	}

	return nil
}

func getSimpleValue(item reflect.Value) (string, error) {
	switch item.Kind() {
	case reflect.String:
		return item.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(item.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(item.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strings.TrimSuffix(strconv.FormatFloat(item.Float(), 'f', 6, 64), ".000000"), nil
	case reflect.Bool:
		return strconv.FormatBool(item.Bool()), nil
	default:
		return "", fmt.Errorf("unsupported simple value type: %s", item.Kind().String())
	}
}

func sortKeys(vData reflect.Value, filters []string) []reflect.Value {
	var sortedKeys []reflect.Value

	for _, v := range vData.MapKeys() {
		rValue := reflect.ValueOf(v.Interface())
		key := rValue.String()

		if len(filters) == 0 {
			sortedKeys = append(sortedKeys, rValue)
			continue
		}

		for _, filter := range filters {
			if strings.EqualFold(key, filter) {
				sortedKeys = append(sortedKeys, rValue)
				continue
			}
		}
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].String() < sortedKeys[j].String()
	})

	return sortedKeys
}
