package accesslog

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	defaultAsyncWriterSyncSize = 1024
)

type asyncWriter struct {
	originalFile *os.File
	stopCh       chan interface{}
	writerStream chan []byte
	mu           sync.Mutex

	sync.WaitGroup
}

func newAsyncWriter(chanSize int64, originalFile *os.File) *asyncWriter {
	cSize := chanSize
	if cSize == 0 {
		cSize = defaultAsyncWriterSyncSize
	}
	stopCh := make(chan interface{}, 0)
	w := make(chan []byte, cSize)
	aWriter := &asyncWriter{
		writerStream: w,
		originalFile: originalFile,
		stopCh:       stopCh,
	}

	aWriter.Add(1)
	go func() {
		defer aWriter.Done()
		for {
			select {
			case log := <-aWriter.writerStream:
				printLog(aWriter, aWriter.originalFile, log)
			case <-stopCh:
				return
			}
		}
	}()

	return aWriter

}

func (w *asyncWriter) Write(p []byte) (n int, err error) {
	size := len(p)
	b := make([]byte, size)
	copy(b, p)
	w.writerStream <- b
	return size, nil
}

func printLog(w *asyncWriter, writer io.Writer, log []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := writer.Write(log)
	if err != nil {
		logrus.Error(err)
	}
}

func (w *asyncWriter) drainChannel() {
	for log := range w.writerStream {
		printLog(w, w.originalFile, log)
	}
}

func (w *asyncWriter) Close() error {
	close(w.stopCh)
	w.Wait()
	// drain channel before close file
	close(w.writerStream)
	w.drainChannel()
	return w.originalFile.Close()
}
