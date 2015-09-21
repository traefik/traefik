package fun

import (
	"reflect"

	"github.com/BurntSushi/ty"
)

// Memo has a parametric type:
//
//	func Memo(f func(A) B) func(A) B
//
// Memo memoizes any function of a single argument that returns a single value.
// The type `A` must be a Go type for which the comparison operators `==` and
// `!=` are fully defined (this rules out functions, maps and slices).
func Memo(f interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A) ty.B)),
		f)
	vf := chk.Args[0]

	saved := make(map[interface{}]reflect.Value)
	memo := func(in []reflect.Value) []reflect.Value {
		val := in[0].Interface()
		ret, ok := saved[val]
		if ok {
			return []reflect.Value{ret}
		}

		ret = call1(vf, in[0])
		saved[val] = ret
		return []reflect.Value{ret}
	}
	return reflect.MakeFunc(vf.Type(), memo).Interface()
}
