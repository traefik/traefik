/*
Package channels provides a collection of helper functions, interfaces and implementations for
working with and extending the capabilities of golang's existing channels. The main interface of
interest is Channel, though sub-interfaces are also provided for cases where the full Channel interface
cannot be met (for example, InChannel for write-only channels).

For integration with native typed golang channels, functions Wrap and Unwrap are provided which do the
appropriate type conversions. The NativeChannel, NativeInChannel and NativeOutChannel type definitions
are also provided for use with native channels which already carry values of type interface{}.

The heart of the package consists of several distinct implementations of the Channel interface, including
channels backed by special buffers (resizable, infinite, ring buffers, etc) and other useful types. A
"black hole" channel for discarding unwanted values (similar in purpose to ioutil.Discard or /dev/null)
rounds out the set.

Helper functions for operating on Channels include Pipe and Tee (which behave much like their Unix
namesakes), as well as Multiplex and Distribute. "Weak" versions of these functions also exist, which
do not close their output channel(s) on completion.

Due to limitations of Go's type system, importing this library directly is often not practical for
production code. It serves equally well, however, as a reference guide and template for implementing
many common idioms; if you use it in this way I would appreciate the inclusion of some sort of credit
in the resulting code.

Warning: several types in this package provide so-called "infinite" buffers. Be *very* careful using
these, as no buffer is truly infinite - if such a buffer grows too large your program will run out of
memory and crash. Caveat emptor.
*/
package channels

import "reflect"

// BufferCap represents the capacity of the buffer backing a channel. Valid values consist of all
// positive integers, as well as the special values below.
type BufferCap int

const (
	// None is the capacity for channels that have no buffer at all.
	None BufferCap = 0
	// Infinity is the capacity for channels with no limit on their buffer size.
	Infinity BufferCap = -1
)

// Buffer is an interface for any channel that provides access to query the state of its buffer.
// Even unbuffered channels can implement this interface by simply returning 0 from Len() and None from Cap().
type Buffer interface {
	Len() int       // The number of elements currently buffered.
	Cap() BufferCap // The maximum number of elements that can be buffered.
}

// SimpleInChannel is an interface representing a writeable channel that does not necessarily
// implement the Buffer interface.
type SimpleInChannel interface {
	In() chan<- interface{} // The writeable end of the channel.
	Close()                 // Closes the channel. It is an error to write to In() after calling Close().
}

// InChannel is an interface representing a writeable channel with a buffer.
type InChannel interface {
	SimpleInChannel
	Buffer
}

// SimpleOutChannel is an interface representing a readable channel that does not necessarily
// implement the Buffer interface.
type SimpleOutChannel interface {
	Out() <-chan interface{} // The readable end of the channel.
}

// OutChannel is an interface representing a readable channel implementing the Buffer interface.
type OutChannel interface {
	SimpleOutChannel
	Buffer
}

// SimpleChannel is an interface representing a channel that is both readable and writeable,
// but does not necessarily implement the Buffer interface.
type SimpleChannel interface {
	SimpleInChannel
	SimpleOutChannel
}

// Channel is an interface representing a channel that is readable, writeable and implements
// the Buffer interface
type Channel interface {
	SimpleChannel
	Buffer
}

func pipe(input SimpleOutChannel, output SimpleInChannel, closeWhenDone bool) {
	for elem := range input.Out() {
		output.In() <- elem
	}
	if closeWhenDone {
		output.Close()
	}
}

func multiplex(output SimpleInChannel, inputs []SimpleOutChannel, closeWhenDone bool) {
	inputCount := len(inputs)
	cases := make([]reflect.SelectCase, inputCount)
	for i := range cases {
		cases[i].Dir = reflect.SelectRecv
		cases[i].Chan = reflect.ValueOf(inputs[i].Out())
	}
	for inputCount > 0 {
		chosen, recv, recvOK := reflect.Select(cases)
		if recvOK {
			output.In() <- recv.Interface()
		} else {
			cases[chosen].Chan = reflect.ValueOf(nil)
			inputCount--
		}
	}
	if closeWhenDone {
		output.Close()
	}
}

func tee(input SimpleOutChannel, outputs []SimpleInChannel, closeWhenDone bool) {
	cases := make([]reflect.SelectCase, len(outputs))
	for i := range cases {
		cases[i].Dir = reflect.SelectSend
	}
	for elem := range input.Out() {
		for i := range cases {
			cases[i].Chan = reflect.ValueOf(outputs[i].In())
			cases[i].Send = reflect.ValueOf(elem)
		}
		for _ = range cases {
			chosen, _, _ := reflect.Select(cases)
			cases[chosen].Chan = reflect.ValueOf(nil)
		}
	}
	if closeWhenDone {
		for i := range outputs {
			outputs[i].Close()
		}
	}
}

