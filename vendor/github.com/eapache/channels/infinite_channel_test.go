package channels

import "testing"

func TestInfiniteChannel(t *testing.T) {
	var ch Channel

	ch = NewInfiniteChannel()
	testChannel(t, "infinite channel", ch)

	ch = NewInfiniteChannel()
	testChannelPair(t, "infinite channel", ch, ch)

	ch = NewInfiniteChannel()
	testChannelConcurrentAccessors(t, "infinite channel", ch)
}

func BenchmarkInfiniteChannelSerial(b *testing.B) {
	ch := NewInfiniteChannel()
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	for i := 0; i < b.N; i++ {
		<-ch.Out()
	}
}

func BenchmarkInfiniteChannelParallel(b *testing.B) {
	ch := NewInfiniteChannel()
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch.Out()
		}
		ch.Close()
	}()
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
	}
	<-ch.Out()
}

func BenchmarkInfiniteChannelTickTock(b *testing.B) {
	ch := NewInfiniteChannel()
	for i := 0; i < b.N; i++ {
		ch.In() <- nil
		<-ch.Out()
	}
}
