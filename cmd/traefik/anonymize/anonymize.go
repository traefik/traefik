package anonymize

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/copystructure"
	"github.com/mvdan/xurls"
)

const (
	maskShort = "xxxx"
	maskLarge = maskShort + maskShort + maskShort + maskShort + maskShort + maskShort + maskShort + maskShort
)

// Do configuration.
func Do(baseConfig interface{}, indent bool) (string, error) {
	anomConfig, err := copystructure.Copy(baseConfig)
	if err != nil {
		return "", err
	}

	val := reflect.ValueOf(anomConfig)

	err = doOnStruct(val)
	if err != nil {
		return "", err
	}

	configJSON, err := marshal(anomConfig, indent)
	if err != nil {
		return "", err
	}

	return doOnJSON(string(configJSON)), nil
}

func doOnJSON(input string) string {
	mailExp := regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}"`)
	return xurls.Relaxed.ReplaceAllString(mailExp.ReplaceAllString(input, maskLarge+"\""), maskLarge)
}

func doOnStruct(field reflect.Value) error {
	if reflect.Ptr == field.Kind() && !field.IsNil() {
		if err := doOnStruct(field.Elem()); err != nil {
			return err
		}
	} else if reflect.Struct == field.Kind() {
		for i := 0; i < field.NumField(); i++ {
			fld := field.Field(i)
			stField := field.Type().Field(i)
			if !isExported(stField.Name) {
				continue
			}
			if stField.Tag.Get("export") == "true" {
				if err := doOnStruct(fld); err != nil {
					return err
				}
			} else {
				if err := reset(fld, stField.Name); err != nil {
					return err
				}
			}
		}
	} else if reflect.Map == field.Kind() {
		for _, key := range field.MapKeys() {
			if err := doOnStruct(field.MapIndex(key)); err != nil {
				return err
			}
		}
	} else if reflect.Slice == field.Kind() {
		for j := 0; j < field.Len(); j++ {
			if err := doOnStruct(field.Index(j)); err != nil {
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
	if reflect.Ptr == field.Kind() {
		if !field.IsNil() {
			field.Set(reflect.Zero(field.Type()))
		}
	} else if reflect.String == field.Kind() {
		if field.String() != "" {
			field.Set(reflect.ValueOf(maskShort))
		}
	} else if reflect.Struct == field.Kind() {
		if field.IsValid() {
			field.Set(reflect.Zero(field.Type()))
		}
	} else if reflect.Map == field.Kind() {
		if field.Len() > 0 {
			field.Set(reflect.MakeMap(field.Type()))
		}
	} else if reflect.Slice == field.Kind() {
		if field.Len() > 0 {
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		}
	} else if reflect.Interface == field.Kind() {
		if !field.IsNil() {
			return reset(field.Elem(), "")
		}
	} else {
		// Primitive type
		field.Set(reflect.Zero(field.Type()))
	}
	return nil
}

// isExported return true is the field (from fieldName) is exported, else false
func isExported(fieldName string) bool {
	if len(fieldName) < 1 {
		return false
	}
	return string(fieldName[0]) == strings.ToUpper(string(fieldName[0]))
}

func marshal(anomConfig interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(anomConfig, "", " ")
	}
	return json.Marshal(anomConfig)
}
