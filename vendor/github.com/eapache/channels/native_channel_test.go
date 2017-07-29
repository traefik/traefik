package channels

import "testing"

func TestNativeChannels(t *testing.T) {
	var ch Channel

	ch = NewNativeChannel(None)
	testChannel(t, "bufferless native channel", ch)

	ch = NewNativeChannel(None)
	testChannelPair(t, "bufferless native channel", ch, ch)

	ch = NewNativeChannel(5)
	testChannel(t, "5-buffer native channel", ch)

	ch = NewNativeChannel(5)
	testChannelPair(t, "5-buffer native channel", ch, ch)

	ch = NewNativeChannel(None)
	testChannelConcurrentAccessors(t, "native channel", ch)
}

func TestNativeInOutChannels(t *testing.T) {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})

	Pipe(NativeOutChannel(ch1), NativeInChannel(ch2))
	NativeInChannel(ch1).Close()
}

func TestDeadChannel(t *testing.T) {
	ch := NewDeadChannel()

	if ch.Len() != 0 {
		t.Error("dead channel length not 0")
	}
	if ch.Cap() != 0 {
		t.Error("dead channel cap not 0")
	}

	select {
	case <-ch.Out():
		t.Error("read from a dead channel")
	default:
	}

	select {
	case ch.In() <- nil:
		t.Error("wrote to a dead channel")
	default:
	}

	ch.Close()

	ch = NewDeadChannel()
	testChannelConcurrentAccessors(t, "dead channel", ch)
}
