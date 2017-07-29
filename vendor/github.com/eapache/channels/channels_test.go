package channels

import (
	"math/rand"
	"testing"
	"time"
)

func testChannel(t *testing.T, name string, ch Channel) {
	go func() {
		for i := 0; i < 1000; i++ {
			ch.In() <- i
		}
		ch.Close()
	}()
	for i := 0; i < 1000; i++ {
		val := <-ch.Out()
		if i != val.(int) {
			t.Fatal(name, "expected", i, "but got", val.(int))
		}
	}
}

func testChannelPair(t *testing.T, name string, in InChannel, out OutChannel) {
	go func() {
		for i := 0; i < 1000; i++ {
			in.In() <- i
		}
		in.Close()
	}()
	for i := 0; i < 1000; i++ {
		val := <-out.Out()
		if i != val.(int) {
			t.Fatal("pair", name, "expected", i, "but got", val.(int))
		}
	}
}

func testChannelConcurrentAccessors(t *testing.T, name string, ch Channel) {
	// no asserts here, this is just for the race detector's benefit
	go ch.Len()
	go ch.Cap()

	go func() {
		ch.In() <- nil
	}()

	go func() {
		<-ch.Out()
	}()
}

func TestPipe(t *testing.T) {
	a := NewNativeChannel(None)
	b := NewNativeChannel(None)

	Pipe(a, b)

	testChannelPair(t, "pipe", a, b)
}

func TestWeakPipe(t *testing.T) {
	a := NewNativeChannel(None)
	b := NewNativeChannel(None)

	WeakPipe(a, b)

	testChannelPair(t, "pipe", a, b)
}

func testMultiplex(t *testing.T, multi func(output SimpleInChannel, inputs ...SimpleOutChannel)) {
	a := NewNativeChannel(None)
	b := NewNativeChannel(None)

	multi(b, a)

	testChannelPair(t, "simple multiplex", a, b)

	a = NewNativeChannel(None)
	inputs := []Channel{
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
	}

	multi(a, inputs[0], inputs[1], inputs[2], inputs[3])

	go func() {
		rand.Seed(time.Now().Unix())
		for i := 0; i < 1000; i++ {
			inputs[rand.Intn(len(inputs))].In() <- i
		}
		for i := range inputs {
			inputs[i].Close()
		}
	}()
	for i := 0; i < 1000; i++ {
		val := <-a.Out()
		if i != val.(int) {
			t.Fatal("multiplexing expected", i, "but got", val.(int))
		}
	}
}

func TestMultiplex(t *testing.T) {
	testMultiplex(t, Multiplex)
}

func TestWeakMultiplex(t *testing.T) {
	testMultiplex(t, WeakMultiplex)
}

func testTee(t *testing.T, tee func(input SimpleOutChannel, outputs ...SimpleInChannel)) {
	a := NewNativeChannel(None)
	b := NewNativeChannel(None)

	tee(a, b)

	testChannelPair(t, "simple tee", a, b)

	a = NewNativeChannel(None)
	outputs := []Channel{
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
	}

	tee(a, outputs[0], outputs[1], outputs[2], outputs[3])

	go func() {
		for i := 0; i < 1000; i++ {
			a.In() <- i
		}
		a.Close()
	}()
	for i := 0; i < 1000; i++ {
		for _, output := range outputs {
			val := <-output.Out()
			if i != val.(int) {
				t.Fatal("teeing expected", i, "but got", val.(int))
			}
		}
	}
}

func TestTee(t *testing.T) {
	testTee(t, Tee)
}

func TestWeakTee(t *testing.T) {
	testTee(t, WeakTee)
}

func testDistribute(t *testing.T, dist func(input SimpleOutChannel, outputs ...SimpleInChannel)) {
	a := NewNativeChannel(None)
	b := NewNativeChannel(None)

	dist(a, b)

	testChannelPair(t, "simple distribute", a, b)

	a = NewNativeChannel(None)
	outputs := []Channel{
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
		NewNativeChannel(None),
	}

	dist(a, outputs[0], outputs[1], outputs[2], outputs[3])

	go func() {
		for i := 0; i < 1000; i++ {
			a.In() <- i
		}
		a.Close()
	}()

	received := make([]bool, 1000)
	for _ = range received {
		var val interface{}
		select {
		case val = <-outputs[0].Out():
		case val = <-outputs[1].Out():
		case val = <-outputs[2].Out():
		case val = <-outputs[3].Out():
		}
		if received[val.(int)] {
			t.Fatal("distribute got value twice", val.(int))
		}
		received[val.(int)] = true
	}
	for i := range received {
		if !received[i] {
			t.Fatal("distribute missed", i)
		}
	}
}

func TestDistribute(t *testing.T) {
	testDistribute(t, Distribute)
}

func TestWeakDistribute(t *testing.T) {
	testDistribute(t, WeakDistribute)
}

func TestWrap(t *testing.T) {
	rawChan := make(chan int, 5)
	ch := Wrap(rawChan)

	for i := 0; i < 5; i++ {
		rawChan <- i
	}
	close(rawChan)

	for i := 0; i < 5; i++ {
		x := (<-ch.Out()).(int)
		if x != i {
			t.Error("Wrapped value", x, "was expecting", i)
		}
	}
	_, ok := <-ch.Out()
	if ok {
		t.Error("Wrapped channel didn't close")
	}
}

func TestUnwrap(t *testing.T) {
	rawChan := make(chan int)
	ch := NewNativeChannel(5)
	Unwrap(ch, rawChan)

	for i := 0; i < 5; i++ {
		ch.In() <- i
	}
	ch.Close()

	for i := 0; i < 5; i++ {
		x := <-rawChan
		if x != i {
			t.Error("Unwrapped value", x, "was expecting", i)
		}
	}
	_, ok := <-rawChan
	if ok {
		t.Error("Unwrapped channel didn't close")
	}
}

func ExampleChannel() {
	var ch Channel

	ch = NewInfiniteChannel()

	for i := 0; i < 10; i++ {
		ch.In() <- nil
	}

	for i := 0; i < 10; i++ {
		<-ch.Out()
	}
}
