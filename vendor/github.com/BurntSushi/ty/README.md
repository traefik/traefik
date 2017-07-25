Package `ty` provides utilities for writing type parametric functions with run
time type safety.

This package contains two sub-packages `fun` and `data` which define some
potentially useful functions and abstractions using the type checker in
this package.

## Requirements

Go tip (or 1.1 when it's released) is required. This package will not work
with Go 1.0.x or earlier.

The very foundation of this package only recently became possible with the
addition of 3 new functions in the standard library `reflect` package:
SliceOf, MapOf and ChanOf. In particular, it provides the ability to
dynamically construct types at run time from component types.

Further extensions to this package can be made if similar functions are added
for structs and functions(?).

## Installation

```bash
go get github.com/BurntSushi/ty
go get github.com/BurntSushi/ty/fun
```

## Examples

Squaring each integer in a slice:

```go
square := func(x int) int { return x * x }
nums := []int{1, 2, 3, 4, 5}
squares := Map(square, nums).([]int)
```

Reversing any slice:

```go
slice := []string{"a", "b", "c"}
reversed := Reverse(slice).([]string)
```

Sorting any slice:

```go
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
```

Parallel map:

```go
// Compute the prime factorization concurrently
// for every integer in [1000, 10000].
primeFactors := func(n int) []int { // compute prime factors }
factors := ParMap(primeFactors, Range(1000, 10001)).([]int)
```

Asynchronous channel without a fixed size buffer:

```go
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
```

Shuffle any slice in place:

```go
jumbleMe := []string{"The", "quick", "brown", "fox"}
Shuffle(jumbleMe)
```

Function memoization:

```go
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
```