func distribute(input SimpleOutChannel, outputs []SimpleInChannel, closeWhenDone bool) {
	cases := make([]reflect.SelectCase, len(outputs))
	for i := range cases {
		cases[i].Dir = reflect.SelectSend
		cases[i].Chan = reflect.ValueOf(outputs[i].In())
	}
	for elem := range input.Out() {
		for i := range cases {
			cases[i].Send = reflect.ValueOf(elem)
		}
		reflect.Select(cases)
	}
	if closeWhenDone {
		for i := range outputs {
			outputs[i].Close()
		}
	}
}

// Pipe connects the input channel to the output channel so that
// they behave as if a single channel.
func Pipe(input SimpleOutChannel, output SimpleInChannel) {
	go pipe(input, output, true)
}

// Multiplex takes an arbitrary number of input channels and multiplexes their output into a single output
// channel. When all input channels have been closed, the output channel is closed. Multiplex with a single
// input channel is equivalent to Pipe (though slightly less efficient).
func Multiplex(output SimpleInChannel, inputs ...SimpleOutChannel) {
	if len(inputs) == 0 {
		panic("channels: Multiplex requires at least one input")
	}
	go multiplex(output, inputs, true)
}

// Tee (like its Unix namesake) takes a single input channel and an arbitrary number of output channels
// and duplicates each input into every output. When the input channel is closed, all outputs channels are closed.
// Tee with a single output channel is equivalent to Pipe (though slightly less efficient).
func Tee(input SimpleOutChannel, outputs ...SimpleInChannel) {
	if len(outputs) == 0 {
		panic("channels: Tee requires at least one output")
	}
	go tee(input, outputs, true)
}

// Distribute takes a single input channel and an arbitrary number of output channels and duplicates each input
// into *one* available output. If multiple outputs are waiting for a value, one is chosen at random. When the
// input channel is closed, all outputs channels are closed. Distribute with a single output channel is
// equivalent to Pipe (though slightly less efficient).
func Distribute(input SimpleOutChannel, outputs ...SimpleInChannel) {
	if len(outputs) == 0 {
		panic("channels: Distribute requires at least one output")
	}
	go distribute(input, outputs, true)
}

// WeakPipe behaves like Pipe (connecting the two channels) except that it does not close
// the output channel when the input channel is closed.
func WeakPipe(input SimpleOutChannel, output SimpleInChannel) {
	go pipe(input, output, false)
}

// WeakMultiplex behaves like Multiplex (multiplexing multiple inputs into a single output) except that it does not close
// the output channel when the input channels are closed.
func WeakMultiplex(output SimpleInChannel, inputs ...SimpleOutChannel) {
	if len(inputs) == 0 {
		panic("channels: WeakMultiplex requires at least one input")
	}
	go multiplex(output, inputs, false)
}

// WeakTee behaves like Tee (duplicating a single input into multiple outputs) except that it does not close
// the output channels when the input channel is closed.
func WeakTee(input SimpleOutChannel, outputs ...SimpleInChannel) {
	if len(outputs) == 0 {
		panic("channels: WeakTee requires at least one output")
	}
	go tee(input, outputs, false)
}

// WeakDistribute behaves like Distribute (distributing a single input amongst multiple outputs) except that
// it does not close the output channels when the input channel is closed.
func WeakDistribute(input SimpleOutChannel, outputs ...SimpleInChannel) {
	if len(outputs) == 0 {
		panic("channels: WeakDistribute requires at least one output")
	}
	go distribute(input, outputs, false)
}

// Wrap takes any readable channel type (chan or <-chan but not chan<-) and
// exposes it as a SimpleOutChannel for easy integration with existing channel sources.
// It panics if the input is not a readable channel.
func Wrap(ch interface{}) SimpleOutChannel {
	t := reflect.TypeOf(ch)
	if t.Kind() != reflect.Chan || t.ChanDir()&reflect.RecvDir == 0 {
		panic("channels: input to Wrap must be readable channel")
	}
	realChan := make(chan interface{})

	go func() {
		v := reflect.ValueOf(ch)
		for {
			x, ok := v.Recv()
			if !ok {
				close(realChan)
				return
			}
			realChan <- x.Interface()
		}
	}()

	return NativeOutChannel(realChan)
}

// Unwrap takes a SimpleOutChannel and uses reflection to pipe it to a typed native channel for
// easy integration with existing channel sources. Output can be any writable channel type (chan or chan<-).
// It panics if the output is not a writable channel, or if a value is received that cannot be sent on the
// output channel.
func Unwrap(input SimpleOutChannel, output interface{}) {
	t := reflect.TypeOf(output)
	if t.Kind() != reflect.Chan || t.ChanDir()&reflect.SendDir == 0 {
		panic("channels: input to Unwrap must be readable channel")
	}

	go func() {
		v := reflect.ValueOf(output)
		for {
			x, ok := <-input.Out()
			if !ok {
				v.Close()
				return
			}
			v.Send(reflect.ValueOf(x))
		}
	}()
}
