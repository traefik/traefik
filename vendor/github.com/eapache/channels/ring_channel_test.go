package channels

import "testing"

func TestRingChannel(t *testing.T) {
	var ch Channel

	ch = NewRingChannel(Infinity) // yes this is rather silly, but it should work
	testChannel(t, "infinite ring-buffer channel", ch)

	ch = NewRingChannel(None)
	go func() {
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	prev := -1
	for i := range ch.Out() {
		if prev >= i.(int) {
			t.Fatal("ring channel prev", prev, "but got", i.(int))
		}
	}

	ch = NewRingChannel(10)
	for i := 0; i < 1000; i++ {
		ch.In() <- i
	}
	ch.Close()
	for i := 990; i < 1000; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal("ring channel expected", i, "but got", val.(int))
		}
	}
	if val, open := <-ch.Out(); open == true {
		t.Fatal("ring channel expected closed but got", val)
	}

	ch = NewRingChannel(None)
	ch.In() <- 0
	ch.Close()
	if val, open := <-ch.Out(); open == true {
		t.Fatal("ring channel expected closed but got", val)
	}

	ch = NewRingChannel(2)
	testChannelConcurrentAccessors(t, "ring channel", ch)
}
