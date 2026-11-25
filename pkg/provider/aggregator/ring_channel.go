package aggregator

import (
	"context"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// RingChannel implements a channel in a way that never blocks the writer.
// Specifically, if a value is written to a RingChannel when its buffer is full then the oldest
// value in the buffer is discarded to make room (just like a standard ring-buffer).
// Note that Go's scheduler can cause discarded values when they could be avoided, simply by scheduling
// the writer before the reader, so caveat emptor.
type RingChannel struct {
	input, output chan dynamic.Message
	buffer        *dynamic.Message
}

// inputBufferSize matches the configuration watcher channel buffer:
// providers send without watching for cancellation, so once ringBuffer has returned,
// a send in flight during shutdown must not block the pool from stopping.
const inputBufferSize = 100

func newRingChannel() *RingChannel {
	return &RingChannel{
		input:  make(chan dynamic.Message, inputBufferSize),
		output: make(chan dynamic.Message),
	}
}

func (ch *RingChannel) in() chan<- dynamic.Message {
	return ch.input
}

func (ch *RingChannel) out() <-chan dynamic.Message {
	return ch.output
}

// for all buffered cases.
func (ch *RingChannel) ringBuffer(ctx context.Context) {
	var input, output chan dynamic.Message
	var next dynamic.Message
	input = ch.input

	for input != nil || output != nil {
		select {
		// Prefer to write if possible, which is surprisingly effective in reducing
		// dropped elements due to overflow. The naive read/write select chooses randomly
		// when both channels are ready, which produces unnecessary drops 50% of the time.
		case output <- next:
			ch.buffer = nil
		default:
			select {
			case elem, open := <-input:
				if !open {
					input = nil
					break
				}

				ch.buffer = &elem
			case output <- next:
				ch.buffer = nil
			case <-ctx.Done():
				return
			}
		}

		if ch.buffer == nil {
			output = nil
			continue
		}

		output = ch.output
		next = *ch.buffer
	}

	close(ch.output)
}
