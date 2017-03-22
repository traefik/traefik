package streams

import (
	"github.com/containous/traefik/middlewares/audittap/audittypes"
	"testing"

	"bufio"
	"github.com/stretchr/testify/assert"
	"os"
)

const tmpFile = "/tmp/testFileSink"

var encodedJSONSample = audittypes.Encoded{
	Bytes: []byte("[1,2,3]"),
	Err:   nil,
}

func TestFileSink(t *testing.T) {
	w, err := NewFileSink(tmpFile, "foo", noopRenderer)
	assert.NoError(t, err)

	defer func() {
		e := os.Remove(tmpFile + "-foo.json")
		assert.NoError(t, e)
	}()

	err = w.Audit(encodedJSONSample)
	assert.NoError(t, err)

	err = w.Audit(encodedJSONSample)
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	f, err := os.Open(tmpFile + "-foo.json")
	assert.NoError(t, err)

	scanner := bufio.NewScanner(f) // default behaviour splits on line boundaries

	// line 1
	assert.True(t, scanner.Scan())
	assert.Equal(t, "[", scanner.Text())

	// line 2
	assert.True(t, scanner.Scan())
	line := scanner.Text()
	assert.Equal(t, `[1,2,3],`, line)

	// line 3
	assert.True(t, scanner.Scan())
	line = scanner.Text()
	assert.Equal(t, `[1,2,3]`, line)

	// line 4
	assert.True(t, scanner.Scan())
	assert.Equal(t, "]", scanner.Text())

	// end of file
	assert.False(t, scanner.Scan())
}
