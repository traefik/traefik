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

package strutil

import (
	"bytes"
	"regexp"
)

// A MatchCounter is a discarding io.Writer that retains up to N
// matches to its Regexp before just counting matches.
//
// It does not work with regexps that cross newlines; in fact it will
// probably not work if the data written isn't line-orineted.
type MatchCounter struct {
	// Regexp to use to find matches in the stream
	Regexp *regexp.Regexp
	// Maximum number of matches to keep; if < 0, keep all matches
	N int

	count   int
	matches []string
	partial bytes.Buffer
}

func (w *MatchCounter) Write(p []byte) (int, error) {
	n := len(p)
	if w.partial.Len() > 0 {
		idx := bytes.IndexByte(p, '\n')
		if idx < 0 {
			// no newline yet, carry on accumulating
			w.partial.Write(p)
			return n, nil
		}
		idx++
		w.partial.Write(p[:idx])
		w.check(w.partial.Bytes())
		p = p[idx:]
	}
	w.partial.Reset()
	idx := bytes.LastIndexByte(p, '\n')
	if idx < 0 {
		w.partial.Write(p)
		return n, nil
	}
	idx++
	w.partial.Write(p[idx:])
	w.check(p[:idx])
	return n, nil
}

func (w *MatchCounter) check(p []byte) {
	matches := w.Regexp.FindAll(p, -1)
	for _, match := range matches {
		if w.N >= 0 && len(w.matches) >= w.N {
			break
		}
		w.matches = append(w.matches, string(match))
	}
	w.count += len(matches)
}

// Matches returns the first few matches, and the total number of matches seen.
func (w *MatchCounter) Matches() ([]string, int) {
	return w.matches, w.count
}
