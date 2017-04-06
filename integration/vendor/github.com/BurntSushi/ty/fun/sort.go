package fun

import (
	"reflect"
	"sort"

	"github.com/BurntSushi/ty"
)

// QuickSort has a parametric type:
//
//	func QuickSort(less func(x1 A, x2 A) bool, []A) []A
//
// QuickSort applies the "quicksort" algorithm to return a new sorted list
// of `xs`, where `xs` is not modified.
//
// `less` should be a function that returns true if and only if `x1` is less
// than `x2`.
func QuickSort(less, xs interface{}) interface{} {
	chk := ty.Check(
		new(func(func(ty.A, ty.A) bool, []ty.A) []ty.A),
		less, xs)
	vless, vxs, tys := chk.Args[0], chk.Args[1], chk.Returns[0]

	var qsort func(left, right int)
	var partition func(left, right, pivot int) int
	xsind := Range(0, vxs.Len())

	qsort = func(left, right int) {
		if left >= right {
			return
		}
		pivot := (left + right) / 2
		pivot = partition(left, right, pivot)

		qsort(left, pivot-1)
		qsort(pivot+1, right)
	}
	partition = func(left, right, pivot int) int {
		vpivot := xsind[pivot]
		xsind[pivot], xsind[right] = xsind[right], xsind[pivot]

		ind := left
		for i := left; i < right; i++ {
			if call1(vless, vxs.Index(xsind[i]), vxs.Index(vpivot)).Bool() {
				xsind[i], xsind[ind] = xsind[ind], xsind[i]
				ind++
			}
		}
		xsind[ind], xsind[right] = xsind[right], xsind[ind]
		return ind
	}

	// Sort `xsind` in place.
	qsort(0, len(xsind)-1)

	vys := reflect.MakeSlice(tys, len(xsind), len(xsind))
	for i, xsIndex := range xsind {
		vys.Index(i).Set(vxs.Index(xsIndex))
	}
	return vys.Interface()
}

// Sort has a parametric type:
//
//	func Sort(less func(x1 A, x2 A) bool, []A)
//
// Sort uses the standard library `sort` package to sort `xs` in place.
//
// `less` should be a function that returns true if and only if `x1` is less
// than `x2`.
func Sort(less, xs interface{}) {
	chk := ty.Check(
		new(func(func(ty.A, ty.A) bool, []ty.A)),
		less, xs)

	vless, vxs := chk.Args[0], chk.Args[1]
	sort.Sort(&sortable{vless, vxs, swapperOf(vxs.Type().Elem())})
}

type sortable struct {
	less    reflect.Value
	xs      reflect.Value
	swapper swapper
}

func (s *sortable) Less(i, j int) bool {
	ith, jth := s.xs.Index(i), s.xs.Index(j)
	return call1(s.less, ith, jth).Bool()
}

func (s *sortable) Swap(i, j int) {
	s.swapper.swap(s.xs.Index(i), s.xs.Index(j))
}

func (s *sortable) Len() int {
	return s.xs.Len()
}
