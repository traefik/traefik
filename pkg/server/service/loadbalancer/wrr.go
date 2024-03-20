package loadbalancer

import (
	"container/heap"
)

// strategyWRR is a WeightedRoundRobin load balancer strategy based on Earliest Deadline First (EDF).
// (https://en.wikipedia.org/wiki/Earliest_deadline_first_scheduling)
// Each pick from the schedule has the earliest deadline entry selected.
// Entries have deadlines set at currentDeadline + 1 / weight,
// providing weighted round-robin behavior with floating point weights and an O(log n) pick time.
type strategyWRR struct {
	handlers    []*namedHandler
	curDeadline float64
	deadline    float64
}

func newStrategyWRR() strategy {
	return &strategyWRR{}
}

func (s *strategyWRR) nextServer(status map[string]struct{}) *namedHandler {

	var handler *namedHandler
	for {
		// Pick handler with closest deadline.
		handler = heap.Pop(s).(*namedHandler)

		// curDeadline should be handler's deadline so that new added entry would have a fair competition environment with the old ones.
		s.curDeadline = handler.deadline
		handler.deadline += 1 / handler.weight

		heap.Push(s, handler)
		if _, ok := status[handler.name]; ok {
			break
		}
	}
	return handler
}

func (s *strategyWRR) add(h *namedHandler) {
	h.deadline = s.curDeadline + 1/h.weight
	heap.Push(s, h)
}

func (s *strategyWRR) setUp(string, bool) {}

func (s *strategyWRR) name() string {
	return "wrr"
}

func (s *strategyWRR) len() int {
	return len(s.handlers)
}

// Len implements heap.Interface/sort.Interface.
func (s *strategyWRR) Len() int { return s.len() }

// Less implements heap.Interface/sort.Interface.
func (s *strategyWRR) Less(i, j int) bool {
	return s.handlers[i].deadline < s.handlers[j].deadline
}

// Swap implements heap.Interface/sort.Interface.
func (s *strategyWRR) Swap(i, j int) {
	s.handlers[i], s.handlers[j] = s.handlers[j], s.handlers[i]
}

// Push implements heap.Interface for pushing an item into the heap.
func (s *strategyWRR) Push(x interface{}) {
	h, ok := x.(*namedHandler)
	if !ok {
		return
	}

	s.handlers = append(s.handlers, h)
}

// Pop implements heap.Interface for popping an item from the heap.
// It panics if b.Len() < 1.
func (s *strategyWRR) Pop() interface{} {
	h := s.handlers[len(s.handlers)-1]
	s.handlers = s.handlers[0 : len(s.handlers)-1]
	return h
}
