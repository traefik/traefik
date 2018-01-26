package snappy

import (
	"bytes"
	"encoding/binary"

	master "github.com/golang/snappy"
)

var xerialHeader = []byte{130, 83, 78, 65, 80, 80, 89, 0}

// Encode encodes data as snappy with no framing header.
func Encode(src []byte) []byte {
	return master.Encode(nil, src)
}

// Decode decodes snappy data whether it is traditional unframed
// or includes the xerial framing format.
func Decode(src []byte) ([]byte, error) {
	if !bytes.Equal(src[:8], xerialHeader) {
		return master.Decode(nil, src)
	}

	var (
		pos   = uint32(16)
		max   = uint32(len(src))
		dst   = make([]byte, 0, len(src))
		chunk []byte
		err   error
	)
	for pos < max {
		size := binary.BigEndian.Uint32(src[pos : pos+4])
		pos += 4

		chunk, err = master.Decode(chunk, src[pos:pos+size])
		if err != nil {
			return nil, err
		}
		pos += size
		dst = append(dst, chunk...)
	}
	return dst, nil
}
