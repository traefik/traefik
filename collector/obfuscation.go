package collector

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

// Obfuscate configuration.
func Obfuscate(baseConfig interface{}, indent bool) (string, error) {
	obfuscatedConfig, err := copystructure.Copy(baseConfig)
	if err != nil {
		return "", err
	}

	val := reflect.ValueOf(obfuscatedConfig)

	err = obfuscateStruct(val)
	if err != nil {
		return "", err
	}

	configJSON, err := marshal(obfuscatedConfig, indent)
	if err != nil {
		return "", err
	}

	return obfuscateJSON(string(configJSON)), nil
}

func obfuscateJSON(input string) string {
	mailExp := regexp.MustCompile(`\w[-._\w]*\w@\w[-._\w]*\w\.\w{2,3}"`)
	return xurls.Relaxed.ReplaceAllString(mailExp.ReplaceAllString(input, maskLarge+"\""), maskLarge)
}

func obfuscateStruct(field reflect.Value) error {
	if reflect.Ptr == field.Kind() && !field.IsNil() {
		if err := obfuscateStruct(field.Elem()); err != nil {
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
				if err := obfuscateStruct(fld); err != nil {
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
			if err := obfuscateStruct(field.MapIndex(key)); err != nil {
				return err
			}
		}
	} else if reflect.Slice == field.Kind() {
		for j := 0; j < field.Len(); j++ {
			if err := obfuscateStruct(field.Index(j)); err != nil {
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

func marshal(obfuscatedConfig interface{}, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(obfuscatedConfig, "", " ")
	}
	return json.Marshal(obfuscatedConfig)
}
