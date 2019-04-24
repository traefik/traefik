[![godoc](https://godoc.org/github.com/pierrec/lz4?status.png)](https://godoc.org/github.com/pierrec/lz4)
[![Build Status](https://travis-ci.org/pierrec/lz4.svg?branch=master)](https://travis-ci.org/pierrec/lz4)

# lz4
LZ4 compression and decompression in pure Go

## Usage

```go
import "github.com/pierrec/lz4"
```

## Description

Package lz4 implements reading and writing lz4 compressed data (a frame),
as specified in http://fastcompression.blogspot.fr/2013/04/lz4-streaming-format-final.html,
using an io.Reader (decompression) and io.Writer (compression).
It is designed to minimize memory usage while maximizing throughput by being able to
[de]compress data concurrently.

The Reader and the Writer support concurrent processing provided the supplied buffers are
large enough (in multiples of BlockMaxSize) and there is no block dependency.
Reader.WriteTo and Writer.ReadFrom do leverage the concurrency transparently.
The runtime.GOMAXPROCS() value is used to apply concurrency or not.

Although the block level compression and decompression functions are exposed and are fully compatible
with the lz4 block format definition, they are low level and should not be used directly.
For a complete description of an lz4 compressed block, see:
http://fastcompression.blogspot.fr/2011/05/lz4-explained.html

See https://github.com/Cyan4973/lz4 for the reference C implementation.
