package fun

import (
	"reflect"
)

func zeroValue(typ reflect.Type) reflect.Value {
	return reflect.New(typ).Elem()
}

type swapper reflect.Value

func swapperOf(typ reflect.Type) swapper {
	return swapper(zeroValue(typ))
}

func (s swapper) swap(a, b reflect.Value) {
	vs := reflect.Value(s)
	vs.Set(a)
	a.Set(b)
	b.Set(vs)
}

func call(f reflect.Value, args ...reflect.Value) {
	f.Call(args)
}

func call1(f reflect.Value, args ...reflect.Value) reflect.Value {
	return f.Call(args)[0]
}

func call2(f reflect.Value, args ...reflect.Value) (
	reflect.Value, reflect.Value) {

	ret := f.Call(args)
	return ret[0], ret[1]
}
