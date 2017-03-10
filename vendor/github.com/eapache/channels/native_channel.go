package channels

// NativeInChannel implements the InChannel interface by wrapping a native go write-only channel.
type NativeInChannel chan<- interface{}

func (ch NativeInChannel) In() chan<- interface{} {
	return ch
}

func (ch NativeInChannel) Len() int {
	return len(ch)
}

func (ch NativeInChannel) Cap() BufferCap {
	return BufferCap(cap(ch))
}

func (ch NativeInChannel) Close() {
	close(ch)
}

// NativeOutChannel implements the OutChannel interface by wrapping a native go read-only channel.
type NativeOutChannel <-chan interface{}

func (ch NativeOutChannel) Out() <-chan interface{} {
	return ch
}

func (ch NativeOutChannel) Len() int {
	return len(ch)
}

func (ch NativeOutChannel) Cap() BufferCap {
	return BufferCap(cap(ch))
}

// NativeChannel implements the Channel interface by wrapping a native go channel.
type NativeChannel chan interface{}

// NewNativeChannel makes a new NativeChannel with the given buffer size. Just a convenience wrapper
// to avoid having to cast the result of make().
func NewNativeChannel(size BufferCap) NativeChannel {
	return make(chan interface{}, size)
}

func (ch NativeChannel) In() chan<- interface{} {
	return ch
}

func (ch NativeChannel) Out() <-chan interface{} {
	return ch
}

func (ch NativeChannel) Len() int {
	return len(ch)
}

func (ch NativeChannel) Cap() BufferCap {
	return BufferCap(cap(ch))
}

func (ch NativeChannel) Close() {
	close(ch)
}

// DeadChannel is a placeholder implementation of the Channel interface with no buffer
// that is never ready for reading or writing. Closing a dead channel is a no-op.
// Behaves almost like NativeChannel(nil) except that closing a nil NativeChannel will panic.
type DeadChannel struct{}

func NewDeadChannel() DeadChannel {
	return DeadChannel{}
}

func (ch DeadChannel) In() chan<- interface{} {
	return nil
}

func (ch DeadChannel) Out() <-chan interface{} {
	return nil
}

func (ch DeadChannel) Len() int {
	return 0
}

func (ch DeadChannel) Cap() BufferCap {
	return BufferCap(0)
}

func (ch DeadChannel) Close() {
}
