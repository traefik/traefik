// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"io"
	"strings"
)

// A simple stack
type stack struct {
	data []interface{}
}

func (s *stack) empty() bool {
	return len(s.data) == 0
}

func (s *stack) push(value interface{}) {
	s.data = append(s.data, value)
}

func (s *stack) pop() interface{} {
	value := s.data[len(s.data)-1]
	s.data[len(s.data)-1] = nil
	s.data = s.data[:len(s.data)-1]
	return value
}

func (s *stack) peek() interface{} {
	return s.data[len(s.data)-1]
}

// A fifo is a simple first-in-first-out queue.
type fifo struct {
	data       []interface{}
	head, tail int
}

func (f *fifo) add(value interface{}) {
	if f.len()+1 >= len(f.data) {
		f.grow()
	}
	f.data[f.tail] = value
	if f.tail++; f.tail == len(f.data) {
		f.tail = 0
	}
}

func (f *fifo) remove() interface{} {
	value := f.data[f.head]
	f.data[f.head] = nil
	if f.head++; f.head == len(f.data) {
		f.head = 0
	}
	return value
}

func (f *fifo) len() int {
	if f.tail >= f.head {
		return f.tail - f.head
	}
	return len(f.data) - f.head + f.tail
}

func (f *fifo) grow() {
	c := len(f.data) * 2
	if c == 0 {
		c = 4
	}
	buf, count := make([]interface{}, c), f.len()
	if f.tail >= f.head {
		copy(buf[0:count], f.data[f.head:f.tail])
	} else {
		hindex := len(f.data) - f.head
		copy(buf[0:hindex], f.data[f.head:])
		copy(buf[hindex:count], f.data[:f.tail])
	}
	f.data, f.head, f.tail = buf, 0, count
}

// countReader implements a proxy reader that counts the number of
// bytes read from its encapsulated reader.
type countReader struct {
	r     io.Reader
	bytes int64
}

func newCountReader(r io.Reader) *countReader {
	return &countReader{r: r}
}

func (cr *countReader) Read(p []byte) (n int, err error) {
	b, err := cr.r.Read(p)
	cr.bytes += int64(b)
	return b, err
}

// countWriter implements a proxy writer that counts the number of
// bytes written by its encapsulated writer.
type countWriter struct {
	w     io.Writer
	bytes int64
}

func newCountWriter(w io.Writer) *countWriter {
	return &countWriter{w: w}
}

func (cw *countWriter) Write(p []byte) (n int, err error) {
	b, err := cw.w.Write(p)
	cw.bytes += int64(b)
	return b, err
}

// isWhitespace returns true if the byte slice contains only
// whitespace characters.
func isWhitespace(s string) bool {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

// spaceMatch returns true if namespace a is the empty string
// or if namespace a equals namespace b.
func spaceMatch(a, b string) bool {
	switch {
	case a == "":
		return true
	default:
		return a == b
	}
}

// spaceDecompose breaks a namespace:tag identifier at the ':'
// and returns the two parts.
func spaceDecompose(str string) (space, key string) {
	colon := strings.IndexByte(str, ':')
	if colon == -1 {
		return "", str
	}
	return str[:colon], str[colon+1:]
}

// Strings used by crIndent
const (
	crsp  = "\n                                                                "
	crtab = "\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
)

// crIndent returns a carriage return followed by n copies of the
// first non-CR character in the source string.
func crIndent(n int, source string) string {
	switch {
	case n < 0:
		return source[:1]
	case n < len(source):
		return source[:n+1]
	default:
		return source + strings.Repeat(source[1:2], n-len(source)+1)
	}
}

// nextIndex returns the index of the next occurrence of sep in s,
// starting from offset.  It returns -1 if the sep string is not found.
func nextIndex(s, sep string, offset int) int {
	switch i := strings.Index(s[offset:], sep); i {
	case -1:
		return -1
	default:
		return offset + i
	}
}

// isInteger returns true if the string s contains an integer.
func isInteger(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && !(i == 0 && s[i] == '-') {
			return false
		}
	}
	return true
}
