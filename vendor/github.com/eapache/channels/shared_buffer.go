package channels

import (
	"reflect"

	"github.com/eapache/queue"
)

//sharedBufferChannel implements SimpleChannel and is created by the public
//SharedBuffer type below
type sharedBufferChannel struct {
	in     chan interface{}
	out    chan interface{}
	buf    *queue.Queue
	closed bool
}

func (sch *sharedBufferChannel) In() chan<- interface{} {
	return sch.in
}

func (sch *sharedBufferChannel) Out() <-chan interface{} {
	return sch.out
}

func (sch *sharedBufferChannel) Close() {
	close(sch.in)
}

//SharedBuffer implements the Buffer interface, and permits multiple SimpleChannel instances to "share" a single buffer.
//Each channel spawned by NewChannel has its own internal queue (so values flowing through do not get mixed up with
//other channels) but the total number of elements buffered by all spawned channels is limited to a single capacity. This
//means *all* such channels block and unblock for writing together. The primary use case is for implementing pipeline-style
//parallelism with goroutines, limiting the total number of elements in the pipeline without limiting the number of elements
//at any particular step.
type SharedBuffer struct {
	cases []reflect.SelectCase   // 2n+1 of these; [0] is for control, [1,3,5...] for recv, [2,4,6...] for send
	chans []*sharedBufferChannel // n of these
	count int
	size  BufferCap
	in    chan *sharedBufferChannel
}

func NewSharedBuffer(size BufferCap) *SharedBuffer {
	if size < 0 && size != Infinity {
		panic("channels: invalid negative size in NewSharedBuffer")
	} else if size == None {
		panic("channels: SharedBuffer does not support unbuffered behaviour")
	}

	buf := &SharedBuffer{
		size: size,
		in:   make(chan *sharedBufferChannel),
	}

	buf.cases = append(buf.cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(buf.in),
	})

	go buf.mainLoop()

	return buf
}

//NewChannel spawns and returns a new channel sharing the underlying buffer.
func (buf *SharedBuffer) NewChannel() SimpleChannel {
	ch := &sharedBufferChannel{
		in:  make(chan interface{}),
		out: make(chan interface{}),
		buf: queue.New(),
	}
	buf.in <- ch
	return ch
}

//Close shuts down the SharedBuffer. It is an error to call Close while channels are still using
//the buffer (I'm not really sure what would happen if you do so).
func (buf *SharedBuffer) Close() {
	// TODO: what if there are still active channels using this buffer?
	close(buf.in)
}

func (buf *SharedBuffer) mainLoop() {
	for {
		i, val, ok := reflect.Select(buf.cases)

		if i == 0 {
			if !ok {
				//Close was called on the SharedBuffer itself
				return
			}

			//NewChannel was called on the SharedBuffer
			ch := val.Interface().(*sharedBufferChannel)
			buf.chans = append(buf.chans, ch)
			buf.cases = append(buf.cases,
				reflect.SelectCase{Dir: reflect.SelectRecv},
				reflect.SelectCase{Dir: reflect.SelectSend},
			)
			if buf.size == Infinity || buf.count < int(buf.size) {
				buf.cases[len(buf.cases)-2].Chan = reflect.ValueOf(ch.in)
			}
		} else if i%2 == 0 {
			//Send
			if buf.count == int(buf.size) {
				//room in the buffer again, re-enable all recv cases
				for j := range buf.chans {
					if !buf.chans[j].closed {
						buf.cases[(j*2)+1].Chan = reflect.ValueOf(buf.chans[j].in)
					}
				}
			}
			buf.count--
			ch := buf.chans[(i-1)/2]
			if ch.buf.Length() > 0 {
				buf.cases[i].Send = reflect.ValueOf(ch.buf.Peek())
				ch.buf.Remove()
			} else {
				//nothing left for this channel to send, disable sending
				buf.cases[i].Chan = reflect.Value{}
				buf.cases[i].Send = reflect.Value{}
				if ch.closed {
					// and it was closed, so close the output channel
					//TODO: shrink slice
					close(ch.out)
				}
			}
		} else {
			ch := buf.chans[i/2]
			if ok {
				//Receive
				buf.count++
				if ch.buf.Length() == 0 && !buf.cases[i+1].Chan.IsValid() {
					//this channel now has something to send
					buf.cases[i+1].Chan = reflect.ValueOf(ch.out)
					buf.cases[i+1].Send = val
				} else {
					ch.buf.Add(val.Interface())
				}
				if buf.count == int(buf.size) {
					//buffer full, disable recv cases
					for j := range buf.chans {
						buf.cases[(j*2)+1].Chan = reflect.Value{}
					}
				}
			} else {
				//Close
				buf.cases[i].Chan = reflect.Value{}
				ch.closed = true
				if ch.buf.Length() == 0 && !buf.cases[i+1].Chan.IsValid() {
					//nothing pending, close the out channel right away
					//TODO: shrink slice
					close(ch.out)
				}
			}
		}
	}
}

func (buf *SharedBuffer) Len() int {
	return buf.count
}

func (buf *SharedBuffer) Cap() BufferCap {
	return buf.size
}
