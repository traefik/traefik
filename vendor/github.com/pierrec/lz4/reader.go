package lz4

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
	"sync/atomic"
)

// ErrInvalid is returned when the data being read is not an LZ4 archive
// (LZ4 magic number detection failed).
var ErrInvalid = errors.New("invalid lz4 data")

// errEndOfBlock is returned by readBlock when it has reached the last block of the frame.
// It is not an error.
var errEndOfBlock = errors.New("end of block")

// Reader implements the LZ4 frame decoder.
// The Header is set after the first call to Read().
// The Header may change between Read() calls in case of concatenated frames.
type Reader struct {
	Pos int64 // position within the source
	Header
	src      io.Reader
	checksum hash.Hash32    // frame hash
	wg       sync.WaitGroup // decompressing go routine wait group
	data     []byte         // buffered decompressed data
	window   []byte         // 64Kb decompressed data window
}

// NewReader returns a new LZ4 frame decoder.
// No access to the underlying io.Reader is performed.
func NewReader(src io.Reader) *Reader {
	return &Reader{
		src:      src,
		checksum: hashPool.Get(),
	}
}

// readHeader checks the frame magic number and parses the frame descriptoz.
// Skippable frames are supported even as a first frame although the LZ4
// specifications recommends skippable frames not to be used as first frames.
func (z *Reader) readHeader(first bool) error {
	defer z.checksum.Reset()

	for {
		var magic uint32
		if err := binary.Read(z.src, binary.LittleEndian, &magic); err != nil {
			if !first && err == io.ErrUnexpectedEOF {
				return io.EOF
			}
			return err
		}
		z.Pos += 4
		if magic>>8 == frameSkipMagic>>8 {
			var skipSize uint32
			if err := binary.Read(z.src, binary.LittleEndian, &skipSize); err != nil {
				return err
			}
			z.Pos += 4
			m, err := io.CopyN(ioutil.Discard, z.src, int64(skipSize))
			z.Pos += m
			if err != nil {
				return err
			}
			continue
		}
		if magic != frameMagic {
			return ErrInvalid
		}
		break
	}

	// header
	var buf [8]byte
	if _, err := io.ReadFull(z.src, buf[:2]); err != nil {
		return err
	}
	z.Pos += 2

	b := buf[0]
	if b>>6 != Version {
		return fmt.Errorf("lz4.Read: invalid version: got %d expected %d", b>>6, Version)
	}
	z.BlockDependency = b>>5&1 == 0
	z.BlockChecksum = b>>4&1 > 0
	frameSize := b>>3&1 > 0
	z.NoChecksum = b>>2&1 == 0
	// 	z.Dict = b&1 > 0

	bmsID := buf[1] >> 4 & 0x7
	bSize, ok := bsMapID[bmsID]
	if !ok {
		return fmt.Errorf("lz4.Read: invalid block max size: %d", bmsID)
	}
	z.BlockMaxSize = bSize

	z.checksum.Write(buf[0:2])

	if frameSize {
		if err := binary.Read(z.src, binary.LittleEndian, &z.Size); err != nil {
			return err
		}
		z.Pos += 8
		binary.LittleEndian.PutUint64(buf[:], z.Size)
		z.checksum.Write(buf[0:8])
	}

	// 	if z.Dict {
	// 		if err := binary.Read(z.src, binary.LittleEndian, &z.DictID); err != nil {
	// 			return err
	// 		}
	// 		z.Pos += 4
	// 		binary.LittleEndian.PutUint32(buf[:], z.DictID)
	// 		z.checksum.Write(buf[0:4])
	// 	}

	// header checksum
	if _, err := io.ReadFull(z.src, buf[:1]); err != nil {
		return err
	}
	z.Pos++
	if h := byte(z.checksum.Sum32() >> 8 & 0xFF); h != buf[0] {
		return fmt.Errorf("lz4.Read: invalid header checksum: got %v expected %v", buf[0], h)
	}

	z.Header.done = true

	return nil
}

// Read decompresses data from the underlying source into the supplied buffer.
//
// Since there can be multiple streams concatenated, Header values may
// change between calls to Read(). If that is the case, no data is actually read from
// the underlying io.Reader, to allow for potential input buffer resizing.
//
// Data is buffered if the input buffer is too small, and exhausted upon successive calls.
//
// If the buffer is large enough (typically in multiples of BlockMaxSize) and there is
// no block dependency, then the data will be decompressed concurrently based on the GOMAXPROCS value.
func (z *Reader) Read(buf []byte) (n int, err error) {
	if !z.Header.done {
		if err = z.readHeader(true); err != nil {
			return
		}
	}

	if len(buf) == 0 {
		return
	}

	// exhaust remaining data from previous Read()
	if len(z.data) > 0 {
		n = copy(buf, z.data)
		z.data = z.data[n:]
		if len(z.data) == 0 {
			z.data = nil
		}
		return
	}

	// Break up the input buffer into BlockMaxSize blocks with at least one block.
	// Then decompress into each of them concurrently if possible (no dependency).
	// In case of dependency, the first block will be missing the window (except on the
	// very first call), the rest will have it already since it comes from the previous block.
	wbuf := buf
	zn := (len(wbuf) + z.BlockMaxSize - 1) / z.BlockMaxSize
	zblocks := make([]block, zn)
	for zi, abort := 0, uint32(0); zi < zn && atomic.LoadUint32(&abort) == 0; zi++ {
		zb := &zblocks[zi]
		// last block may be too small
		if len(wbuf) < z.BlockMaxSize+len(z.window) {
			wbuf = make([]byte, z.BlockMaxSize+len(z.window))
		}
		copy(wbuf, z.window)
		if zb.err = z.readBlock(wbuf, zb); zb.err != nil {
			break
		}
		wbuf = wbuf[z.BlockMaxSize:]
		if !z.BlockDependency {
			z.wg.Add(1)
			go z.decompressBlock(zb, &abort)
			continue
		}
		// cannot decompress concurrently when dealing with block dependency
		z.decompressBlock(zb, nil)
		// the last block may not contain enough data
		if len(z.window) == 0 {
			z.window = make([]byte, winSize)
		}
		if len(zb.data) >= winSize {
			copy(z.window, zb.data[len(zb.data)-winSize:])
		} else {
			copy(z.window, z.window[len(zb.data):])
			copy(z.window[len(zb.data)+1:], zb.data)
		}
	}
	z.wg.Wait()

	// since a block size may be less then BlockMaxSize, trim the decompressed buffers
	for _, zb := range zblocks {
		if zb.err != nil {
			if zb.err == errEndOfBlock {
				return n, z.close()
			}
			return n, zb.err
		}
		bLen := len(zb.data)
		if !z.NoChecksum {
			z.checksum.Write(zb.data)
		}
		m := copy(buf[n:], zb.data)
		// buffer the remaining data (this is necessarily the last block)
		if m < bLen {
			z.data = zb.data[m:]
		}
		n += m
	}

	return
}

