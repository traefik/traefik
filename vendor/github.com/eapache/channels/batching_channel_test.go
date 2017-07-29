package channels

import "testing"

func testBatches(t *testing.T, ch Channel) {
	go func() {
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()

	i := 0
	for val := range ch.Out() {
		for _, elem := range val.([]interface{}) {
			if i != elem.(int) {
				t.Fatal("batching channel expected", i, "but got", elem.(int))
			}
			i++
		}
	}
}

func TestBatchingChannel(t *testing.T) {
	ch := NewBatchingChannel(Infinity)
	testBatches(t, ch)

	ch = NewBatchingChannel(2)
	testBatches(t, ch)

	ch = NewBatchingChannel(1)
	testChannelConcurrentAccessors(t, "batching channel", ch)
}

func TestBatchingChannelCap(t *testing.T) {
	ch := NewBatchingChannel(Infinity)
	if ch.Cap() != Infinity {
		t.Error("incorrect capacity on infinite channel")
	}

	ch = NewBatchingChannel(5)
	if ch.Cap() != 5 {
		t.Error("incorrect capacity on infinite channel")
	}
}
