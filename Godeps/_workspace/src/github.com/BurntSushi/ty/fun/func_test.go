package fun

import (
	"fmt"
	"testing"
)

// The benchmarks are the test. The memo version (w/ reflection or not) should
// always be faster than the non-memo version given big enough N.

func BenchmarkFibonacciMemo(b *testing.B) {
	if flagBuiltin {
		benchmarkFibonacciMemoBuiltin(b)
	} else {
		benchmarkFibonacciMemoReflect(b)
	}
}

func benchmarkFibonacciMemoBuiltin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibonacciMemoBuiltin(30)
	}
}

func benchmarkFibonacciMemoReflect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibonacciMemoReflect(30)
	}
}

func BenchmarkFibonacciNoMemo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fibonacci(30)
	}
}

func fibonacciMemoBuiltin(n int) int {
	memo := func(f func(int) int) func(int) int {
		saved := make(map[int]int)
		return func(n int) int {
			ret, ok := saved[n]
			if ok {
				return ret
			}

			ret = f(n)
			saved[n] = ret
			return ret
		}
	}

	var fib func(n int) int
	fib = memo(func(n int) int {
		switch n {
		case 0:
			return 0
		case 1:
			return 1
		}
		return fib(n-1) + fib(n-2)
	})
	return fib(n)
}

func fibonacciMemoReflect(n int) int {
	var fib func(n int) int
	fib = Memo(func(n int) int {
		switch n {
		case 0:
			return 0
		case 1:
			return 1
		}
		return fib(n-1) + fib(n-2)
	}).(func(int) int)
	return fib(n)
}

func fibonacci(n int) int {
	switch n {
	case 0:
		return 0
	case 1:
		return 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func ExampleMemo() {
	// Memoizing a recursive function like `fibonacci`.
	// Write it like normal:
	var fib func(n int64) int64
	fib = func(n int64) int64 {
		switch n {
		case 0:
			return 0
		case 1:
			return 1
		}
		return fib(n-1) + fib(n-2)
	}

	// And wrap it with `Memo`.
	fib = Memo(fib).(func(int64) int64)

	// Will keep your CPU busy for a long time
	// without memoization.
	fmt.Println(fib(80))
	// Output:
	// 23416728348467685
}
