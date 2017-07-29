package colorable

import (
	"bytes"
	"testing"
)

// checkEncoding checks that colorable is output encoding agnostic as long as
// the encoding is a superset of ASCII. This implies that one byte not part of
// an ANSI sequence must give exactly one byte in output
func checkEncoding(t *testing.T, data []byte) {
	// Send non-UTF8 data to colorable
	b := bytes.NewBuffer(make([]byte, 0, 10))
	if b.Len() != 0 {
		t.FailNow()
	}
	// TODO move colorable wrapping outside the test
	c := NewNonColorable(b)
	c.Write(data)
	if b.Len() != len(data) {
		t.Fatalf("%d bytes expected, got %d", len(data), b.Len())
	}
}

func TestEncoding(t *testing.T) {
	checkEncoding(t, []byte{})      // Empty
	checkEncoding(t, []byte(`abc`)) // "abc"
	checkEncoding(t, []byte(`é`))   // "é" in UTF-8
	checkEncoding(t, []byte{233})   // 'é' in Latin-1
}
