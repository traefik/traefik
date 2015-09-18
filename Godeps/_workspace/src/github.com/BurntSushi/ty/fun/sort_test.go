package fun

import (
	"sort"
	"testing"
)

func TestSort(t *testing.T) {
	tosort := []int{10, 3, 5, 1, 15, 6}
	Sort(func(a, b int) bool { return b < a }, tosort)

	assertDeep(t, tosort, []int{15, 10, 6, 5, 3, 1})
}

func TestQuickSort(t *testing.T) {
	tosort := []int{10, 3, 5, 1, 15, 6}
	sorted := QuickSort(
		func(a, b int) bool {
			return b < a
		}, tosort).([]int)

	assertDeep(t, sorted, []int{15, 10, 6, 5, 3, 1})
}

func BenchmarkSort(b *testing.B) {
	if flagBuiltin {
		benchmarkSortBuiltin(b)
	} else {
		benchmarkSortReflect(b)
	}
}

func benchmarkSortReflect(b *testing.B) {
	less := func(a, b int) bool { return a < b }

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		list := randIntSlice(1000, 0)
		b.StartTimer()

		Sort(less, list)
	}
}

func benchmarkSortBuiltin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		list := randIntSlice(1000, 0)
		b.StartTimer()

		sort.Sort(sort.IntSlice(list))
	}
}

func BenchmarkQuickSort(b *testing.B) {
	if flagBuiltin {
		benchmarkQuickSortBuiltin(b)
	} else {
		benchmarkQuickSortReflect(b)
	}
}

func benchmarkQuickSortReflect(b *testing.B) {
	less := func(a, b int) bool { return a < b }

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		list := randIntSlice(1000, 0)
		b.StartTimer()

		_ = QuickSort(less, list)
	}
}

func benchmarkQuickSortBuiltin(b *testing.B) {
	less := func(a, b int) bool { return a < b }

	quicksort := func(xs []int) []int {
		ys := make([]int, len(xs))
		copy(ys, xs)

		var qsort func(left, right int)
		var partition func(left, right, pivot int) int

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
			vpivot := ys[pivot]
			ys[pivot], ys[right] = ys[right], ys[pivot]

			ind := left
			for i := left; i < right; i++ {
				if less(ys[i], vpivot) {
					ys[i], ys[ind] = ys[ind], ys[i]
					ind++
				}
			}
			ys[ind], ys[right] = ys[right], ys[ind]
			return ind
		}

		return ys
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		list := randIntSlice(1000, 0)
		b.StartTimer()

		_ = quicksort(list)
	}
}
