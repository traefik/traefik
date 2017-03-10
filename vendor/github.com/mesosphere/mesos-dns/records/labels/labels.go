package labels

import (
	"bytes"
	"strings"
)

// Sep is the default domain fragment separator.
const Sep = "."

// DomainFrag mangles the given name in order to produce a valid domain fragment.
// A valid domain fragment will consist of one or more host name labels
// concatenated by the given separator.
func DomainFrag(name, sep string, label Func) string {
	var labels []string
	for _, part := range strings.Split(name, sep) {
		if lab := label(part); lab != "" {
			labels = append(labels, lab)
		}
	}
	return strings.Join(labels, sep)
}

// Func is a function type representing label functions.
type Func func(string) string

// RFC952 mangles a name to conform to the DNS label rules specified in RFC952.
// See http://www.rfc-base.org/txt/rfc-952.txt
func RFC952(name string) string {
	return string(label([]byte(name), 24, "-0123456789", "-"))
}

// RFC1123 mangles a name to conform to the DNS label rules specified in RFC1123.
// See http://www.rfc-base.org/txt/rfc-1123.txt
func RFC1123(name string) string {
	return string(label([]byte(name), 63, "-", "-"))
}

// label computes a label from the given name with maxlen length and the
// left and right cutsets trimmed from their respective ends.
func label(name []byte, maxlen int, left, right string) []byte {
	return trimCut(bytes.Map(mapping, name), maxlen, left, right)
}

// mapping maps a given rune to its valid DNS label counterpart.
func mapping(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return r - ('A' - 'a')
	case r >= 'a' && r <= 'z':
		fallthrough
	case r >= '0' && r <= '9':
		return r
	case r == '-' || r == '.' || r == '_':
		return '-'
	default:
		return -1
	}
}

// trimCut cuts the given label at min(maxlen, len(label)) and ensures the left
// and right cutsets are trimmed from their respective ends.
func trimCut(label []byte, maxlen int, left, right string) []byte {
	trim := bytes.TrimLeft(label, left)
	size := min(len(trim), maxlen)
	head := bytes.TrimRight(trim[:size], right)
	if len(head) == size {
		return head
	}
	tail := bytes.TrimLeft(trim[size:], right)
	if len(tail) > 0 {
		return append(head, tail[:size-len(head)]...)
	}
	return head
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
