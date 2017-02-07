package fun

import (
	"reflect"

	"github.com/BurntSushi/ty"
)

// Set has a parametric type:
//
//	func Set(xs []A) map[A]bool
//
// Set creates a set from a list.
func Set(xs interface{}) interface{} {
	chk := ty.Check(
		new(func([]ty.A) map[ty.A]bool),
		xs)
	vxs, tset := chk.Args[0], chk.Returns[0]

	vtrue := reflect.ValueOf(true)
	vset := reflect.MakeMap(tset)
	xsLen := vxs.Len()
	for i := 0; i < xsLen; i++ {
		vset.SetMapIndex(vxs.Index(i), vtrue)
	}
	return vset.Interface()
}

// Union has a parametric type:
//
//	func Union(a map[A]bool, b map[A]bool) map[A]bool
//
// Union returns the union of two sets, where a set is represented as a
// `map[A]bool`. The sets `a` and `b` are not modified.
func Union(a, b interface{}) interface{} {
	chk := ty.Check(
		new(func(map[ty.A]bool, map[ty.A]bool) map[ty.A]bool),
		a, b)
	va, vb, tc := chk.Args[0], chk.Args[1], chk.Returns[0]

	vtrue := reflect.ValueOf(true)
	vc := reflect.MakeMap(tc)
	for _, vkey := range va.MapKeys() {
		vc.SetMapIndex(vkey, vtrue)
	}
	for _, vkey := range vb.MapKeys() {
		vc.SetMapIndex(vkey, vtrue)
	}
	return vc.Interface()
}

// Intersection has a parametric type:
//
//	func Intersection(a map[A]bool, b map[A]bool) map[A]bool
//
// Intersection returns the intersection of two sets, where a set is
// represented as a `map[A]bool`. The sets `a` and `b` are not modified.
func Intersection(a, b interface{}) interface{} {
	chk := ty.Check(
		new(func(map[ty.A]bool, map[ty.A]bool) map[ty.A]bool),
		a, b)
	va, vb, tc := chk.Args[0], chk.Args[1], chk.Returns[0]

	vtrue := reflect.ValueOf(true)
	vc := reflect.MakeMap(tc)
	for _, vkey := range va.MapKeys() {
		if vb.MapIndex(vkey).IsValid() {
			vc.SetMapIndex(vkey, vtrue)
		}
	}
	for _, vkey := range vb.MapKeys() {
		if va.MapIndex(vkey).IsValid() {
			vc.SetMapIndex(vkey, vtrue)
		}
	}
	return vc.Interface()
}

// Difference has a parametric type:
//
//	func Difference(a map[A]bool, b map[A]bool) map[A]bool
//
// Difference returns a set with all elements in `a` that are not in `b`.
// The sets `a` and `b` are not modified.
func Difference(a, b interface{}) interface{} {
	chk := ty.Check(
		new(func(map[ty.A]bool, map[ty.A]bool) map[ty.A]bool),
		a, b)
	va, vb, tc := chk.Args[0], chk.Args[1], chk.Returns[0]

	vtrue := reflect.ValueOf(true)
	vc := reflect.MakeMap(tc)
	for _, vkey := range va.MapKeys() {
		if !vb.MapIndex(vkey).IsValid() {
			vc.SetMapIndex(vkey, vtrue)
		}
	}
	return vc.Interface()
}
