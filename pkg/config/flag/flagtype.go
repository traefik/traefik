package flag

import (
	"reflect"
	"strings"

	"github.com/containous/traefik/pkg/config/parser"
)

func getFlagTypes(element interface{}) map[string]reflect.Kind {
	ref := map[string]reflect.Kind{}

	if element == nil {
		return ref
	}

	tp := reflect.TypeOf(element).Elem()

	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)

		if !parser.IsExported(field) {
			continue
		}

		if field.Anonymous {
			addFlagType(ref, "", field.Type)
		} else {
			addFlagType(ref, getName(field.Name), field.Type)
		}
	}

	return ref
}

func addFlagType(ref map[string]reflect.Kind, name string, typ reflect.Type) {
	switch typ.Kind() {
	case reflect.Bool, reflect.Slice:
		ref[name] = typ.Kind()

	case reflect.Map:
		addFlagType(ref, getName(name, parser.MapNamePlaceholder), typ.Elem())

	case reflect.Ptr:
		if typ.Elem().Kind() == reflect.Struct {
			ref[name] = typ.Kind()
		}
		addFlagType(ref, name, typ.Elem())

	case reflect.Struct:
		for j := 0; j < typ.NumField(); j++ {
			subField := typ.Field(j)

			if !parser.IsExported(subField) {
				continue
			}

			if subField.Anonymous {
				addFlagType(ref, getName(name), subField.Type)
			} else {
				addFlagType(ref, getName(name, subField.Name), subField.Type)
			}
		}

	default:
		// noop
	}
}

func getName(names ...string) string {
	return strings.TrimPrefix(strings.ToLower(strings.Join(names, ".")), ".")
}
