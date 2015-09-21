package fun

import (
	"reflect"
	"runtime"
	"sync"

	"github.com/BurntSushi/ty"
)

// All has a parametric type:
//
//	func All(p func(A) bool, xs []A) bool
//
// All returns `true` if and only if every element in `xs` satisfies `p`.
func All(f, xs interface{}) bool {
	chk := ty.Check(
		new(func(func(ty.A) bool, []ty.A) bool),
		f, xs)
	vf, vxs := chk.Args[0], chk.Args[1]

	xsLen := vxs.Len()
	for i := 0; i < xsLen; i++ {
		if !call1(vf, vxs.Index(i)).Interface().(bool) {
			return false
		}
	}
	return true
}

// Exists has a parametric type:
//
//	func Exists(p func(A) bool, xs []A) bool
//
// Exists returns `true` if and only if an element in `xs` satisfies `p`.
func Exists(f, xs interface{}) bool {
	chk := ty.Check(
		new(func(func(ty.A) bool, []ty.A) bool),
		f, xs)
	vf, vxs := chk.Args[0], chk.Args[1]

	xsLen := vxs.Len()
	for i := 0; i < xsLen; i++ {
		if call1(vf, vxs.Index(i)).Interface().(bool) {
			return true
		}
	}
	return false
}

// In has a parametric type:
//
//	func In(needle A, haystack []A) bool
//
// In returns `true` if and only if `v` can be found in `xs`. The equality test
// used is Go's standard `==` equality and NOT deep equality.
//
// Note that this requires that `A` be a type that can be meaningfully compared.
func In(needle, haystack interface{}) bool {
	chk := ty.Check(
		new(func(ty.A, []ty.A) bool),
		needle, haystack)
	vhaystack := chk.Args[1]

	length := vhaystack.Len()
	for i := 0; i < length; i++ {
		if vhaystack.Index(i).Interface() == needle {
			return true
		}
	}
	return false
}

// Map has a parametric type:
//
//	func Map(f func(A) B, xs []A) []B
//
// Map returns the list corresponding to the return value of applying
// `f` to each element in `xs`.
func Map(f, xs interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A) ty.B, []ty.A) []ty.B),
		f, xs)
	vf, vxs, tys := chk.Args[0], chk.Args[1], chk.Returns[0]

	xsLen := vxs.Len()
	vys := reflect.MakeSlice(tys, xsLen, xsLen)
	for i := 0; i < xsLen; i++ {
		vy := call1(vf, vxs.Index(i))
		vys.Index(i).Set(vy)
	}
	return vys.Interface()
}

// Filter has a parametric type:
//
//	func Filter(p func(A) bool, xs []A) []A
//
// Filter returns a new list only containing the elements of `xs` that satisfy
// the predicate `p`.
func Filter(p, xs interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A) bool, []ty.A) []ty.A),
		p, xs)
	vp, vxs, tys := chk.Args[0], chk.Args[1], chk.Returns[0]

	xsLen := vxs.Len()
	vys := reflect.MakeSlice(tys, 0, xsLen)
	for i := 0; i < xsLen; i++ {
		vx := vxs.Index(i)
		if call1(vp, vx).Bool() {
			vys = reflect.Append(vys, vx)
		}
	}
	return vys.Interface()
}

// Foldl has a parametric type:
//
//	func Foldl(f func(A, B) B, init B, xs []A) B
//
// Foldl reduces a list of A to a single element B using a left fold with
// an initial value `init`.
func Foldl(f, init, xs interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A, ty.B) ty.B, ty.B, []ty.A) ty.B),
		f, init, xs)
	vf, vinit, vxs, tb := chk.Args[0], chk.Args[1], chk.Args[2], chk.Returns[0]

	xsLen := vxs.Len()
	vb := zeroValue(tb)
	vb.Set(vinit)
	if xsLen == 0 {
		return vb.Interface()
	}

	vb.Set(call1(vf, vxs.Index(0), vb))
	for i := 1; i < xsLen; i++ {
		vb.Set(call1(vf, vxs.Index(i), vb))
	}
	return vb.Interface()
}

// Foldr has a parametric type:
//
//	func Foldr(f func(A, B) B, init B, xs []A) B
//
// Foldr reduces a list of A to a single element B using a right fold with
// an initial value `init`.
func Foldr(f, init, xs interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A, ty.B) ty.B, ty.B, []ty.A) ty.B),
		f, init, xs)
	vf, vinit, vxs, tb := chk.Args[0], chk.Args[1], chk.Args[2], chk.Returns[0]

	xsLen := vxs.Len()
	vb := zeroValue(tb)
	vb.Set(vinit)
	if xsLen == 0 {
		return vb.Interface()
	}

	vb.Set(call1(vf, vxs.Index(xsLen-1), vb))
	for i := xsLen - 2; i >= 0; i-- {
		vb.Set(call1(vf, vxs.Index(i), vb))
	}
	return vb.Interface()
}

