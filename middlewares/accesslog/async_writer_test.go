package accesslog

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteAsync(t *testing.T) {
	r, w, _ := os.Pipe()
	aLog := newAsyncWriter(1024, w)
	defer aLog.Close()
	defer r.Close()
	length, _ := aLog.Write([]byte("hello"))
	assert.Equal(t, length, 5)
}

func TestDrainWriter(t *testing.T) {
	r, w, _ := os.Pipe()
	aLog := newAsyncWriter(1024, w)

	close(aLog.stopCh)
	aLog.Wait()

	expected := 10
	for i := 0; i < expected; i++ {
		_, _ = aLog.Write([]byte(strconv.Itoa(i) + "|"))
	}

	close(aLog.writerStream)
	aLog.drainChannel()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	defer r.Close()

	assert.Equal(t, strings.Count(buf.String(), "|"), expected)
}
