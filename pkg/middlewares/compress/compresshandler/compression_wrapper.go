package compresshandler

import (
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type Algorithm string

const (
	Undefined Algorithm = "N/A"
	Brotli    Algorithm = "br"
	Zstandard Algorithm = "zstd"
)

type CompressionWriter interface {
	// Write data to the encoder.
	// Input data will be buffered and as the buffer fills up
	// content will be compressed and written to the output.
	// When done writing, use Close to flush the remaining output
	// and write CRC if requested.
	Write(p []byte) (n int, err error)
	// Flush will send the currently written data to output
	// and block until everything has been written.
	// This should only be used on rare occasions where pushing the currently queued data is critical.
	Flush() error
	// Close closes the underlying writers if/when appropriate.
	// Note that the compressed writer should not be closed if we never used it,
	// as it would otherwise send some extra "end of compression" bytes.
	// Close also makes sure to flush whatever was left to write from the buffer.
	Close() error
	// ContentEncoding content encoding of the compression for HTTP Header
	ContentEncoding() string
}

type zstdWriter struct {
	zstd *zstd.Encoder
}

type brotliWriter struct {
	brotli *brotli.Writer
}

func NewCompressionWriter(algo Algorithm, in io.Writer) (CompressionWriter, error) {
	switch algo {
	case Brotli:
		return NewBrWriter(in)
	case Zstandard:
		return NewZstdWriter(in)
	default:
		return nil, fmt.Errorf("unknown compression algo: %s", algo)
	}
}

func NewZstdWriter(in io.Writer) (CompressionWriter, error) {
	writer, err := zstd.NewWriter(in)
	if err != nil {
		return nil, err
	}

	return &zstdWriter{
		zstd: writer,
	}, nil
}

func (z *zstdWriter) Close() error {
	return z.zstd.Close()
}

func (z *zstdWriter) Flush() error {
	return z.zstd.Flush()
}

func (z *zstdWriter) Write(p []byte) (n int, err error) {
	return z.zstd.Write(p)
}

func (z *zstdWriter) ContentEncoding() string {
	return string(Zstandard)
}

func NewBrWriter(in io.Writer) (CompressionWriter, error) {
	writer := brotli.NewWriter(in)
	return &brotliWriter{
		brotli: writer,
	}, nil
}

func (z *brotliWriter) Close() error {
	return z.brotli.Close()
}

func (z *brotliWriter) Flush() error {
	return z.brotli.Flush()
}

func (z *brotliWriter) Write(p []byte) (n int, err error) {
	return z.brotli.Write(p)
}

func (z *brotliWriter) ContentEncoding() string {
	return string(Brotli)
}