// readBlock reads an entire frame block from the frame.
// The input buffer is the one that will receive the decompressed data.
// If the end of the frame is detected, it returns the errEndOfBlock error.
func (z *Reader) readBlock(buf []byte, b *block) error {
	var bLen uint32
	if err := binary.Read(z.src, binary.LittleEndian, &bLen); err != nil {
		return err
	}
	atomic.AddInt64(&z.Pos, 4)

	switch {
	case bLen == 0:
		return errEndOfBlock
	case bLen&(1<<31) == 0:
		b.compressed = true
		b.data = buf
		b.zdata = make([]byte, bLen)
	default:
		bLen = bLen & (1<<31 - 1)
		if int(bLen) > len(buf) {
			return fmt.Errorf("lz4.Read: invalid block size: %d", bLen)
		}
		b.data = buf[:bLen]
		b.zdata = buf[:bLen]
	}
	if _, err := io.ReadFull(z.src, b.zdata); err != nil {
		return err
	}

	if z.BlockChecksum {
		if err := binary.Read(z.src, binary.LittleEndian, &b.checksum); err != nil {
			return err
		}
		xxh := hashPool.Get()
		defer hashPool.Put(xxh)
		xxh.Write(b.zdata)
		if h := xxh.Sum32(); h != b.checksum {
			return fmt.Errorf("lz4.Read: invalid block checksum: got %x expected %x", h, b.checksum)
		}
	}

	return nil
}

// decompressBlock decompresses a frame block.
// In case of an error, the block err is set with it and abort is set to 1.
func (z *Reader) decompressBlock(b *block, abort *uint32) {
	if abort != nil {
		defer z.wg.Done()
	}
	if b.compressed {
		n := len(z.window)
		m, err := UncompressBlock(b.zdata, b.data, n)
		if err != nil {
			if abort != nil {
				atomic.StoreUint32(abort, 1)
			}
			b.err = err
			return
		}
		b.data = b.data[n : n+m]
	}
	atomic.AddInt64(&z.Pos, int64(len(b.data)))
}

// close validates the frame checksum (if any) and checks the next frame (if any).
func (z *Reader) close() error {
	if !z.NoChecksum {
		var checksum uint32
		if err := binary.Read(z.src, binary.LittleEndian, &checksum); err != nil {
			return err
		}
		if checksum != z.checksum.Sum32() {
			return fmt.Errorf("lz4.Read: invalid frame checksum: got %x expected %x", z.checksum.Sum32(), checksum)
		}
	}

	// get ready for the next concatenated frame, but do not change the position
	pos := z.Pos
	z.Reset(z.src)
	z.Pos = pos

	// since multiple frames can be concatenated, check for another one
	return z.readHeader(false)
}

// Reset discards the Reader's state and makes it equivalent to the
// result of its original state from NewReader, but reading from r instead.
// This permits reusing a Reader rather than allocating a new one.
func (z *Reader) Reset(r io.Reader) {
	z.Header = Header{}
	z.Pos = 0
	z.src = r
	z.checksum.Reset()
	z.data = nil
	z.window = nil
}

// WriteTo decompresses the data from the underlying io.Reader and writes it to the io.Writer.
// Returns the number of bytes written.
func (z *Reader) WriteTo(w io.Writer) (n int64, err error) {
	cpus := runtime.GOMAXPROCS(0)
	var buf []byte

	// The initial buffer being nil, the first Read will be only read the compressed frame options.
	// The buffer can then be sized appropriately to support maximum concurrency decompression.
	// If multiple frames are concatenated, Read() will return with no data decompressed but with
	// potentially changed options. The buffer will be resized accordingly, always trying to
	// maximize concurrency.
	for {
		nsize := 0
		// the block max size can change if multiple streams are concatenated.
		// Check it after every Read().
		if z.BlockDependency {
			// in case of dependency, we cannot decompress concurrently,
			// so allocate the minimum buffer + window size
			nsize = len(z.window) + z.BlockMaxSize
		} else {
			// if no dependency, allocate a buffer large enough for concurrent decompression
			nsize = cpus * z.BlockMaxSize
		}
		if nsize != len(buf) {
			buf = make([]byte, nsize)
		}

		m, er := z.Read(buf)
		if er != nil && er != io.EOF {
			return n, er
		}
		m, err = w.Write(buf[:m])
		n += int64(m)
		if err != nil || er == io.EOF {
			return
		}
	}
}
