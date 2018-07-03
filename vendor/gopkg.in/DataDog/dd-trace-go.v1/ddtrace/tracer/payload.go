package tracer

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/tinylib/msgp/msgp"
)

// payload is a wrapper on top of the msgpack encoder which allows constructing an
// encoded array by pushing its entries sequentially, one at a time. It basically
// allows us to encode as we would with a stream, except that the contents of the stream
// can be read as a slice by the msgpack decoder at any time. It follows the guidelines
// from the msgpack array spec:
// https://github.com/msgpack/msgpack/blob/master/spec.md#array-format-family
//
// payload implements io.Reader and can be used with the decoder directly. To create
// a new payload use the newPayload method.
//
// payload is not safe for concurrent use.
//
// This structure basically allows us to push traces into the payload one at a time
// in order to always have knowledge of the payload size, but also making it possible
// for the agent to decode it as an array.
type payload struct {
	// header specifies the first few bytes in the msgpack stream
	// indicating the type of array (fixarray, array16 or array32)
	// and the number of items contained in the stream.
	header []byte

	// off specifies the current read position on the header.
	off int

	// count specifies the number of items in the stream.
	count uint64

	// buf holds the sequence of msgpack-encoded items.
	buf bytes.Buffer
}

var _ io.Reader = (*payload)(nil)

// newPayload returns a ready to use payload.
func newPayload() *payload {
	p := &payload{
		header: make([]byte, 8),
		off:    8,
	}
	return p
}

// push pushes a new item into the stream.
func (p *payload) push(t spanList) error {
	if err := msgp.Encode(&p.buf, t); err != nil {
		return err
	}
	p.count++
	p.updateHeader()
	return nil
}

// itemCount returns the number of items available in the srteam.
func (p *payload) itemCount() int {
	return int(p.count)
}

// size returns the payload size in bytes. After the first read the value becomes
// inaccurate by up to 8 bytes.
func (p *payload) size() int {
	return p.buf.Len() + len(p.header) - p.off
}

// reset resets the internal buffer, counter and read offset.
func (p *payload) reset() {
	p.off = 8
	p.count = 0
	p.buf.Reset()
}

// https://github.com/msgpack/msgpack/blob/master/spec.md#array-format-family
const (
	msgpackArrayFix byte = 144  // up to 15 items
	msgpackArray16       = 0xdc // up to 2^16-1 items, followed by size in 2 bytes
	msgpackArray32       = 0xdd // up to 2^32-1 items, followed by size in 4 bytes
)

// updateHeader updates the payload header based on the number of items currently
// present in the stream.
func (p *payload) updateHeader() {
	n := p.count
	switch {
	case n <= 15:
		p.header[7] = msgpackArrayFix + byte(n)
		p.off = 7
	case n <= 1<<16-1:
		binary.BigEndian.PutUint64(p.header, n) // writes 2 bytes
		p.header[5] = msgpackArray16
		p.off = 5
	default: // n <= 1<<32-1
		binary.BigEndian.PutUint64(p.header, n) // writes 4 bytes
		p.header[3] = msgpackArray32
		p.off = 3
	}
}

// Read implements io.Reader. It reads from the msgpack-encoded stream.
func (p *payload) Read(b []byte) (n int, err error) {
	if p.off < len(p.header) {
		// reading header
		n = copy(b, p.header[p.off:])
		p.off += n
		return n, nil
	}
	return p.buf.Read(b)
}
