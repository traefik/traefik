package channels

import "github.com/eapache/queue"

// ResizableChannel implements the Channel interface with a resizable buffer between the input and the output.
// The channel initially has a buffer size of 1, but can be resized by calling Resize().
//
// Resizing to a buffer capacity of None is, unfortunately, not supported and will panic
// (see https://github.com/eapache/channels/issues/1).
// Resizing back and forth between a finite and infinite buffer is fully supported.
type ResizableChannel struct {
	input, output    chan interface{}
	length           chan int
	capacity, resize chan BufferCap
	size             BufferCap
	buffer           *queue.Queue
}

func NewResizableChannel() *ResizableChannel {
	ch := &ResizableChannel{
		input:    make(chan interface{}),
		output:   make(chan interface{}),
		length:   make(chan int),
		capacity: make(chan BufferCap),
		resize:   make(chan BufferCap),
		size:     1,
		buffer:   queue.New(),
	}
	go ch.magicBuffer()
	return ch
}

func (ch *ResizableChannel) In() chan<- interface{} {
	return ch.input
}

func (ch *ResizableChannel) Out() <-chan interface{} {
	return ch.output
}

func (ch *ResizableChannel) Len() int {
	return <-ch.length
}

func (ch *ResizableChannel) Cap() BufferCap {
	val, open := <-ch.capacity
	if open {
		return val
	} else {
		return ch.size
	}
}

func (ch *ResizableChannel) Close() {
	close(ch.input)
}

func (ch *ResizableChannel) Resize(newSize BufferCap) {
	if newSize == None {
		panic("channels: ResizableChannel does not support unbuffered behaviour")
	}
	if newSize < 0 && newSize != Infinity {
		panic("channels: invalid negative size trying to resize channel")
	}
	ch.resize <- newSize
}

func (ch *ResizableChannel) magicBuffer() {
	var input, output, nextInput chan interface{}
	var next interface{}
	nextInput = ch.input
	input = nextInput

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				ch.buffer.Add(elem)
			} else {
				input = nil
				nextInput = nil
			}
		case output <- next:
			ch.buffer.Remove()
		case ch.size = <-ch.resize:
		case ch.length <- ch.buffer.Length():
		case ch.capacity <- ch.size:
		}

		if ch.buffer.Length() == 0 {
			output = nil
			next = nil
		} else {
			output = ch.output
			next = ch.buffer.Peek()
		}

		if ch.size != Infinity && ch.buffer.Length() >= int(ch.size) {
			input = nil
		} else {
			input = nextInput
		}
	}

	close(ch.output)
	close(ch.resize)
	close(ch.length)
	close(ch.capacity)
}
