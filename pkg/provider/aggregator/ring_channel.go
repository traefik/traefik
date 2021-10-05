package aggregator

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

// RingChannel implements the Channel interface in a way that never blocks the writer.
// Specifically, if a value is written to a RingChannel when its buffer is full then the oldest
// value in the buffer is discarded to make room (just like a standard ring-buffer).
// Note that Go's scheduler can cause discarded values when they could be avoided, simply by scheduling
// the writer before the reader, so caveat emptor.
// For the opposite behaviour (discarding the newest element, not the oldest) see OverflowingChannel.
type RingChannel struct {
	input, output chan dynamic.Message
	buffer        *queue
}

func NewRingChannel() *RingChannel {
	ch := &RingChannel{
		input:  make(chan dynamic.Message),
		output: make(chan dynamic.Message),
		buffer: newQueue(),
	}
	go ch.ringBuffer()
	return ch
}

func (ch *RingChannel) In() chan<- dynamic.Message {
	return ch.input
}

func (ch *RingChannel) Out() <-chan dynamic.Message {
	return ch.output
}

func (ch *RingChannel) Close() {
	close(ch.input)
}

// for all buffered cases
func (ch *RingChannel) ringBuffer() {
	var input, output chan dynamic.Message
	var next dynamic.Message
	input = ch.input

	for input != nil || output != nil {
		select {
		// Prefer to write if possible, which is surprisingly effective in reducing
		// dropped elements due to overflow. The naive read/write select chooses randomly
		// when both channels are ready, which produces unnecessary drops 50% of the time.
		case output <- next:
			ch.buffer.remove()
		default:
			select {
			case elem, open := <-input:
				if !open {
					input = nil
					break
				}

				ch.buffer.add(elem)
				if ch.buffer.length() > 1 {
					ch.buffer.remove()
				}
			case output <- next:
				ch.buffer.remove()
			}
		}

		if ch.buffer.length() == 0 {
			output = nil
			continue
		}

		output = ch.output
		next = ch.buffer.peek().(dynamic.Message)
	}

	close(ch.output)
}

// minQueueLen is smallest capacity that queue may have.
// Must be power of 2 for bitwise modulus: x % n == x & (n - 1).
const minQueueLen = 16

// queue represents a single instance of the queue data structure.
type queue struct {
	buf               []interface{}
	head, tail, count int
}

// newQueue constructs and returns a new queue.
func newQueue() *queue {
	return &queue{
		buf: make([]interface{}, minQueueLen),
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *queue) length() int {
	return q.count
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
func (q *queue) resize() {
	newBuf := make([]interface{}, q.count<<1)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// add puts an element on the end of the queue.
func (q *queue) add(elem interface{}) {
	if q.count == len(q.buf) {
		q.resize()
	}

	q.buf[q.tail] = elem
	// bitwise modulus
	q.tail = (q.tail + 1) & (len(q.buf) - 1)
	q.count++
}

// peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *queue) peek() interface{} {
	if q.count <= 0 {
		panic("queue: peek() called on empty queue")
	}
	return q.buf[q.head]
}

// remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *queue) remove() interface{} {
	if q.count <= 0 {
		panic("queue: remove() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	// bitwise modulus
	q.head = (q.head + 1) & (len(q.buf) - 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if len(q.buf) > minQueueLen && (q.count<<2) == len(q.buf) {
		q.resize()
	}
	return ret
}
