package fun

import (
	"math/rand"
	"testing"
)

func TestShuffle(t *testing.T) {
	nums := Range(0, 100)
	Shuffle(nums)

	assertDeep(t, Set(nums), Set(Range(0, 100)))
}

func TestSample(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	nums := Range(0, 100)
	sample := SampleGen(nums, 3, rng).([]int)

	assertDeep(t, Set([]int{35, 40, 50}), Set(sample))
}

func BenchmarkShuffle(b *testing.B) {
	if flagBuiltin {
		benchmarkShuffleBuiltin(b)
	} else {
		benchmarkShuffleReflect(b)
	}
}

func benchmarkShuffleBuiltin(b *testing.B) {
	b.StopTimer()
	list := randIntSlice(10000, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		shuffle(list)
	}
}

func benchmarkShuffleReflect(b *testing.B) {
	b.StopTimer()
	list := randIntSlice(10000, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Shuffle(list)
	}
}

func shuffle(xs []int) {
	for i := len(xs) - 1; i >= 1; i-- {
		j := randNumGen.Intn(i + 1)
		xs[i], xs[j] = xs[j], xs[i]
	}
}

func BenchmarkSample(b *testing.B) {
	if flagBuiltin {
		benchmarkSampleBuiltin(b)
	} else {
		benchmarkSampleReflect(b)
	}
}

func benchmarkSampleBuiltin(b *testing.B) {
	b.StopTimer()
	list := randIntSlice(10000, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		sample(list, 100)
	}
}

func benchmarkSampleReflect(b *testing.B) {
	b.StopTimer()
	list := randIntSlice(10000, 0)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Sample(list, 100)
	}
}

func sample(pop []int, n int) []int {
	if n == 0 {
		return []int{}
	}
	if n > len(pop) {
		n = len(pop)
	}

	samp := make([]int, n, n)
	choices := randNumGen.Perm(len(pop))
	for i := 0; i < n; i++ {
		samp[i] = pop[choices[i]]
	}
	return samp
}
