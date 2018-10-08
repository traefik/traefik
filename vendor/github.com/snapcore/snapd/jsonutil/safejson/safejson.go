// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package safejson

import (
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/snapcore/snapd/strutil"
)

// String accepts any valid JSON string. Its Clean method will remove
// characters that aren't expected in a short descriptive text.
// I.e.: Cc, Co, Cf, Cs, noncharacters, and � (U+FFFD, the replacement
// character) are removed.
type String struct {
	s string
}

func (str *String) UnmarshalJSON(in []byte) (err error) {
	str.s, err = unmarshal(in, uOpt{})
	return
}

// Clean returns the string, with Cc, Co, Cf, Cs, noncharacters,
// and � (U+FFFD) removed.
func (str String) Clean() string {
	return str.s
}

// Paragraph accepts any valid JSON string. Its Clean method will remove
// characters that aren't expected in a long descriptive text.
// I.e.: Cc (except for \n), Co, Cf, Cs, noncharacters, and � (U+FFFD,
// the replacement character) are removed.
type Paragraph struct {
	s string
}

func (par *Paragraph) UnmarshalJSON(in []byte) (err error) {
	par.s, err = unmarshal(in, uOpt{nlOK: true})
	return
}

// Clean returns the string, with Cc minus \n, Co, Cf, Cs, noncharacters,
// and � (U+FFFD) removed.
func (par Paragraph) Clean() string {
	return par.s
}

func unescapeUCS2(in []byte) (rune, bool) {
	if len(in) < 6 || in[0] != '\\' || in[1] != 'u' {
		return -1, false
	}
	u, err := strconv.ParseUint(string(in[2:6]), 16, 32)
	if err != nil {
		return -1, false
	}
	return rune(u), true
}

type uOpt struct {
	nlOK   bool
	simple bool
}

func unmarshal(in []byte, o uOpt) (string, error) {
	// heavily based on (inspired by?) unquoteBytes from encoding/json

	if len(in) < 2 || in[0] != '"' || in[len(in)-1] != '"' {
		// maybe it's a null and that's alright
		if len(in) == 4 && in[0] == 'n' && in[1] == 'u' && in[2] == 'l' && in[3] == 'l' {
			return "", nil
		}
		return "", fmt.Errorf("missing string delimiters: %q", in)
	}

	// prune the quotes
	in = in[1 : len(in)-1]
	i := 0
	// try the fast track
	for i < len(in) {
		// 0x00..0x19 is the first of Cc
		// 0x20..0x7e is all of printable ASCII (minus control chars)
		if in[i] < 0x20 || in[i] > 0x7e || in[i] == '\\' || in[i] == '"' {
			break
		}
		i++
	}
	if i == len(in) {
		// wee
		return string(in), nil
	}
	if o.simple {
		return "", fmt.Errorf("character %q in string %q unsupported for this value", in[i], in)
	}
	// in[i] is the first problematic one
	out := make([]byte, i, len(in)+2*utf8.UTFMax)
	copy(out, in)
	var r, r2 rune
	var n int
	var c byte
	var ubuf [utf8.UTFMax]byte
	var ok bool
	for i < len(in) {
		c = in[i]
		switch {
		case c == '"':
			return "", fmt.Errorf("unexpected unescaped quote at %d in \"%s\"", i, in)
		case c < 0x20:
			return "", fmt.Errorf("unexpected control character at %d in %q", i, in)
		case c == '\\':
			// handle escapes
			i++
			if i == len(in) {
				return "", fmt.Errorf("unexpected end of string (trailing backslash) in \"%s\"", in)
			}
			switch in[i] {
			case 'u':
				// oh dear, a unicode wotsit
				r, ok = unescapeUCS2(in[i-1:])
				if !ok {
					x := in[i-1:]
					if len(x) > 6 {
						x = x[:6]
					}
					return "", fmt.Errorf(`badly formed \u escape %q at %d of "%s"`, x, i, in)
				}
				i += 5
				if utf16.IsSurrogate(r) {
					// sigh
					r2, ok = unescapeUCS2(in[i:])
					if !ok {
						x := in[i:]
						if len(x) > 6 {
							x = x[:6]
						}
						return "", fmt.Errorf(`badly formed \u escape %q at %d of "%s"`, x, i, in)
					}
					i += 6
					r = utf16.DecodeRune(r, r2)
				}
				if r <= 0x9f {
					// otherwise, it's Cc (both halves, as we're looking at runes)
					if (o.nlOK && r == '\n') || (r >= 0x20 && r <= 0x7e) {
						out = append(out, byte(r))
					}
				} else if r != unicode.ReplacementChar && !unicode.Is(strutil.Ctrl, r) {
					n = utf8.EncodeRune(ubuf[:], r)
					out = append(out, ubuf[:n]...)
				}
			case 'b', 'f', 'r', 't':
				// do nothing
				i++
			case 'n':
				if o.nlOK {
					out = append(out, '\n')
				}
				i++
			case '"', '/', '\\':
				// the spec says just ", / and \ can be backslash-escaped
				// but go adds ' to the list (in unquoteBytes)
				out = append(out, in[i])
				i++
			default:
				return "", fmt.Errorf(`unknown escape '%c' at %d of "%s"`, in[i], i, in)
			}
		case c <= 0x7e:
			// printable ASCII, except " or \
			out = append(out, c)
			i++
		default:
			r, n = utf8.DecodeRune(in[i:])
			j := i + n
			if r > 0x9f && r != unicode.ReplacementChar && !unicode.Is(strutil.Ctrl, r) {
				out = append(out, in[i:j]...)
			}
			i = j
		}
	}

	return string(out), nil
}
