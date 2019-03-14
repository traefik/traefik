package egoscale

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func csQuotePlus(s string) string {
	s = strings.Replace(s, "+", "%20", -1)
	return s
}

func csEncode(s string) string {
	return csQuotePlus(url.QueryEscape(s))
}

// info returns the meta info of a command
//
// command is not a Command so it's easier to Test
func info(command interface{}) (*CommandInfo, error) {
	typeof := reflect.TypeOf(command)

	// Going up the pointer chain to find the underlying struct
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
	}

	field, ok := typeof.FieldByName("_")
	if !ok {
		return nil, fmt.Errorf(`missing meta ("_") field in %#v`, command)
	}

	name, nameOk := field.Tag.Lookup("name")
	description, _ := field.Tag.Lookup("description")

	if !nameOk {
		return nil, fmt.Errorf(`missing "name" key in the tag string of %#v`, command)
	}

	info := &CommandInfo{
		Name:        name,
		Description: description,
	}

	return info, nil
}

// prepareValues uses a command to build a POST request
//
// command is not a Command so it's easier to Test
func prepareValues(prefix string, command interface{}) (url.Values, error) {
	params := url.Values{}

	value := reflect.ValueOf(command)
	typeof := reflect.TypeOf(command)

	// Going up the pointer chain to find the underlying struct
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		value = value.Elem()
	}

	// Checking for nil commands
	if !value.IsValid() {
		return nil, fmt.Errorf("cannot serialize the invalid value %#v", command)
	}

	for i := 0; i < typeof.NumField(); i++ {
		field := typeof.Field(i)
		if field.Name == "_" {
			continue
		}

		val := value.Field(i)
		tag := field.Tag

		var err error
		var name string
		var value interface{}

		if json, ok := tag.Lookup("json"); ok {
			n, required := ExtractJSONTag(field.Name, json)
			name = prefix + n

			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value, err = prepareInt(val.Int(), required)

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value, err = prepareUint(val.Uint(), required)

			case reflect.Float32, reflect.Float64:
				value, err = prepareFloat(val.Float(), required)

			case reflect.String:
				value, err = prepareString(val.String(), required)

			case reflect.Bool:
				value, err = prepareBool(val.Bool(), required)

			case reflect.Map:
				if val.Len() == 0 {
					if required {
						err = fmt.Errorf("field is required, got empty map")
					}
				} else {
					value, err = prepareMap(name, val.Interface())
				}

			case reflect.Ptr:
				value, err = preparePtr(field.Type.Elem().Kind(), val, required)

			case reflect.Slice:
				value, err = prepareSlice(name, field.Type, val, required)

			case reflect.Struct:
				value, err = prepareStruct(val.Interface(), required)

			default:
				if required {
					err = fmt.Errorf("unsupported type")
				}
			}
		} else {
			switch val.Kind() {
			case reflect.Struct:
				value, err = prepareEmbedStruct(val.Interface())
			default:
				log.Printf("[SKIP] %s.%s no json label found", typeof.Name(), field.Name)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("%s.%s (%v) %s", typeof.Name(), field.Name, val.Kind(), err)
		}

		switch v := value.(type) {
		case *string:
			if name != "" && v != nil {
				params.Set(name, *v)
			}
		case url.Values:
			for k, xs := range v {
				for _, x := range xs {
					params.Add(k, x)
				}
			}
		}
	}

	return params, nil
}

func prepareInt(v int64, required bool) (*string, error) {
	if v == 0 {
		if required {
			return nil, fmt.Errorf("field is required, got %d", v)
		}
		return nil, nil
	}
	value := strconv.FormatInt(v, 10)
	return &value, nil
}

func prepareUint(v uint64, required bool) (*string, error) {
	if v == 0 {
		if required {
			return nil, fmt.Errorf("field is required, got %d", v)
		}
		return nil, nil
	}

	value := strconv.FormatUint(v, 10)
	return &value, nil
}

func prepareFloat(v float64, required bool) (*string, error) {
	if v == 0 {
		if required {
			return nil, fmt.Errorf("field is required, got %f", v)
		}
		return nil, nil
	}
	value := strconv.FormatFloat(v, 'f', -1, 64)
	return &value, nil
}

func prepareString(v string, required bool) (*string, error) {
	if v == "" {
		if required {
			return nil, fmt.Errorf("field is required, got %q", v)
		}
		return nil, nil
	}
	return &v, nil
}

func prepareBool(v bool, required bool) (*string, error) {
	value := strconv.FormatBool(v)
	if !v {
		if required {
			return &value, nil
		}
		return nil, nil
	}

	return &value, nil
}

