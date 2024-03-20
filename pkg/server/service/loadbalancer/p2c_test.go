package loadbalancer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRand struct {
	vals  []int
	calls int
}

func (m *mockRand) IntN(int) int {
	defer func() {
		m.calls++
	}()
	return m.vals[m.calls]
}

func testHandlers(inflights ...int) []*namedHandler {
	var out []*namedHandler
	for i, inflight := range inflights {
		h := &namedHandler{
			name: fmt.Sprint(i),
		}
		h.inflight.Store(int64(inflight))
		out = append(out, h)
	}
	return out
}

func TestStrategyTwoRandomChoices_AllHealthy(t *testing.T) {
	cases := []struct {
		name          string
		handlers      []*namedHandler
		rand          *mockRand
		expectHandler string
	}{
		{
			name:          "oneHealthyHandler",
			handlers:      testHandlers(0),
			rand:          nil,
			expectHandler: "0",
		},
		{
			name:          "twoHandlersZeroInflight",
			handlers:      testHandlers(0, 0),
			rand:          &mockRand{vals: []int{1, 0}},
			expectHandler: "1",
		},
		{
			name:          "choosesLowerOfTwo",
			handlers:      testHandlers(0, 1),
			rand:          &mockRand{vals: []int{1, 0}},
			expectHandler: "0",
		},
		{
			name:          "choosesLowerOfThree",
			handlers:      testHandlers(10, 90, 40),
			rand:          &mockRand{vals: []int{1, 1}},
			expectHandler: "2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := newStrategyP2C()
			strategy.(*strategyPowerOfTwoChoices).rand = tc.rand

			status := map[string]struct{}{}
			for _, h := range tc.handlers {
				strategy.add(h)
				status[h.name] = struct{}{}
			}

			got := strategy.nextServer(status)

			assert.Equal(t, tc.expectHandler, got.name, "balancer strategy gave unexpected backend handler")
		})
	}
}
