package ty

import (
	"fmt"
	"reflect"
	"strings"
)

// TypeError corresponds to any error reported by the `Check` function.
// Since `Check` panics, if you want to run `Check` safely, it is
// appropriate to recover and use a type switch to discover a `TypeError`
// value.
type TypeError string

func (te TypeError) Error() string {
	return string(te)
}

func pe(format string, v ...interface{}) TypeError {
	return TypeError(fmt.Sprintf(format, v...))
}

func ppe(format string, v ...interface{}) {
	panic(pe(format, v...))
}

// Typed corresponds to the information returned by `Check`.
type Typed struct {
	// In correspondence with the `as` parameter to `Check`.
	Args []reflect.Value

	// In correspondence with the return types of `f` in `Check`.
	Returns []reflect.Type

	// The type environment generated via unification in `Check`.
	// (Its usefulness in the public API is questionable.)
	TypeEnv map[string]reflect.Type
}

// Check accepts a function `f`, which may have a parametric type, along with a
// number of arguments in correspondence with the arguments to `f`,
// and returns inferred Go type information. This type information includes
// a list of `reflect.Value` in correspondence with `as`, a list of
// `reflect.Type` in correspondence with the return types of `f` and a type
// environment mapping type variables to `reflect.Type`.
//
// The power of `Check` comes from the following invariant: if `Check` returns,
// then the types of the arguments corresponding to `as` are consistent
// with the parametric type of `f`, *and* the parametric return types of `f`
// were made into valid Go types that are not parametric. Otherwise, there is
// a bug in `Check`.
//
// More concretely, consider a simple parametric function `Map`, which
// transforms a list of elements by applying a function to each element in
// order to generate a new list. Such a function constructed only for integers
// might have a type like
//
//	func Map(func(int) int, []int) []int
//
// But the parametric type of `Map` could be given with
//
//	func Map(func(A) B, []A) []B
//
// which in English reads, "Given a function from any type `A` to any type `B`
// and a slice of `A`, `Map` returns a slice of `B`."
//
// To write a parametric function like `Map`, one can pass a pointer
// to a nil function of the desired parametric type to get the reflection
// information:
//
//	func Map(f, xs interface{}) interface{} {
//		// Given the parametric type and the arguments, Check will
//		// return all the reflection information you need to write `Map`.
//		uni := ty.Check(
//			new(func(func(ty.A) ty.B, []ty.A) []ty.B),
//			f, xs)
//
//		// `vf` and `vxs` are `reflect.Value`s of `f` and `xs`.
//		vf, vxs := uni.Args[0], uni.Args[1]
//
//		// `tys` is a `reflect.Type` of `[]ty.B` where `ty.B` is replaced
//		// with the return type of the given function `f`.
//		tys := uni.Returns[0]
//
//		// Given the promise of `Check`, we now know that `vf` has
//		// type `func(ty.A) ty.B` and `vxs` has type `[]ty.A`.
//		xsLen := vxs.Len()
//
//		// Constructs a new slice which will have type `[]ty.B`.
//		vys := reflect.MakeSlice(tys, xsLen, xsLen)
//
//		// Actually perform the `Map` operation, but in the world of
//		// reflection.
//		for i := 0; i < xsLen; i++ {
//			vy := vf.Call([]reflect.Value{vxs.Index(i)})[0]
//			vys.Index(i).Set(vy)
//		}
//
//		// The `reflect.Value.Interface` method is how we exit the world of
//		// reflection. The onus is now on the caller to type assert it to
//		// the appropriate type.
//		return vys.Interface()
//	}
//
// Working in the reflection world is certainly more inconvenient than writing
// regular Go code, but the information and invariants held by `Check` provide
// a more convenient experience than how one normally works with reflection.
// (Notice that there is no error-prone type switching or boiler plate to
// construct new types, since `Check` guarantees the types are consistent
// with the inputs for us.)
//
// And while writing such functions is still not so convenient,
// invoking them is simple:
//
//	square := func(x int) int { return x * x }
//	squared := Map(square, []int{1, 2, 3, 4, 5}).([]int)
//
// Restrictions
//
// There are a few restrictions imposed on the parametric return types of
// `f`: type variables may only be found in types that can be composed by the
// `reflect` package. This *only* includes channels, maps, pointers and slices.
// If a type variable is found in an array, function or struct, `Check` will
// panic.
//
// Also, type variables inside of structs are ignored in the types of the
// arguments `as`. This restriction may be lifted in the future.
//
// To be clear: type variables *may* appear in arrays or functions in the types
// of the arguments `as`.
func Check(f interface{}, as ...interface{}) *Typed {
	rf := reflect.ValueOf(f)
	tf := rf.Type()

	if tf.Kind() == reflect.Ptr {
		rf = reflect.Indirect(rf)
		tf = rf.Type()
	}
	if tf.Kind() != reflect.Func {
		ppe("The type of `f` must be a function, but it is a '%s'.", tf.Kind())
	}
	if tf.NumIn() != len(as) {
		ppe("`f` expects %d arguments, but only %d were given.",
			tf.NumIn(), len(as))
	}

	// Populate the argument value list.
	args := make([]reflect.Value, len(as))
	for i := 0; i < len(as); i++ {
		args[i] = reflect.ValueOf(as[i])
	}

	// Populate our type variable environment through unification.
	tyenv := make(tyenv)
	for i := 0; i < len(args); i++ {
		tp := typePair{tyenv, tf.In(i), args[i].Type()}

		// Mutates the type variable environment.
		if err := tp.unify(tp.param, tp.input); err != nil {
			argTypes := make([]string, len(args))
			for i := range args {
				argTypes[i] = args[i].Type().String()
			}
			ppe("\nError type checking\n\t%s\nwith argument types\n\t(%s)\n%s",
				tf, strings.Join(argTypes, ", "), err)
		}
	}

	// Now substitute those types into the return types of `f`.
	retTypes := make([]reflect.Type, tf.NumOut())
	for i := 0; i < tf.NumOut(); i++ {
		retTypes[i] = (&returnType{tyenv, tf.Out(i)}).tysubst(tf.Out(i))
	}
	return &Typed{args, retTypes, map[string]reflect.Type(tyenv)}
}

