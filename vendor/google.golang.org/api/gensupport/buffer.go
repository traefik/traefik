// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gensupport

import (
	"bytes"
	"io"

	"google.golang.org/api/googleapi"
)

// ResumableBuffer buffers data from an io.Reader to support uploading media in retryable chunks.
type ResumableBuffer struct {
	media io.Reader

	chunk []byte // The current chunk which is pending upload.  The capacity is the chunk size.
	err   error  // Any error generated when populating chunk by reading media.

	// The absolute position of chunk in the underlying media.
	off int64
}

func NewResumableBuffer(media io.Reader, chunkSize int) *ResumableBuffer {
	return &ResumableBuffer{media: media, chunk: make([]byte, 0, chunkSize)}
}

// Chunk returns the current buffered chunk, the offset in the underlying media
// from which the chunk is drawn, and the size of the chunk.
// Successive calls to Chunk return the same chunk between calls to Next.
func (rb *ResumableBuffer) Chunk() (chunk io.Reader, off int64, size int, err error) {
	// There may already be data in chunk if Next has not been called since the previous call to Chunk.
	if rb.err == nil && len(rb.chunk) == 0 {
		rb.err = rb.loadChunk()
	}
	return bytes.NewReader(rb.chunk), rb.off, len(rb.chunk), rb.err
}

// loadChunk will read from media into chunk, up to the capacity of chunk.
func (rb *ResumableBuffer) loadChunk() error {
	bufSize := cap(rb.chunk)
	rb.chunk = rb.chunk[:bufSize]

	read := 0
	var err error
	for err == nil && read < bufSize {
		var n int
		n, err = rb.media.Read(rb.chunk[read:])
		read += n
	}
	rb.chunk = rb.chunk[:read]
	return err
}

// Next advances to the next chunk, which will be returned by the next call to Chunk.
// Calls to Next without a corresponding prior call to Chunk will have no effect.
func (rb *ResumableBuffer) Next() {
	rb.off += int64(len(rb.chunk))
	rb.chunk = rb.chunk[0:0]
}

type readerTyper struct {
	io.Reader
	googleapi.ContentTyper
}

// ReaderAtToReader adapts a ReaderAt to be used as a Reader.
// If ra implements googleapi.ContentTyper, then the returned reader
// will also implement googleapi.ContentTyper, delegating to ra.
func ReaderAtToReader(ra io.ReaderAt, size int64) io.Reader {
	r := io.NewSectionReader(ra, 0, size)
	if typer, ok := ra.(googleapi.ContentTyper); ok {
		return readerTyper{r, typer}
	}
	return r
}
