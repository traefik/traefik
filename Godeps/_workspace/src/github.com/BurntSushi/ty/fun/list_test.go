package fun

import (
	"testing"
)

func TestMap(t *testing.T) {
	square := func(x int) int { return x * x }
	squares := Map(square, []int{1, 2, 3, 4, 5}).([]int)

	assertDeep(t, squares, []int{1, 4, 9, 16, 25})
	assertDeep(t, []int{}, Map(square, []int{}).([]int))

	strlen := func(s string) int { return len(s) }
	lens := Map(strlen, []string{"abc", "ab", "a"}).([]int)
	assertDeep(t, lens, []int{3, 2, 1})
}

func TestFilter(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	evens := Filter(even, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).([]int)

	assertDeep(t, evens, []int{2, 4, 6, 8, 10})
	assertDeep(t, []int{}, Filter(even, []int{}).([]int))
}

func TestFoldl(t *testing.T) {
	// Use an operation that isn't associative so that we know we've got
	// the left/right folds done correctly.
	reducer := func(a, b int) int { return b % a }
	v := Foldl(reducer, 7, []int{4, 5, 6}).(int)

	assertDeep(t, v, 3)
	assertDeep(t, 0, Foldl(reducer, 0, []int{}).(int))
}

func TestFoldr(t *testing.T) {
	// Use an operation that isn't associative so that we know we've got
	// the left/right folds done correctly.
	reducer := func(a, b int) int { return b % a }
	v := Foldr(reducer, 7, []int{4, 5, 6}).(int)

	assertDeep(t, v, 1)
	assertDeep(t, 0, Foldr(reducer, 0, []int{}).(int))
}

func TestConcat(t *testing.T) {
	toflat := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	flat := Concat(toflat).([]int)

	assertDeep(t, flat, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
}

func TestReverse(t *testing.T) {
	reversed := Reverse([]int{1, 2, 3, 4, 5}).([]int)

	assertDeep(t, reversed, []int{5, 4, 3, 2, 1})
}

func TestCopy(t *testing.T) {
	orig := []int{1, 2, 3, 4, 5}
	copied := Copy(orig).([]int)

	orig[1] = 999

	assertDeep(t, copied, []int{1, 2, 3, 4, 5})
}

func TestPointers(t *testing.T) {
	type temp struct {
		val int
	}
	square := func(t *temp) *temp { return &temp{t.val * t.val} }
	squares := Map(square, []*temp{
		{1}, {2}, {3}, {4}, {5},
	})

	assertDeep(t, squares, []*temp{
		{1}, {4}, {9}, {16}, {25},
	})
}

func BenchmarkMapSquare(b *testing.B) {
	if flagBuiltin {
		benchmarkMapSquareBuiltin(b)
	} else {
		benchmarkMapSquareReflect(b)
	}
}

func benchmarkMapSquareReflect(b *testing.B) {
	b.StopTimer()
	square := func(a int64) int64 { return a * a }
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = Map(square, list).([]int64)
	}
}

func benchmarkMapSquareBuiltin(b *testing.B) {
	b.StopTimer()
	square := func(a int64) int64 { return a * a }
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ret := make([]int64, len(list))
		for i := 0; i < len(list); i++ {
			ret[i] = square(list[i])
		}
	}
}

func BenchmarkMapPrime(b *testing.B) {
	if flagBuiltin {
		benchmarkMapPrimeBuiltin(b)
	} else {
		benchmarkMapPrimeReflect(b)
	}
}

func benchmarkMapPrimeReflect(b *testing.B) {
	b.StopTimer()
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = Map(primeFactors, list).([][]int64)
	}
}

func benchmarkMapPrimeBuiltin(b *testing.B) {
	b.StopTimer()
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ret := make([][]int64, len(list))
		for i := 0; i < len(list); i++ {
			ret[i] = primeFactors(list[i])
		}
	}
}
