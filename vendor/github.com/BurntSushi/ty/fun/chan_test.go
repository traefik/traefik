package fun

import (
	"testing"
)

func TestAsyncChan(t *testing.T) {
	s, r := AsyncChan(new(chan int))
	send, recv := s.(chan<- int), r.(<-chan int)

	sending := randIntSlice(1000, 0)
	for _, v := range sending {
		send <- v
	}
	close(send)

	received := make([]int, 0)
	for v := range recv {
		received = append(received, v)
	}

	assertDeep(t, sending, received)
}
