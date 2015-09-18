package fun

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var (
	pf          = fmt.Printf
	rng         *rand.Rand
	flagBuiltin = false
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	flag.BoolVar(&flagBuiltin, "builtin", flagBuiltin,
		"When set, benchmarks for non-type parametric functions are run.")
}

func assertDeep(t *testing.T, v1, v2 interface{}) {
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("%v != %v", v1, v2)
	}
}

func randIntSlice(size, max int) []int {
	if max == 0 {
		max = 1000000
	}
	slice := make([]int, size)
	for i := 0; i < size; i++ {
		slice[i] = rng.Intn(max)
	}
	return slice
}

func randInt64Slice(size, max int64) []int64 {
	if max == 0 {
		max = 1000000
	}
	slice := make([]int64, size)
	for i := int64(0); i < size; i++ {
		slice[i] = rng.Int63n(max)
	}
	return slice
}
