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

type LimitedBuffer struct {
	buffer   []byte
	maxLines int
	maxBytes int
}

func NewLimitedBuffer(maxLines, maxBytes int) *LimitedBuffer {
	return &LimitedBuffer{
		maxLines: maxLines,
		maxBytes: maxBytes,
	}
}

func (lb *LimitedBuffer) Write(data []byte) (int, error) {
	drop := len(lb.buffer) + len(data) - lb.maxBytes
	switch {
	case drop < 0:
		lb.buffer = append(lb.buffer, data...)
	case drop > len(lb.buffer):
		lb.buffer = append(lb.buffer[:0], data[drop-len(lb.buffer):]...)
	default:
		keep := copy(lb.buffer, lb.buffer[drop:])
		lb.buffer = append(lb.buffer[:keep], data...)
	}
	return len(data), nil
}

func (lb *LimitedBuffer) Bytes() []byte {
	return TruncateOutput(lb.buffer, lb.maxLines, lb.maxBytes)
}