// Concat has a parametric type:
//
//	func Concat(xs [][]A) []A
//
// Concat returns a new flattened list by appending all elements of `xs`.
func Concat(xs interface{}) interface{} {
	chk := ty.Check(
		new(func([][]ty.A) []ty.A),
		xs)
	vxs, tflat := chk.Args[0], chk.Returns[0]

	xsLen := vxs.Len()
	vflat := reflect.MakeSlice(tflat, 0, xsLen*3)
	for i := 0; i < xsLen; i++ {
		vflat = reflect.AppendSlice(vflat, vxs.Index(i))
	}
	return vflat.Interface()
}

// Reverse has a parametric type:
//
//	func Reverse(xs []A) []A
//
// Reverse returns a new slice that is the reverse of `xs`.
func Reverse(xs interface{}) interface{} {
	chk := ty.Check(
		new(func([]ty.A) []ty.A),
		xs)
	vxs, tys := chk.Args[0], chk.Returns[0]

	xsLen := vxs.Len()
	vys := reflect.MakeSlice(tys, xsLen, xsLen)
	for i := 0; i < xsLen; i++ {
		vys.Index(i).Set(vxs.Index(xsLen - 1 - i))
	}
	return vys.Interface()
}

// Copy has a parametric type:
//
//	func Copy(xs []A) []A
//
// Copy returns a copy of `xs` using Go's `copy` operation.
func Copy(xs interface{}) interface{} {
	chk := ty.Check(
		new(func([]ty.A) []ty.A),
		xs)
	vxs, tys := chk.Args[0], chk.Returns[0]

	xsLen := vxs.Len()
	vys := reflect.MakeSlice(tys, xsLen, xsLen)
	reflect.Copy(vys, vxs)
	return vys.Interface()
}

// ParMap has a parametric type:
//
//	func ParMap(f func(A) B, xs []A) []B
//
// ParMap is just like Map, except it applies `f` to each element in `xs`
// concurrently using N worker goroutines (where N is the number of CPUs
// available reported by the Go runtime). If you want to control the number
// of goroutines spawned, use `ParMapN`.
//
// It is important that `f` not be a trivial operation, otherwise the overhead
// of executing it concurrently will result in worse performance than using
// a `Map`.
func ParMap(f, xs interface{}) interface{} {
	n := runtime.NumCPU()
	if n < 1 {
		n = 1
	}
	return ParMapN(f, xs, n)
}

// ParMapN has a parametric type:
//
//	func ParMapN(f func(A) B, xs []A, n int) []B
//
// ParMapN is just like Map, except it applies `f` to each element in `xs`
// concurrently using `n` worker goroutines.
//
// It is important that `f` not be a trivial operation, otherwise the overhead
// of executing it concurrently will result in worse performance than using
// a `Map`.
func ParMapN(f, xs interface{}, n int) interface{} {
	chk := ty.Check(
		new(func(func(ty.A) ty.B, []ty.A) []ty.B),
		f, xs)
	vf, vxs, tys := chk.Args[0], chk.Args[1], chk.Returns[0]

	xsLen := vxs.Len()
	ys := reflect.MakeSlice(tys, xsLen, xsLen)

	if n < 1 {
		n = 1
	}
	work := make(chan int, n)
	wg := new(sync.WaitGroup)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			for j := range work {
				// Good golly miss molly. Is `reflect.Value.Index`
				// safe to access/set from multiple goroutines?
				// XXX: If not, we'll need an extra wave of allocation to
				// use real slices of `reflect.Value`.
				ys.Index(j).Set(call1(vf, vxs.Index(j)))
			}
			wg.Done()
		}()
	}
	for i := 0; i < xsLen; i++ {
		work <- i
	}
	close(work)
	wg.Wait()
	return ys.Interface()
}

// Range generates a list of integers corresponding to every integer in
// the half-open interval [x, y).
//
// Range will panic if `end < start`.
func Range(start, end int) []int {
	if end < start {
		panic("range must have end greater than or equal to start")
	}
	r := make([]int, end-start)
	for i := start; i < end; i++ {
		r[i-start] = i
	}
	return r
}
