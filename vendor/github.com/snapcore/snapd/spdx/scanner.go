// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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

package spdx

import (
	"bufio"
	"io"
)

type Scanner struct {
	*bufio.Scanner
}

func spdxSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// skip WS
	start := 0
	for ; start < len(data); start++ {
		if data[start] != ' ' && data[start] != '\n' {
			break
		}
	}
	if start == len(data) {
		return start, nil, nil
	}

	switch data[start] {
	// found ( or )
	case '(', ')':
		return start + 1, data[start : start+1], nil
	}

	for i := start; i < len(data); i++ {
		switch data[i] {
		// token finished
		case ' ', '\n':
			return i + 1, data[start:i], nil
			// found ( or ) - we need to rescan it
		case '(', ')':
			return i, data[start:i], nil
		}
	}
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	return start, nil, nil
}

func NewScanner(r io.Reader) *Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(spdxSplit)
	return &Scanner{scanner}
}
