package channels

import "testing"

func TestOverflowingChannel(t *testing.T) {
	var ch Channel

	ch = NewOverflowingChannel(Infinity) // yes this is rather silly, but it should work
	testChannel(t, "infinite overflowing channel", ch)

	ch = NewOverflowingChannel(None)
	go func() {
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	prev := -1
	for i := range ch.Out() {
		if prev >= i.(int) {
			t.Fatal("overflowing channel prev", prev, "but got", i.(int))
		}
	}

	ch = NewOverflowingChannel(10)
	for i := 0; i < 1000; i++ {
		ch.In() <- i
	}
	ch.Close()
	for i := 0; i < 10; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("overflowing channel expected", i, "but got", val.(int))
		}
	}
	if val, open := <-ch.Out(); open == true {
		t.Fatal("overflowing channel expected closed but got", val)
	}

	ch = NewOverflowingChannel(None)
	ch.In() <- 0
	ch.Close()
	if val, open := <-ch.Out(); open == true {
		t.Fatal("overflowing channel expected closed but got", val)
	}

	ch = NewOverflowingChannel(2)
	testChannelConcurrentAccessors(t, "overflowing channel", ch)
}
