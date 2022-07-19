package capture

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestReader(t *testing.T) {
	buff := bytes.NewBuffer([]byte("foo"))
	rr := newRequestReader(io.NopCloser(buff))
	assert.Equal(t, int64(0), rr.Size())

	n, err := rr.Read([]byte("bar"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	err = rr.Close()
	require.NoError(t, err)
	assert.Equal(t, int64(3), rr.Size())
}
