package loadbalancer

import (
	crand "crypto/rand"
	"math/rand/v2"
)

type rnd interface {
	IntN(int) int
}

// strategyPowerOfTwoChoices implements "the power-of-two-random-choices" algorithm for load balancing.
// The idea of this is two take two of the backends at random from the available backends, and select
// the backend that has the fewest in-flight requests. This algorithm more effectively balances the
// load than a round-robin approach, while also being constant time when picking: The strategy also
// has more beneficial "herd" behaviour than the "fewest connections" algorithm, especially when the
// load balancer doesn't have perfect knowledge about the global number of connections to the backend,
// for example, when running in a distributed fashion.
type strategyPowerOfTwoChoices struct {
	healthy   []*namedHandler
	unhealthy []*namedHandler
	rand      rnd
}

func newStrategyP2C() strategy {
	return &strategyPowerOfTwoChoices{
		rand: newRand(),
	}
}

func (s *strategyPowerOfTwoChoices) nextServer(map[string]struct{}) *namedHandler {
	if len(s.healthy) == 1 {
		return s.healthy[0]
	}
	// in order to not get the same backend twice, we make the second call to s.rand.IntN one fewer
	// than the length of the slice. We then have to shift over the second index if it is equal or
	// greater than the first index, wrapping round if needed.
	n1, n2 := s.rand.IntN(len(s.healthy)), s.rand.IntN(len(s.healthy)-1)
	if n2 >= n1 {
		n2 = (n2 + 1) % len(s.healthy)
	}

	h1, h2 := s.healthy[n1], s.healthy[n2]
	// ensure h1 has fewer inflight requests than h2
	if h2.inflight.Load() < h1.inflight.Load() {
		h1, h2 = h2, h1
	}

	return h1
}

func (s *strategyPowerOfTwoChoices) add(h *namedHandler) {
	s.healthy = append(s.healthy, h)
}

func (s *strategyPowerOfTwoChoices) setUp(name string, up bool) {
	if up {
		var healthy *namedHandler
		healthy, s.unhealthy = deleteAndPop(s.unhealthy, name)
		s.healthy = append(s.healthy, healthy)
		return
	}

	var unhealthy *namedHandler
	unhealthy, s.healthy = deleteAndPop(s.healthy, name)
	s.unhealthy = append(s.unhealthy, unhealthy)
}

func (s *strategyPowerOfTwoChoices) name() string {
	return "p2c"
}

func (s *strategyPowerOfTwoChoices) len() int {
	return len(s.healthy) + len(s.unhealthy)
}

func newRand() *rand.Rand {
	var seed [16]byte
	_, err := crand.Read(seed[:])
	if err != nil {
		panic(err)
	}
	var seed1, seed2 uint64
	for i := 0; i < 16; i += 8 {
		seed1 = seed1<<8 + uint64(seed[i])
		seed2 = seed2<<8 + uint64(seed[i+1])
	}
	return rand.New(rand.NewPCG(seed1, seed2))
}

// we always overwrite slice that is passed in, so it doesn't matter if we mutate the parameter
func deleteAndPop(handlers []*namedHandler, name string) (deleted *namedHandler, remaining []*namedHandler) {
	for i, h := range handlers {
		if h.name == name {
			// swap positions
			handlers[i], handlers[len(handlers)-1] = handlers[len(handlers)-1], handlers[i]
			// pop
			deleted = handlers[len(handlers)-1]
			remaining = handlers[:len(handlers)-1]
			return
		}
	}
	// this should never happen
	panic("unreachable")
}
