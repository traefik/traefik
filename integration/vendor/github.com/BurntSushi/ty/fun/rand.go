package fun

import (
	"math/rand"
	"reflect"
	"time"

	"github.com/BurntSushi/ty"
)

var randNumGen *rand.Rand

func init() {
	randNumGen = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// ShuffleGen has a parametric type:
//
//	func ShuffleGen(xs []A, rng *rand.Rand)
//
// ShuffleGen shuffles `xs` in place using the given random number
// generator `rng`.
func ShuffleGen(xs interface{}, rng *rand.Rand) {
	chk := ty.Check(
		new(func([]ty.A, *rand.Rand)),
		xs, rng)
	vxs := chk.Args[0]

	// Implements the Fisher-Yates shuffle: http://goo.gl/Hb9vg
	xsLen := vxs.Len()
	swapper := swapperOf(vxs.Type().Elem())
	for i := xsLen - 1; i >= 1; i-- {
		j := rng.Intn(i + 1)
		swapper.swap(vxs.Index(i), vxs.Index(j))
	}
}

// Shuffle has a parametric type:
//
//	func Shuffle(xs []A)
//
// Shuffle shuffles `xs` in place using a default random number
// generator seeded once at program initialization.
func Shuffle(xs interface{}) {
	ShuffleGen(xs, randNumGen)
}

// Sample has a parametric type:
//
//	func Sample(population []A, n int) []A
//
// Sample returns a random sample of size `n` from a list
// `population` using a default random number generator seeded once at
// program initialization.
// All elements in `population` have an equal chance of being selected.
// If `n` is greater than the size of `population`, then `n` is set to
// the size of the population.
func Sample(population interface{}, n int) interface{} {
	return SampleGen(population, n, randNumGen)
}

// SampleGen has a parametric type:
//
//	func SampleGen(population []A, n int, rng *rand.Rand) []A
//
// SampleGen returns a random sample of size `n` from a list
// `population` using a given random number generator `rng`.
// All elements in `population` have an equal chance of being selected.
// If `n` is greater than the size of `population`, then `n` is set to
// the size of the population.
func SampleGen(population interface{}, n int, rng *rand.Rand) interface{} {
	chk := ty.Check(
		new(func([]ty.A, int, *rand.Rand) []ty.A),
		population, n, rng)
	rpop, tsamp := chk.Args[0], chk.Returns[0]

	popLen := rpop.Len()
	if n == 0 {
		return reflect.MakeSlice(tsamp, 0, 0).Interface()
	}
	if n > popLen {
		n = popLen
	}

	// TODO(burntsushi): Implement an algorithm that doesn't depend on
	// the size of the population.

	rsamp := reflect.MakeSlice(tsamp, n, n)
	choices := rng.Perm(popLen)
	for i := 0; i < n; i++ {
		rsamp.Index(i).Set(rpop.Index(choices[i]))
	}
	return rsamp.Interface()
}
