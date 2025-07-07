package redactor

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/copystructure"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
	"mvdan.cc/xurls/v2"
)

const (
	maskShort   = "xxxx"
	maskLarge   = maskShort + maskShort + maskShort + maskShort + maskShort + maskShort + maskShort + maskShort
	tagLoggable = "loggable"
	tagExport   = "export"
)

// Anonymize redacts the configuration fields that do not have an export=true struct tag.
// It returns the resulting marshaled configuration.
func Anonymize(baseConfig interface{}) (string, error) {
	return anonymize(baseConfig, false)
}

func anonymize(baseConfig interface{}, indent bool) (string, error) {
	conf, err := do(baseConfig, tagExport, true, indent)
	if err != nil {
		return "", err
	}
	return doOnJSON(conf), nil
}

// RemoveCredentials redacts the configuration fields that have a loggable=false struct tag.
// It returns the resulting marshaled configuration.
func RemoveCredentials(baseConfig interface{}) (string, error) {
	return removeCredentials(baseConfig, false)
}

func removeCredentials(baseConfig interface{}, indent bool) (string, error) {
	return do(baseConfig, tagLoggable, false, indent)
}

// do marshals the given configuration, while redacting some of the fields
// respectively to the given tag.
func do(baseConfig interface{}, tag string, redactByDefault, indent bool) (string, error) {
	anomConfig, err := copystructure.Copy(baseConfig)
	if err != nil {
		return "", err
	}

	val := reflect.ValueOf(anomConfig)

	err = doOnStruct(val, tag, redactByDefault)
	if err != nil {
		return "", err
	}

	configJSON, err := marshal(anomConfig, indent)
	if err != nil {
		return "", err
	}

	return string(configJSON), nil
}

func doOnJSON(input string) string {
	return xurls.Relaxed().ReplaceAllString(input, maskLarge)
}

func doOnStruct(field reflect.Value, tag string, redactByDefault bool) error {
	if field.Type().AssignableTo(reflect.TypeOf(dynamic.PluginConf{})) {
		resetPlugin(field)
		return nil
	}

	switch field.Kind() {
	case reflect.Ptr:
		if !field.IsNil() {
			if err := doOnStruct(field.Elem(), tag, redactByDefault); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := range field.NumField() {
			fld := field.Field(i)
			stField := field.Type().Field(i)
			if !isExported(stField) {
				continue
			}

			if stField.Tag.Get(tag) == "false" || stField.Tag.Get(tag) != "true" && redactByDefault {
				if err := reset(fld, stField.Name); err != nil {
					return err
				}
				continue
			}

			// A struct field cannot be set it must be filled as pointer.
			if fld.Kind() == reflect.Struct {
				fldPtr := reflect.New(fld.Type())
				fldPtr.Elem().Set(fld)

				if err := doOnStruct(fldPtr, tag, redactByDefault); err != nil {
					return err
				}

				fld.Set(fldPtr.Elem())

				continue
			}

			if err := doOnStruct(fld, tag, redactByDefault); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range field.MapKeys() {
			val := field.MapIndex(key)

			// A struct value cannot be set it must be filled as pointer.
			if val.Kind() == reflect.Struct {
				valPtr := reflect.New(val.Type())
				valPtr.Elem().Set(val)

				if err := doOnStruct(valPtr, tag, redactByDefault); err != nil {
					return err
				}

				field.SetMapIndex(key, valPtr.Elem())

				continue
			}

			if err := doOnStruct(val, tag, redactByDefault); err != nil {
				return err
			}
		}
	case reflect.Slice:
		for j := range field.Len() {
			if err := doOnStruct(field.Index(j), tag, redactByDefault); err != nil {
				return err
			}
		}
	}

	return nil
}

func reset(field reflect.Value, name string) error {
	if !field.CanSet() {
		return fmt.Errorf("cannot reset field %s", name)
	}

	switch field.Kind() {
	case reflect.Ptr:
		if !field.IsNil() {
			field.Set(reflect.Zero(field.Type()))
		}
	case reflect.Struct:
		if field.IsValid() {
			field.Set(reflect.Zero(field.Type()))
		}
	case reflect.String:
		if field.String() != "" {
			if field.Type().AssignableTo(reflect.TypeOf(types.FileOrContent(""))) {
				field.Set(reflect.ValueOf(types.FileOrContent(maskShort)))
			} else {
				field.Set(reflect.ValueOf(maskShort))
			}
		}
	case reflect.Map:
		if field.Len() > 0 {
			field.Set(reflect.MakeMap(field.Type()))
		}
	case reflect.Slice:
		if field.Len() > 0 {
			switch field.Type().Elem().Kind() {
			case reflect.String:
				slice := reflect.MakeSlice(field.Type(), field.Len(), field.Len())
				for j := range field.Len() {
					slice.Index(j).SetString(maskShort)
				}
				field.Set(slice)
			default:
				field.Set(reflect.MakeSlice(field.Type(), 0, 0))
			}
		}
	case reflect.Interface:
		return fmt.Errorf("reset not supported for interface type (for %s field)", name)
	default:
		// Primitive type
		field.Set(reflect.Zero(field.Type()))
	}
	return nil
}

// resetPlugin resets the plugin configuration so it keep the plugin name but not its configuration.
func resetPlugin(field reflect.Value) {
	for _, key := range field.MapKeys() {
		field.SetMapIndex(key, reflect.ValueOf(struct{}{}))
	}
}

// isExported return true is a struct field is exported, else false.
func isExported(f reflect.StructField) bool {
	if f.PkgPath != "" && !f.Anonymous {
		return false
	}
	return true
}

func marshal(anomConfig interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(anomConfig, "", "  ")
	}
	return json.Marshal(anomConfig)
}
