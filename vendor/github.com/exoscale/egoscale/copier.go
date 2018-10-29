package egoscale

import (
	"fmt"
	"reflect"
)

// Copy copies the value from from into to. The type of "from" must be convertible into the type of "to".
func Copy(to, from interface{}) error {
	tt := reflect.TypeOf(to)
	tv := reflect.ValueOf(to)

	ft := reflect.TypeOf(from)
	fv := reflect.ValueOf(from)

	if tt.Kind() != reflect.Ptr {
		return fmt.Errorf("must copy to a pointer, got %q", tt.Name())
	}

	tt = tt.Elem()
	tv = tv.Elem()

	for {
		if ft.ConvertibleTo(tt) {
			break
		}
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
			fv = fv.Elem()
		} else {
			return fmt.Errorf("cannot convert %q into %q", tt.Name(), ft.Name())
		}
	}

	tv.Set(fv.Convert(tt))
	return nil
}