// tyenv maps type variable names to their inferred Go type.
type tyenv map[string]reflect.Type

// typePair represents a pair of types to be unified. They act as a way to
// report sensible error messages from within the unification algorithm.
//
// It also includes a type environment, which is mutated during unification.
type typePair struct {
	tyenv tyenv
	param reflect.Type
	input reflect.Type
}

func (tp typePair) error(format string, v ...interface{}) error {
	return pe("Type error when unifying type '%s' and '%s': %s",
		tp.param, tp.input, fmt.Sprintf(format, v...))
}

// unify attempts to satisfy a pair of types, where the `param` type is the
// expected type of a function argument and the `input` type is the known
// type of a function argument. The `param` type may be parametric (that is,
// it may contain a type that is convertible to TypeVariable) but the
// `input` type may *not* be parametric.
//
// Any failure to unify the two types results in a panic.
//
// The end result of unification is a type environment: a set of substitutions
// from type variable to a Go type.
func (tp typePair) unify(param, input reflect.Type) error {
	if tyname := tyvarName(input); len(tyname) > 0 {
		return tp.error("Type variables are not allowed in the types of " +
			"arguments.")
	}
	if tyname := tyvarName(param); len(tyname) > 0 {
		if cur, ok := tp.tyenv[tyname]; ok && cur != input {
			return tp.error("Type variable %s expected type '%s' but got '%s'.",
				tyname, cur, input)
		} else if !ok {
			tp.tyenv[tyname] = input
		}
		return nil
	}
	if param.Kind() != input.Kind() {
		return tp.error("Cannot unify different kinds of types '%s' and '%s'.",
			param, input)
	}

	switch param.Kind() {
	case reflect.Array:
		return tp.unify(param.Elem(), input.Elem())
	case reflect.Chan:
		if param.ChanDir() != input.ChanDir() {
			return tp.error("Cannot unify '%s' with '%s' "+
				"(channel directions are different: '%s' != '%s').",
				param, input, param.ChanDir(), input.ChanDir())
		}
		return tp.unify(param.Elem(), input.Elem())
	case reflect.Func:
		if param.NumIn() != input.NumIn() || param.NumOut() != input.NumOut() {
			return tp.error("Cannot unify '%s' with '%s'.", param, input)
		}
		for i := 0; i < param.NumIn(); i++ {
			if err := tp.unify(param.In(i), input.In(i)); err != nil {
				return err
			}
		}
		for i := 0; i < param.NumOut(); i++ {
			if err := tp.unify(param.Out(i), input.Out(i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		if err := tp.unify(param.Key(), input.Key()); err != nil {
			return err
		}
		return tp.unify(param.Elem(), input.Elem())
	case reflect.Ptr:
		return tp.unify(param.Elem(), input.Elem())
	case reflect.Slice:
		return tp.unify(param.Elem(), input.Elem())
	}

	// The only other container types are Interface and Struct.
	// I am unsure about what to do with interfaces. Mind is fuzzy.
	// Structs? I don't think it really makes much sense to use type
	// variables inside of them.
	return nil
}

// returnType corresponds to the type of a single return value of a function,
// in which the type may be parametric. It also contains a type environment
// constructed from unification.
type returnType struct {
	tyenv tyenv
	typ   reflect.Type
}

func (rt returnType) panic(format string, v ...interface{}) {
	ppe("Error substituting in return type '%s': %s",
		rt.typ, fmt.Sprintf(format, v...))
}

// tysubst attempts to substitute all type variables within a single return
// type with their corresponding Go type from the type environment.
//
// tysubst will panic if a type variable is unbound, or if it encounters a
// type that cannot be dynamically created. Such types include arrays,
// functions and structs. (A limitation of the `reflect` package.)
func (rt returnType) tysubst(typ reflect.Type) reflect.Type {
	if tyname := tyvarName(typ); len(tyname) > 0 {
		if thetype, ok := rt.tyenv[tyname]; !ok {
			rt.panic("Unbound type variable %s.", tyname)
		} else {
			return thetype
		}
	}

	switch typ.Kind() {
	case reflect.Array:
		rt.panic("Cannot dynamically create Array types.")
	case reflect.Chan:
		return reflect.ChanOf(typ.ChanDir(), rt.tysubst(typ.Elem()))
	case reflect.Func:
		rt.panic("Cannot dynamically create Function types.")
	case reflect.Interface:
		// rt.panic("TODO")
		// Not sure if this is right.
		return typ
	case reflect.Map:
		return reflect.MapOf(rt.tysubst(typ.Key()), rt.tysubst(typ.Elem()))
	case reflect.Ptr:
		return reflect.PtrTo(rt.tysubst(typ.Elem()))
	case reflect.Slice:
		return reflect.SliceOf(rt.tysubst(typ.Elem()))
	case reflect.Struct:
		rt.panic("Cannot dynamically create Struct types.")
	case reflect.UnsafePointer:
		rt.panic("Cannot dynamically create unsafe.Pointer types.")
	}

	// We've covered all the composite types, so we're only left with
	// base types.
	return typ
}

func tyvarName(t reflect.Type) string {
	if !t.ConvertibleTo(tyvarUnderlyingType) {
		return ""
	}
	return t.Name()
}

// AssertType panics with a `TypeError` if `v` does not have type `t`.
// Otherwise, it returns the `reflect.Value` of `v`.
func AssertType(v interface{}, t reflect.Type) reflect.Value {
	rv := reflect.ValueOf(v)
	tv := rv.Type()
	if tv != t {
		ppe("Value '%v' has type '%s' but expected '%s'.", v, tv, t)
	}
	return rv
}
