package ty

import (
	"reflect"
)

// TypeVariable is the underlying type of every type variable used in
// parametric types. It should not be used directly. Instead, use
//
//	type myOwnTypeVariable TypeVariable
//
// to create your own type variable. For your convenience, this package
// defines some type variables for you. (e.g., `A`, `B`, `C`, ...)
type TypeVariable struct {
	noImitation struct{}
}

// tyvarUnderlyingType is used to discover types that are type variables.
// Namely, any type variable must be convertible to `TypeVariable`.
var tyvarUnderlyingType = reflect.TypeOf(TypeVariable{})

type A TypeVariable
type B TypeVariable
type C TypeVariable
type D TypeVariable
type E TypeVariable
type F TypeVariable
type G TypeVariable
