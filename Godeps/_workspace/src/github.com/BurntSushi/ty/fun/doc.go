/*
Package fun provides type parametric utility functions for lists, sets,
channels and maps.

The central contribution of this package is a set of functions that operate
on values without depending on their types while maintaining type safety at
run time using the `reflect` package.

There are two primary concerns when deciding whether to use this package
or not: the loss of compile time type safety and performance. In particular,
with regard to performance, most functions here are much slower than their
built-in counter parts. However, there are a couple where the overhead of
reflection is relatively insignificant: AsyncChan and ParMap.

In terms of code structure and organization, the price is mostly paid inside
of the package due to the annoyances of operating with `reflect`. The caller
usually only has one obligation other than to provide values consistent with
the type of the function: type assert the result to the desired type.

When the caller provides values that are inconsistent with the parametric type
of the function, the function will panic with a `TypeError`. (Either because
the types cannot be unified or because they cannot be constructed due to
limitations of the `reflect` package. See the `github.com/BurntSushi/ty`
package for more details.)

Requirements

Go tip (or 1.1 when it's released) is required. This package will not work
with Go 1.0.x or earlier.

The very foundation of this package only recently became possible with the
addition of 3 new functions in the standard library `reflect` package:
SliceOf, MapOf and ChanOf. In particular, it provides the ability to
dynamically construct types at run time from component types.

Further extensions to this package can be made if similar functions are added
for structs and functions(?).

Examples

Squaring each integer in a slice:

	square := func(x int) int { return x * x }
	nums := []int{1, 2, 3, 4, 5}
	squares := Map(square, nums).([]int)

Reversing any slice:

	slice := []string{"a", "b", "c"}
	reversed := Reverse(slice).([]string)

Sorting any slice:

	// Sort a slice of structs with first class functions.
	type Album struct {
		Title string
		Year int
	}
	albums := []Album{
		{"Born to Run", 1975},
		{"WIESS",       1973},
		{"Darkness",    1978},
		{"Greetings",   1973},
	}

	less := func(a, b Album) bool { return a.Year < b.Year },
	sorted := QuickSort(less, albums).([]Album)

Parallel map:

	// Compute the prime factorization concurrently
	// for every integer in [1000, 10000].
	primeFactors := func(n int) []int { // compute prime factors }
	factors := ParMap(primeFactors, Range(1000, 10001)).([]int)

Asynchronous channel without a fixed size buffer:

	s, r := AsyncChan(new(chan int))
	send, recv := s.(chan<- int), r.(<-chan int)

	// Send as much as you want.
	for i := 0; i < 100; i++ {
		s <- i
	}
	close(s)
	for i := range recv {
		// do something with `i`
	}

Shuffle any slice in place:

	jumbleMe := []string{"The", "quick", "brown", "fox"}
	Shuffle(jumbleMe)

Function memoization:

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
		return fib(n - 1) + fib(n - 2)
	}

	// And wrap it with `Memo`.
	fib = Memo(fib).(func(int64) int64)

	// Will keep your CPU busy for a long time
	// without memoization.
	fmt.Println(fib(80))

*/
package fun
