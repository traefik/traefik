package file

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/containous/traefik/pkg/config/parser"
)

func decodeRawToNode(data map[string]interface{}, excludes ...string) (*parser.Node, error) {
	root := &parser.Node{
		Name: "traefik",
	}

	vData := reflect.ValueOf(data)
	decodeRaw(root, vData, excludes...)

	return root, nil
}

func decodeRaw(node *parser.Node, vData reflect.Value, excludes ...string) {
	sortedKeys := sortKeys(vData, excludes)

	for _, key := range sortedKeys {
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
			child.Value = getSimpleValue(value)
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
						decodeRaw(ch, sValue)
					} else {
						values = append(values, getSimpleValue(sValue))
					}
				default:
					panic("Unsupported slice type: " + item.Kind().String())
				}
			}

			child.Value = strings.Join(values, ",")
		case reflect.Map:
			decodeRaw(child, value)
		default:
			panic("Unsupported type: " + value.Kind().String())
		}

		node.Children = append(node.Children, child)
	}
}

func getSimpleValue(item reflect.Value) string {
	switch item.Kind() {
	case reflect.String:
		return item.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(item.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(item.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strings.TrimSuffix(strconv.FormatFloat(item.Float(), 'f', 6, 64), ".000000")
	case reflect.Bool:
		return strconv.FormatBool(item.Bool())
	default:
		panic("Unsupported Simple value type: " + item.Kind().String())
	}
}

func sortKeys(vData reflect.Value, excludes []string) []reflect.Value {
	var sortedKeys []reflect.Value

	for _, v := range vData.MapKeys() {
		rValue := reflect.ValueOf(v.Interface())
		key := rValue.String()

		if len(excludes) == 0 {
			sortedKeys = append(sortedKeys, rValue)
			continue
		}

		var found bool
		for _, filter := range excludes {
			if strings.EqualFold(key, filter) {
				found = true
				break
			}
		}

		if !found {
			sortedKeys = append(sortedKeys, rValue)
		}
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i].String() < sortedKeys[j].String()
	})

	return sortedKeys
}
