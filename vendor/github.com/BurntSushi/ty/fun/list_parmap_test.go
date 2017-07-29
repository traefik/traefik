package fun

import (
	"math"
	"runtime"
	"sync"
	"testing"
)

func TestParMap(t *testing.T) {
	square := func(x int) int { return x * x }
	squares := ParMap(square, []int{1, 2, 3, 4, 5}).([]int)

	assertDeep(t, squares, []int{1, 4, 9, 16, 25})
	assertDeep(t, []int{}, ParMap(square, []int{}).([]int))
}

func BenchmarkParMapSquare(b *testing.B) {
	if flagBuiltin {
		benchmarkParMapSquareBuiltin(b)
	} else {
		benchmarkParMapSquareReflect(b)
	}
}

func benchmarkParMapSquareReflect(b *testing.B) {
	b.StopTimer()
	square := func(a int64) int64 { return a * a }
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = ParMap(square, list).([]int64)
	}
}

func benchmarkParMapSquareBuiltin(b *testing.B) {
	b.StopTimer()
	square := func(a int64) int64 { return a * a }
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = parMapInt64(square, list)
	}
}

func BenchmarkParMapPrime(b *testing.B) {
	if flagBuiltin {
		benchmarkParMapPrimeBuiltin(b)
	} else {
		benchmarkParMapPrimeReflect(b)
	}
}

func benchmarkParMapPrimeReflect(b *testing.B) {
	b.StopTimer()
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = ParMap(primeFactors, list).([][]int64)
	}
}

func benchmarkParMapPrimeBuiltin(b *testing.B) {
	b.StopTimer()
	list := randInt64Slice(1000, 1<<30)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = parMapSliceInt64(primeFactors, list)
	}
}

func primeFactors(n int64) []int64 {
	if isPrime(n) {
		return []int64{n}
	}

	bound := int64(math.Floor(math.Sqrt(float64(n))))
	for i := int64(2); i <= bound; i++ {
		if n%i == 0 {
			return append(primeFactors(i), primeFactors(n/i)...)
		}
	}
	panic("unreachable")
}

func isPrime(n int64) bool {
	bound := int64(math.Floor(math.Sqrt(float64(n))))
	for i := int64(2); i <= bound; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func parMapInt64(f func(n int64) int64, xs []int64) []int64 {
	ys := make([]int64, len(xs), len(xs))
	N := runtime.NumCPU()
	if N < 1 {
		N = 1
	}
	work := make(chan int, N)
	wg := new(sync.WaitGroup)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			for j := range work {
				ys[j] = f(xs[j])
			}
			wg.Done()
		}()
	}
	for i := 0; i < len(xs); i++ {
		work <- i
	}
	close(work)
	wg.Wait()
	return ys
}

func parMapSliceInt64(f func(n int64) []int64, xs []int64) [][]int64 {
	ys := make([][]int64, len(xs), len(xs))
	N := runtime.NumCPU()
	if N < 1 {
		N = 1
	}
	work := make(chan int, N)
	wg := new(sync.WaitGroup)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			for j := range work {
				ys[j] = f(xs[j])
			}
			wg.Done()
		}()
	}
	for i := 0; i < len(xs); i++ {
		work <- i
	}
	close(work)
	wg.Wait()
	return ys
}
