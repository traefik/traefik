// Package generator implements the custom initialization of all the fields of an empty interface.
package generator

import (
	"reflect"

	"github.com/containous/traefik/v2/pkg/config/parser"
)

type initializer interface {
	SetDefaults()
}

// Generate recursively initializes an empty structure, calling SetDefaults on each field, when it applies.
func Generate(element interface{}) {
	if element == nil {
		return
	}

	generate(element)
}

func generate(element interface{}) {
	field := reflect.ValueOf(element)

	fill(field)
}

func fill(field reflect.Value) {
	switch field.Kind() {
	case reflect.Ptr:
		setPtr(field)
	case reflect.Struct:
		setStruct(field)
	case reflect.Map:
		setMap(field)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.Struct ||
			field.Type().Elem().Kind() == reflect.Ptr && field.Type().Elem().Elem().Kind() == reflect.Struct {
			slice := reflect.MakeSlice(field.Type(), 1, 1)
			field.Set(slice)

			// use Ptr to allow "SetDefaults"
			value := reflect.New(reflect.PtrTo(field.Type().Elem()))
			setPtr(value)

			elem := value.Elem().Elem()
			field.Index(0).Set(elem)
		} else if field.Len() == 0 {
			slice := reflect.MakeSlice(field.Type(), 0, 0)
			field.Set(slice)
		}
	}
}

func setPtr(field reflect.Value) {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	if field.Type().Implements(reflect.TypeOf((*initializer)(nil)).Elem()) {
		method := field.MethodByName("SetDefaults")
		if method.IsValid() {
			method.Call([]reflect.Value{})
		}
	}

	fill(field.Elem())
}

func setStruct(field reflect.Value) {
	for i := 0; i < field.NumField(); i++ {
		fd := field.Field(i)
		structField := field.Type().Field(i)

		if structField.Tag.Get(parser.TagLabel) == "-" {
			continue
		}

		if parser.IsExported(structField) {
			fill(fd)
		}
	}
}

func setMap(field reflect.Value) {
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	ptrValue := reflect.New(reflect.PtrTo(field.Type().Elem()))
	fill(ptrValue)

	value := ptrValue.Elem().Elem()
	key := reflect.ValueOf(parser.MapNamePlaceholder)
	field.SetMapIndex(key, value)
}
