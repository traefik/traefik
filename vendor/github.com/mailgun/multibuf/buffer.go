// package multibuf implements buffer optimized for streaming large chunks of data,
// multiple reads and  optional partial buffering to disk.
package multibuf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// MultiReader provides Read, Close, Seek and Size methods. In addition to that it supports WriterTo interface
// to provide efficient writing schemes, as functions like io.Copy use WriterTo when it's available.
type MultiReader interface {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo

	// Size calculates and returns the total size of the reader and not the length remaining.
	Size() (int64, error)
}

// WriterOnce implements write once, read many times writer. Create a WriterOnce and write to it, once Reader() function has been
// called, the internal data is transferred to MultiReader and this instance of WriterOnce should be no longer used.
type WriterOnce interface {
	// Write implements io.Writer
	Write(p []byte) (int, error)
	// Reader transfers all data written to this writer to MultiReader. If there was no data written it retuns an error
	Reader() (MultiReader, error)
	// WriterOnce owns the data before Reader has been called, so Close will close all the underlying files if Reader has not been called.
	Close() error
}

// MaxBytes, ignored if set to value >=, if request exceeds the specified limit, the reader will return error,
// by default buffer is not limited, negative values mean no limit
func MaxBytes(m int64) optionSetter {
	return func(o *options) error {
		o.maxBytes = m
		return nil
	}
}

// MemBytes specifies the largest buffer to hold in RAM before writing to disk, default is 1MB
func MemBytes(m int64) optionSetter {
	return func(o *options) error {
		if m < 0 {
			return fmt.Errorf("MemBytes should be >= 0")
		}
		o.memBytes = m
		return nil
	}
}

// NewWriterOnce returns io.ReadWrite compatible object that can limit the size of the buffer and persist large buffers to disk.
// WriterOnce implements write once, read many times writer. Create a WriterOnce and write to it, once Reader() function has been
// called, the internal data is transferred to MultiReader and this instance of WriterOnce should be no longer used.
// By default NewWriterOnce returns unbound buffer that will allow to write up to 1MB in RAM and will start buffering to disk
// It supports multiple functional optional arguments:
//
//    // Buffer up to 1MB in RAM and limit max buffer size to 20MB
//    multibuf.NewWriterOnce(r, multibuf.MemBytes(1024 * 1024), multibuf.MaxBytes(1024 * 1024 * 20))
//
//
func NewWriterOnce(setters ...optionSetter) (WriterOnce, error) {
	o := options{
		memBytes: DefaultMemBytes,
		maxBytes: DefaultMaxBytes,
	}
	if o.memBytes == 0 {
		o.memBytes = DefaultMemBytes
	}
	for _, s := range setters {
		if err := s(&o); err != nil {
			return nil, err
		}
	}
	return &writerOnce{o: o}, nil
}

// New returns MultiReader that can limit the size of the buffer and persist large buffers to disk.
// By default New returns unbound buffer that will read up to 1MB in RAM and will start buffering to disk
// It supports multiple functional optional arguments:
//
//    // Buffer up to 1MB in RAM and limit max buffer size to 20MB
//    multibuf.New(r, multibuf.MemBytes(1024 * 1024), multibuf.MaxBytes(1024 * 1024 * 20))
//
//
func New(input io.Reader, setters ...optionSetter) (MultiReader, error) {
	o := options{
		memBytes: DefaultMemBytes,
		maxBytes: DefaultMaxBytes,
	}

	for _, s := range setters {
		if err := s(&o); err != nil {
			return nil, err
		}
	}
	if o.memBytes == 0 {
		o.memBytes = DefaultMemBytes
	}
	if o.maxBytes > 0 && o.maxBytes < o.memBytes {
		o.memBytes = o.maxBytes
	}

	memReader := &io.LimitedReader{
		R: input,      // Read from this reader
		N: o.memBytes, // Maximum amount of data to read
	}
	readers := make([]io.ReadSeeker, 0, 2)

	buffer, err := ioutil.ReadAll(memReader)
	if err != nil {
		return nil, err
	}
	readers = append(readers, bytes.NewReader(buffer))

	var file *os.File
	// This means that we have exceeded all the memory capacity and we will start buffering the body to disk.
	totalBytes := int64(len(buffer))
	if memReader.N <= 0 {
		file, err = ioutil.TempFile("", tempFilePrefix)
		if err != nil {
			return nil, err
		}
		os.Remove(file.Name())

		readSrc := input
		if o.maxBytes > 0 {
			readSrc = &maxReader{R: input, Max: o.maxBytes - o.memBytes}
		}

		writtenBytes, err := io.Copy(file, readSrc)
		if err != nil {
			return nil, err
		}
		totalBytes += writtenBytes
		file.Seek(0, 0)
		readers = append(readers, file)
	}

	var cleanupFn cleanupFunc
	if file != nil {
		cleanupFn = func() error {
			file.Close()
			return nil
		}
	}
	return newBuf(totalBytes, cleanupFn, readers...), nil
}

