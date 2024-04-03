package loadbalancer

import (
	"math/rand/v2"
	"strconv"
	"testing"
)

// Sink is an exported global variable to prevent the compiler from optimizing
// out benchmarks. It isn't foolproof, but at the time of writing, it works.
var Sink *namedHandler

func benchmarkStrategy(b *testing.B, s strategy, total, down int) {
	if down >= total {
		b.Fatalf("down >= total")
	}

	var handlers []string

	healthy := map[string]struct{}{}
	for i := 0; i < total; i++ {
		name := "handler" + strconv.Itoa(i)
		s.add(&namedHandler{name: name, weight: 1})
		healthy[name] = struct{}{}
		handlers = append(handlers, name)
	}

	rand.Shuffle(total, func(i, j int) {
		handlers[i], handlers[j] = handlers[j], handlers[i]
	})

	for i := 0; i < down; i++ {
		name := handlers[i]
		s.setUp(name, false)
		delete(healthy, name)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sink = s.nextServer(healthy)
	}
	_ = Sink
}

var benches = []struct {
	name  string
	total int
	down  int
}{
	{"5_0Down", 5, 0},
	{"5_1Down", 5, 1},
	{"5_2Down", 5, 2},
	{"5_3Down", 5, 3},
	{"10_0Down", 10, 0},
	{"10_1Down", 10, 1},
	{"10_5Down", 10, 5},
	{"10_9Down", 10, 9},
	{"100_0Down", 100, 0},
	{"100_1Down", 100, 10},
	{"100_50Down", 100, 50},
	{"100_90Down", 100, 90},
	{"100_99Down", 100, 99},
}

func BenchmarkStrategyP2C(b *testing.B) {
	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			benchmarkStrategy(b, newStrategyP2C(), bench.total, bench.down)
		})
	}
}

func BenchmarkStrategyWRR(b *testing.B) {
	for _, bench := range benches {
		b.Run(bench.name, func(b *testing.B) {
			benchmarkStrategy(b, newStrategyWRR(), bench.total, bench.down)
		})
	}
}
