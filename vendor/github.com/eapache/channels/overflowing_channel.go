package channels

import "github.com/eapache/queue"

// OverflowingChannel implements the Channel interface in a way that never blocks the writer.
// Specifically, if a value is written to an OverflowingChannel when its buffer is full
// (or, in an unbuffered case, when the recipient is not ready) then that value is simply discarded.
// Note that Go's scheduler can cause discarded values when they could be avoided, simply by scheduling
// the writer before the reader, so caveat emptor.
// For the opposite behaviour (discarding the oldest element, not the newest) see RingChannel.
type OverflowingChannel struct {
	input, output chan interface{}
	length        chan int
	buffer        *queue.Queue
	size          BufferCap
}

func NewOverflowingChannel(size BufferCap) *OverflowingChannel {
	if size < 0 && size != Infinity {
		panic("channels: invalid negative size in NewOverflowingChannel")
	}
	ch := &OverflowingChannel{
		input:  make(chan interface{}),
		output: make(chan interface{}),
		length: make(chan int),
		size:   size,
	}
	if size == None {
		go ch.overflowingDirect()
	} else {
		ch.buffer = queue.New()
		go ch.overflowingBuffer()
	}
	return ch
}

func (ch *OverflowingChannel) In() chan<- interface{} {
	return ch.input
}

func (ch *OverflowingChannel) Out() <-chan interface{} {
	return ch.output
}

func (ch *OverflowingChannel) Len() int {
	if ch.size == None {
		return 0
	} else {
		return <-ch.length
	}
}

func (ch *OverflowingChannel) Cap() BufferCap {
	return ch.size
}

func (ch *OverflowingChannel) Close() {
	close(ch.input)
}

// for entirely unbuffered cases
func (ch *OverflowingChannel) overflowingDirect() {
	for elem := range ch.input {
		// if we can't write it immediately, drop it and move on
		select {
		case ch.output <- elem:
		default:
		}
	}
	close(ch.output)
}

// for all buffered cases
func (ch *OverflowingChannel) overflowingBuffer() {
	var input, output chan interface{}
	var next interface{}
	input = ch.input

	for input != nil || output != nil {
		select {
		// Prefer to write if possible, which is surprisingly effective in reducing
		// dropped elements due to overflow. The naive read/write select chooses randomly
		// when both channels are ready, which produces unnecessary drops 50% of the time.
		case output <- next:
			ch.buffer.Remove()
		default:
			select {
			case elem, open := <-input:
				if open {
					if ch.size == Infinity || ch.buffer.Length() < int(ch.size) {
						ch.buffer.Add(elem)
					}
				} else {
					input = nil
				}
			case output <- next:
				ch.buffer.Remove()
			case ch.length <- ch.buffer.Length():
			}
		}

		if ch.buffer.Length() > 0 {
			output = ch.output
			next = ch.buffer.Peek()
		} else {
			output = nil
			next = nil
		}
	}

	close(ch.output)
	close(ch.length)
}