// MaxSizeReachedError is returned when the maximum allowed buffer size is reached when reading
type MaxSizeReachedError struct {
	MaxSize int64
}

func (e *MaxSizeReachedError) Error() string {
	return fmt.Sprintf("Maximum size %d was reached", e)
}

const (
	DefaultMemBytes = 1048576
	DefaultMaxBytes = -1
	// Equivalent of bytes.MinRead used in ioutil.ReadAll
	DefaultBufferBytes = 512
)

// Constraints:
//  - Implements io.Reader
//  - Implements Seek(0, 0)
//	- Designed for Write once, Read many times.
type multiReaderSeek struct {
	length  int64
	readers []io.ReadSeeker
	mr      io.Reader
	cleanup cleanupFunc
}

type cleanupFunc func() error

func newBuf(length int64, cleanup cleanupFunc, readers ...io.ReadSeeker) *multiReaderSeek {
	converted := make([]io.Reader, len(readers))
	for i, r := range readers {
		// This conversion is safe as ReadSeeker includes Reader
		converted[i] = r.(io.Reader)
	}

	return &multiReaderSeek{
		length:  length,
		readers: readers,
		mr:      io.MultiReader(converted...),
		cleanup: cleanup,
	}
}

func (mr *multiReaderSeek) Close() (err error) {
	if mr.cleanup != nil {
		return mr.cleanup()
	}
	return nil
}

func (mr *multiReaderSeek) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, DefaultBufferBytes)
	var total int64
	for {
		n, err := mr.mr.Read(b)
		// Recommended way is to always handle non 0 reads despite the errors
		if n > 0 {
			nw, errw := w.Write(b[:n])
			total += int64(nw)
			// Write must return a non-nil error if it returns nw < n
			if nw != n || errw != nil {
				return total, errw
			}
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, err
		}
	}
}

func (mr *multiReaderSeek) Read(p []byte) (n int, err error) {
	return mr.mr.Read(p)
}

func (mr *multiReaderSeek) Size() (int64, error) {
	return mr.length, nil
}

func (mr *multiReaderSeek) Seek(offset int64, whence int) (int64, error) {
	// TODO: implement other whence
	// TODO: implement real offsets

	if whence != 0 {
		return 0, fmt.Errorf("multiReaderSeek: unsupported whence")
	}

	if offset != 0 {
		return 0, fmt.Errorf("multiReaderSeek: unsupported offset")
	}

	for _, seeker := range mr.readers {
		seeker.Seek(0, 0)
	}

	ior := make([]io.Reader, len(mr.readers))
	for i, arg := range mr.readers {
		ior[i] = arg.(io.Reader)
	}
	mr.mr = io.MultiReader(ior...)

	return 0, nil
}

type options struct {
	// MemBufferBytes sets up the size of the memory buffer for this request.
	// If the data size exceeds the limit, the remaining request part will be saved on the file system.
	memBytes int64

	maxBytes int64
}