func prepareList(prefix string, slice interface{}) (url.Values, error) {
	params := url.Values{}
	value := reflect.ValueOf(slice)

	for i := 0; i < value.Len(); i++ {
		ps, err := prepareValues(fmt.Sprintf("%s[%d].", prefix, i), value.Index(i).Interface())
		if err != nil {
			return nil, err
		}

		for k, xs := range ps {
			for _, x := range xs {
				params.Add(k, x)
			}
		}
	}

	return params, nil
}

func prepareMap(prefix string, m interface{}) (url.Values, error) {
	value := url.Values{}
	v := reflect.ValueOf(m)

	for i, key := range v.MapKeys() {
		var keyName string
		var keyValue string

		switch key.Kind() {
		case reflect.String:
			keyName = key.String()
		default:
			return value, fmt.Errorf("only map[string]string are supported (XXX)")
		}

		val := v.MapIndex(key)
		switch val.Kind() {
		case reflect.String:
			keyValue = val.String()
		default:
			return value, fmt.Errorf("only map[string]string are supported (XXX)")
		}

		value.Set(fmt.Sprintf("%s[%d].%s", prefix, i, keyName), keyValue)
	}

	return value, nil
}

func preparePtr(kind reflect.Kind, val reflect.Value, required bool) (*string, error) {
	if val.IsNil() {
		if required {
			return nil, fmt.Errorf("field is required, got empty ptr")
		}
		return nil, nil
	}

	switch kind {
	case reflect.Bool:
		return prepareBool(val.Elem().Bool(), true)
	case reflect.Struct:
		return prepareStruct(val.Interface(), required)
	default:
		return nil, fmt.Errorf("kind %v is not supported as a ptr", kind)
	}
}

func prepareSlice(name string, fieldType reflect.Type, val reflect.Value, required bool) (interface{}, error) {
	switch fieldType.Elem().Kind() {
	case reflect.Uint8:
		switch fieldType {
		case reflect.TypeOf(net.IPv4zero):
			ip := (net.IP)(val.Bytes())
			if ip == nil || ip.Equal(net.IP{}) {
				if required {
					return nil, fmt.Errorf("field is required, got zero IPv4 address")
				}
			} else {
				value := ip.String()
				return &value, nil
			}

		case reflect.TypeOf(MAC48(0, 0, 0, 0, 0, 0)):
			mac := val.Interface().(MACAddress)
			s := mac.String()
			if s == "" {
				if required {
					return nil, fmt.Errorf("field is required, got empty MAC address")
				}
			} else {
				return &s, nil
			}
		default:
			if val.Len() == 0 {
				if required {
					return nil, fmt.Errorf("field is required, got empty slice")
				}
			} else {
				value := base64.StdEncoding.EncodeToString(val.Bytes())
				return &value, nil
			}
		}
	case reflect.String:
		if val.Len() == 0 {
			if required {
				return nil, fmt.Errorf("field is required, got empty slice")
			}
		} else {
			elems := make([]string, 0, val.Len())
			for i := 0; i < val.Len(); i++ {
				// XXX what if the value contains a comma? Double encode?
				s := val.Index(i).String()
				elems = append(elems, s)
			}
			value := strings.Join(elems, ",")
			return &value, nil
		}
	default:
		switch fieldType.Elem() {
		case reflect.TypeOf(CIDR{}), reflect.TypeOf(UUID{}):
			if val.Len() == 0 {
				if required {
					return nil, fmt.Errorf("field is required, got empty slice")
				}
			} else {
				v := reflect.ValueOf(val.Interface())
				ss := make([]string, val.Len())
				for i := 0; i < v.Len(); i++ {
					e := v.Index(i).Interface()
					s, ok := e.(fmt.Stringer)
					if !ok {
						return nil, fmt.Errorf("not a String, %T", e)
					}
					ss[i] = s.String()
				}
				value := strings.Join(ss, ",")
				return &value, nil
			}
		default:
			if val.Len() == 0 {
				if required {
					return nil, fmt.Errorf("field is required, got empty slice")
				}
			} else {
				return prepareList(name, val.Interface())
			}
		}
	}

	return nil, nil
}

func prepareStruct(i interface{}, required bool) (*string, error) {
	s, ok := i.(fmt.Stringer)
	if !ok {
		return nil, fmt.Errorf("struct field not a Stringer")
	}

	if s == nil {
		if required {
			return nil, fmt.Errorf("field is required, got %#v", s)
		}
	}

	return prepareString(s.String(), required)
}

func prepareEmbedStruct(i interface{}) (url.Values, error) {
	return prepareValues("", i)
}

// ExtractJSONTag returns the variable name or defaultName as well as if the field is required (!omitempty)
func ExtractJSONTag(defaultName, jsonTag string) (string, bool) {
	tags := strings.Split(jsonTag, ",")
	name := tags[0]
	required := true
	for _, tag := range tags {
		if tag == "omitempty" {
			required = false
		}
	}

	if name == "" || name == "omitempty" {
		name = defaultName
	}
	return name, required
}
