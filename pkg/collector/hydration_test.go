package collector

import (
	"fmt"
	"reflect"
	"time"

	"github.com/traefik/paerser/types"
)

const (
	sliceItemNumber     = 2
	mapItemNumber       = 2
	defaultString       = "foobar"
	defaultNumber       = 42
	defaultBool         = true
	defaultMapKeyPrefix = "name"
)

func hydrate(element interface{}) error {
	field := reflect.ValueOf(element)
	return fill(field)
}

func fill(field reflect.Value) error {
	switch field.Kind() {
	case reflect.Struct:
		if err := setStruct(field); err != nil {
			return err
		}
	case reflect.Ptr:
		if err := setPointer(field); err != nil {
			return err
		}
	case reflect.Slice:
		if err := setSlice(field); err != nil {
			return err
		}
	case reflect.Map:
		if err := setMap(field); err != nil {
			return err
		}
	case reflect.Interface:
		if err := fill(field.Elem()); err != nil {
			return err
		}
	case reflect.String:
		setTyped(field, defaultString)
	case reflect.Int:
		setTyped(field, defaultNumber)
	case reflect.Int8:
		setTyped(field, int8(defaultNumber))
	case reflect.Int16:
		setTyped(field, int16(defaultNumber))
	case reflect.Int32:
		setTyped(field, int32(defaultNumber))
	case reflect.Int64:
		switch field.Type() {
		case reflect.TypeOf(types.Duration(time.Second)):
			setTyped(field, int64(defaultNumber*int(time.Second)))
		default:
			setTyped(field, int64(defaultNumber))
		}
	case reflect.Uint:
		setTyped(field, uint(defaultNumber))
	case reflect.Uint8:
		setTyped(field, uint8(defaultNumber))
	case reflect.Uint16:
		setTyped(field, uint16(defaultNumber))
	case reflect.Uint32:
		setTyped(field, uint32(defaultNumber))
	case reflect.Uint64:
		setTyped(field, uint64(defaultNumber))
	case reflect.Bool:
		setTyped(field, defaultBool)
	case reflect.Float32:
		setTyped(field, float32(defaultNumber))
	case reflect.Float64:
		setTyped(field, float64(defaultNumber))
	}

	return nil
}

func setTyped(field reflect.Value, i interface{}) {
	baseValue := reflect.ValueOf(i)
	if field.Kind().String() == field.Type().String() {
		field.Set(baseValue)
	} else {
		field.Set(baseValue.Convert(field.Type()))
	}
}

func setMap(field reflect.Value) error {
	field.Set(reflect.MakeMap(field.Type()))

	for i := 0; i < mapItemNumber; i++ {
		baseKeyName := makeKeyName(field.Type().Elem())
		key := reflect.ValueOf(fmt.Sprintf("%s%d", baseKeyName, i))

		// generate value
		ptrType := reflect.PtrTo(field.Type().Elem())
		ptrValue := reflect.New(ptrType)
		if err := fill(ptrValue); err != nil {
			return err
		}
		value := ptrValue.Elem().Elem()

		field.SetMapIndex(key, value)
	}
	return nil
}

func makeKeyName(typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.Ptr:
		return typ.Elem().Name()
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Bool, reflect.Float32, reflect.Float64:
		return defaultMapKeyPrefix
	default:
		return typ.Name()
	}
}

func setStruct(field reflect.Value) error {
	for i := 0; i < field.NumField(); i++ {
		fld := field.Field(i)
		stFld := field.Type().Field(i)

		if !stFld.IsExported() || fld.Kind() == reflect.Func {
			continue
		}

		if err := fill(fld); err != nil {
			return err
		}
	}
	return nil
}

func setSlice(field reflect.Value) error {
	field.Set(reflect.MakeSlice(field.Type(), sliceItemNumber, sliceItemNumber))
	for j := 0; j < field.Len(); j++ {
		if err := fill(field.Index(j)); err != nil {
			return err
		}
	}
	return nil
}

func setPointer(field reflect.Value) error {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
		if err := fill(field.Elem()); err != nil {
			return err
		}
	} else {
		if err := fill(field.Elem()); err != nil {
			return err
		}
	}
	return nil
}