type optionSetter func(o *options) error

// MaxReader does not allow to read more than Max bytes and returns error if this limit has been exceeded.
type maxReader struct {
	R   io.Reader // underlying reader
	N   int64     // bytes read
	Max int64     // max bytes to read
}

func (r *maxReader) Read(p []byte) (int, error) {
	readBytes, err := r.R.Read(p)
	if err != nil && err != io.EOF {
		return readBytes, err
	}

	r.N += int64(readBytes)
	if r.N > r.Max {
		return readBytes, &MaxSizeReachedError{MaxSize: r.Max}
	}
	return readBytes, err
}

const (
	writerInit = iota
	writerMem
	writerFile
	writerCalledRead
	writerErr
)

type writerOnce struct {
	o         options
	err       error
	state     int
	mem       *bytes.Buffer
	file      *os.File
	total     int64
	cleanupFn cleanupFunc
}

// how many bytes we can still write to memory
func (w *writerOnce) writeToMem(p []byte) int {
	left := w.o.memBytes - w.total
	if left <= 0 {
		return 0
	}
	bufLen := len(p)
	if int64(bufLen) < left {
		return bufLen
	}
	return int(left)
}

func (w *writerOnce) Write(p []byte) (int, error) {
	out, err := w.write(p)
	return out, err
}

func (w *writerOnce) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *writerOnce) write(p []byte) (int, error) {
	if w.o.maxBytes > 0 && int64(len(p))+w.total > w.o.maxBytes {
		return 0, fmt.Errorf("total size of %d exceeded allowed %d", int64(len(p))+w.total, w.o.maxBytes)
	}
	switch w.state {
	case writerCalledRead:
		return 0, fmt.Errorf("can not write after reader has been called")
	case writerInit:
		w.mem = &bytes.Buffer{}
		w.state = writerMem
		fallthrough
	case writerMem:
		writeToMem := w.writeToMem(p)
		if writeToMem > 0 {
			wrote, err := w.mem.Write(p[:writeToMem])
			w.total += int64(wrote)
			if err != nil {
				return wrote, err
			}
		}
		left := len(p) - writeToMem
		if left <= 0 {
			return len(p), nil
		}
		// we can't write to memory any more, switch to file
		if err := w.initFile(); err != nil {
			return int(writeToMem), err
		}
		w.state = writerFile
		wrote, err := w.file.Write(p[writeToMem:])
		w.total += int64(wrote)
		return len(p), err
	case writerFile:
		wrote, err := w.file.Write(p)
		w.total += int64(wrote)
		return wrote, err
	}
	return 0, fmt.Errorf("unsupported state: %d", w.state)
}

func (w *writerOnce) initFile() error {
	file, err := ioutil.TempFile("", tempFilePrefix)
	if err != nil {
		return err
	}
	w.file = file
	w.cleanupFn = func() error {
		file.Close()
		os.Remove(file.Name())
		return nil
	}
	return nil
}

func (w *writerOnce) Reader() (MultiReader, error) {
	switch w.state {
	case writerInit:
		return nil, fmt.Errorf("no data ready")
	case writerCalledRead:
		return nil, fmt.Errorf("reader has been called")
	case writerMem:
		w.state = writerCalledRead
		return newBuf(w.total, nil, bytes.NewReader(w.mem.Bytes())), nil
	case writerFile:
		_, err := w.file.Seek(0, 0)
		if err != nil {
			return nil, err
		}
		// we are not responsible for file and buffer any more
		w.state = writerCalledRead
		br, fr := bytes.NewReader(w.mem.Bytes()), w.file
		w.file = nil
		w.mem = nil
		return newBuf(w.total, w.cleanupFn, br, fr), nil
	}
	return nil, fmt.Errorf("unsupported state: %d\n", w.state)
}

const tempFilePrefix = "temp-multibuf-"
