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
func prepareValues(prefix string, params url.Values, command interface{}) error {
	value := reflect.ValueOf(command)
	typeof := reflect.TypeOf(command)

	// Going up the pointer chain to find the underlying struct
	for typeof.Kind() == reflect.Ptr {
		typeof = typeof.Elem()
		value = value.Elem()
	}

	// Checking for nil commands
	if !value.IsValid() {
		return fmt.Errorf("cannot serialize the invalid value %#v", command)
	}

	for i := 0; i < typeof.NumField(); i++ {
		field := typeof.Field(i)
		if field.Name == "_" {
			continue
		}

		val := value.Field(i)
		tag := field.Tag
		if json, ok := tag.Lookup("json"); ok {
			n, required := ExtractJSONTag(field.Name, json)
			name := prefix + n

			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				v := val.Int()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), n, val.Kind())
					}
				} else {
					params.Set(name, strconv.FormatInt(v, 10))
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				v := val.Uint()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), n, val.Kind())
					}
				} else {
					params.Set(name, strconv.FormatUint(v, 10))
				}
			case reflect.Float32, reflect.Float64:
				v := val.Float()
				if v == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got 0", typeof.Name(), n, val.Kind())
					}
				} else {
					params.Set(name, strconv.FormatFloat(v, 'f', -1, 64))
				}
			case reflect.String:
				v := val.String()
				if v == "" {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got \"\"", typeof.Name(), n, val.Kind())
					}
				} else {
					params.Set(name, v)
				}
			case reflect.Bool:
				v := val.Bool()
				if !v {
					if required {
						params.Set(name, "false")
					}
				} else {
					params.Set(name, "true")
				}
			case reflect.Ptr:
				if val.IsNil() {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got empty ptr", typeof.Name(), n, val.Kind())
					}
				} else {
					switch field.Type.Elem().Kind() {
					case reflect.Bool:
						params.Set(name, strconv.FormatBool(val.Elem().Bool()))
					case reflect.Struct:
						i := val.Interface()
						s, ok := i.(fmt.Stringer)
						if !ok {
							return fmt.Errorf("%s.%s (%v) is not a Stringer", typeof.Name(), field.Name, val.Kind())
						}
						if s != nil && s.String() == "" {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty value", typeof.Name(), field.Name, val.Kind())
							}
						} else {
							params.Set(n, s.String())
						}
					default:
						log.Printf("[SKIP] %s.%s (%v) not supported", typeof.Name(), n, field.Type.Elem().Kind())
					}
				}
			case reflect.Slice:
				switch field.Type.Elem().Kind() {
				case reflect.Uint8:
					switch field.Type {
					case reflect.TypeOf(net.IPv4zero):
						ip := (net.IP)(val.Bytes())
						if ip == nil || ip.Equal(net.IPv4zero) {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got zero IPv4 address", typeof.Name(), n, val.Kind())
							}
						} else {
							params.Set(name, ip.String())
						}
					case reflect.TypeOf(MAC48(0, 0, 0, 0, 0, 0)):
						mac := val.Interface().(MACAddress)
						s := mac.String()
						if s == "" {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty MAC address", typeof.Name(), field.Name, val.Kind())
							}
						} else {
							params.Set(name, s)
						}
					default:
						if val.Len() == 0 {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), n, val.Kind())
							}
						} else {
							v := val.Bytes()
							params.Set(name, base64.StdEncoding.EncodeToString(v))
						}
					}
				case reflect.String:
					if val.Len() == 0 {
						if required {
							return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), n, val.Kind())
						}
					} else {
						elems := make([]string, 0, val.Len())
						for i := 0; i < val.Len(); i++ {
							// XXX what if the value contains a comma? Double encode?
							s := val.Index(i).String()
							elems = append(elems, s)
						}
						params.Set(name, strings.Join(elems, ","))
					}
				default:
					switch field.Type.Elem() {
					case reflect.TypeOf(CIDR{}), reflect.TypeOf(UUID{}):
						if val.Len() == 0 {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), n, val.Kind())
							}
						} else {
							value := reflect.ValueOf(val.Interface())
							ss := make([]string, val.Len())
							for i := 0; i < value.Len(); i++ {
								v := value.Index(i).Interface()
								s, ok := v.(fmt.Stringer)
								if !ok {
									return fmt.Errorf("not a String, %T", v)
								}
								ss[i] = s.String()
							}
							params.Set(name, strings.Join(ss, ","))
						}
					default:
						if val.Len() == 0 {
							if required {
								return fmt.Errorf("%s.%s (%v) is required, got empty slice", typeof.Name(), n, val.Kind())
							}
						} else {
							err := prepareList(name, params, val.Interface())
							if err != nil {
								return err
							}
						}
					}
				}
			case reflect.Map:
				if val.Len() == 0 {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got empty map", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					err := prepareMap(name, params, val.Interface())
					if err != nil {
						return err
					}
				}
			case reflect.Struct:
				i := val.Interface()
				s, ok := i.(fmt.Stringer)
				if !ok {
					return fmt.Errorf("%s.%s (%v) is not a Stringer", typeof.Name(), field.Name, val.Kind())
				}
				if s != nil && s.String() == "" {
					if required {
						return fmt.Errorf("%s.%s (%v) is required, got empty value", typeof.Name(), field.Name, val.Kind())
					}
				} else {
					params.Set(n, s.String())
				}
			default:
				if required {
					return fmt.Errorf("unsupported type %s.%s (%v)", typeof.Name(), n, val.Kind())
				}
				fmt.Printf("%s\n", val.Kind())
			}
		} else {
			log.Printf("[SKIP] %s.%s no json label found", typeof.Name(), field.Name)
		}
	}

	return nil
}

func prepareList(prefix string, params url.Values, slice interface{}) error {
	value := reflect.ValueOf(slice)

	for i := 0; i < value.Len(); i++ {
		err := prepareValues(fmt.Sprintf("%s[%d].", prefix, i), params, value.Index(i).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareMap(prefix string, params url.Values, m interface{}) error {
	value := reflect.ValueOf(m)

	for i, key := range value.MapKeys() {
		var keyName string
		var keyValue string

		switch key.Kind() {
		case reflect.String:
			keyName = key.String()
		default:
			return fmt.Errorf("only map[string]string are supported (XXX)")
		}

		val := value.MapIndex(key)
		switch val.Kind() {
		case reflect.String:
			keyValue = val.String()
		default:
			return fmt.Errorf("only map[string]string are supported (XXX)")
		}
		params.Set(fmt.Sprintf("%s[%d].%s", prefix, i, keyName), keyValue)
	}
	return nil
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
