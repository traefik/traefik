package route

import (
	"fmt"
)

// charPos stores the position in the iterator
type charPos struct {
	i  int
	si int
}

// charIter is a iterator over sequence of strings, returns byte-by-byte characters in string by string
type charIter struct {
	i  int // position in the current string
	si int // position in the array of strings

	seq []string // sequence of strings, e.g. ["GET", "/path"]
	sep []byte   // every string in the sequence has an associated separator used for trie matching, e.g. path uses '/' for separator
	// so sequence ["a.host", "/path "]has acoompanying separators ['.', '/']
}

func newIter(seq []string, sep []byte) *charIter {
	return &charIter{
		i:   0,
		si:  0,
		seq: seq,
		sep: sep,
	}
}

func (r *charIter) level() int {
	return r.si
}

func (r *charIter) String() string {
	if r.isEnd() {
		return "<end>"
	}
	return fmt.Sprintf("<%d:%v>", r.i, r.seq[r.si])
}

func (r *charIter) isEnd() bool {
	return len(r.seq) == 0 || // no data at all
		(r.si >= len(r.seq)-1 && r.i >= len(r.seq[r.si])) || // we are at the last char of last seq
		(len(r.seq[r.si]) == 0) // empty input
}

func (r *charIter) position() charPos {
	return charPos{i: r.i, si: r.si}
}

func (r *charIter) setPosition(p charPos) {
	r.i = p.i
	r.si = p.si
}

func (r *charIter) pushBack() {
	if r.i == 0 && r.si == 0 { // this is start
		return
	} else if r.i == 0 && r.si != 0 { // this is start of the next string
		r.si--
		r.i = len(r.seq[r.si]) - 1
		return
	}
	r.i--
}

// next returns current byte in the sequence, separator corresponding to that byte, and boolean idicator of whether it's the end of the sequence
func (r *charIter) next() (byte, byte, bool) {
	// we have reached the last string in the index, end
	if r.isEnd() {
		return 0, 0, false
	}

	b := r.seq[r.si][r.i]
	sep := r.sep[r.si]
	r.i++

	// current string index exceeded the last char of the current string
	// move to the next string if it's present
	if r.i >= len(r.seq[r.si]) && r.si < len(r.seq)-1 {
		r.si++
		r.i = 0
	}

	return b, sep, true
}
