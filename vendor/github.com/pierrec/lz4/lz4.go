// Package lz4 implements reading and writing lz4 compressed data (a frame),
// as specified in http://fastcompression.blogspot.fr/2013/04/lz4-streaming-format-final.html,
// using an io.Reader (decompression) and io.Writer (compression).
// It is designed to minimize memory usage while maximizing throughput by being able to
// [de]compress data concurrently.
//
// The Reader and the Writer support concurrent processing provided the supplied buffers are
// large enough (in multiples of BlockMaxSize) and there is no block dependency.
// Reader.WriteTo and Writer.ReadFrom do leverage the concurrency transparently.
// The runtime.GOMAXPROCS() value is used to apply concurrency or not.
//
// Although the block level compression and decompression functions are exposed and are fully compatible
// with the lz4 block format definition, they are low level and should not be used directly.
// For a complete description of an lz4 compressed block, see:
// http://fastcompression.blogspot.fr/2011/05/lz4-explained.html
//
// See https://github.com/Cyan4973/lz4 for the reference C implementation.
package lz4

import (
	"hash"
	"sync"

	"github.com/pierrec/xxHash/xxHash32"
)

const (
	// Extension is the LZ4 frame file name extension
	Extension = ".lz4"
	// Version is the LZ4 frame format version
	Version = 1

	frameMagic     = uint32(0x184D2204)
	frameSkipMagic = uint32(0x184D2A50)

	// The following constants are used to setup the compression algorithm.
	minMatch   = 4  // the minimum size of the match sequence size (4 bytes)
	winSizeLog = 16 // LZ4 64Kb window size limit
	winSize    = 1 << winSizeLog
	winMask    = winSize - 1 // 64Kb window of previous data for dependent blocks

	// hashLog determines the size of the hash table used to quickly find a previous match position.
	// Its value influences the compression speed and memory usage, the lower the faster,
	// but at the expense of the compression ratio.
	// 16 seems to be the best compromise.
	hashLog       = 16
	hashTableSize = 1 << hashLog
	hashShift     = uint((minMatch * 8) - hashLog)

	mfLimit      = 8 + minMatch // The last match cannot start within the last 12 bytes.
	skipStrength = 6            // variable step for fast scan

	hasher = uint32(2654435761) // prime number used to hash minMatch
)

// map the block max size id with its value in bytes: 64Kb, 256Kb, 1Mb and 4Mb.
var bsMapID = map[byte]int{4: 64 << 10, 5: 256 << 10, 6: 1 << 20, 7: 4 << 20}
var bsMapValue = map[int]byte{}

// Reversed.
func init() {
	for i, v := range bsMapID {
		bsMapValue[v] = i
	}
}

// Header describes the various flags that can be set on a Writer or obtained from a Reader.
// The default values match those of the LZ4 frame format definition (http://fastcompression.blogspot.com/2013/04/lz4-streaming-format-final.html).
//
// NB. in a Reader, in case of concatenated frames, the Header values may change between Read() calls.
// It is the caller responsibility to check them if necessary (typically when using the Reader concurrency).
type Header struct {
	BlockDependency bool   // compressed blocks are dependent (one block depends on the last 64Kb of the previous one)
	BlockChecksum   bool   // compressed blocks are checksumed
	NoChecksum      bool   // frame checksum
	BlockMaxSize    int    // the size of the decompressed data block (one of [64KB, 256KB, 1MB, 4MB]). Default=4MB.
	Size            uint64 // the frame total size. It is _not_ computed by the Writer.
	HighCompression bool   // use high compression (only for the Writer)
	done            bool   // whether the descriptor was processed (Read or Write and checked)
	// Removed as not supported
	// 	Dict            bool   // a dictionary id is to be used
	// 	DictID          uint32 // the dictionary id read from the frame, if any.
}

// xxhPool wraps the standard pool for xxHash items.
// Putting items back in the pool automatically resets them.
type xxhPool struct {
	sync.Pool
}

func (p *xxhPool) Get() hash.Hash32 {
	return p.Pool.Get().(hash.Hash32)
}

func (p *xxhPool) Put(h hash.Hash32) {
	h.Reset()
	p.Pool.Put(h)
}

// hashPool is used by readers and writers and contains xxHash items.
var hashPool = xxhPool{
	Pool: sync.Pool{
		New: func() interface{} { return xxHash32.New(0) },
	},
}
