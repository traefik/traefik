// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package progress

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/snapcore/snapd/strutil/quantity"
)

var stdout io.Writer = os.Stdout

// ANSIMeter is a progress.Meter that uses ANSI escape codes to make
// better use of the available horizontal space.
type ANSIMeter struct {
	label   []rune
	total   float64
	written float64
	spin    int
	t0      time.Time
}

// these are the bits of the ANSI escapes (beyond \r) that we use
// (names of the terminfo capabilities, see terminfo(5))
var (
	// clear to end of line
	clrEOL = "\033[K"
	// make cursor invisible
	cursorInvisible = "\033[?25l"
	// make cursor visible
	cursorVisible = "\033[?25h"
	// turn on reverse video
	enterReverseMode = "\033[7m"
	// go back to normal video
	exitAttributeMode = "\033[0m"
)

var termWidth = func() int {
	col, _, _ := terminal.GetSize(0)
	if col <= 0 {
		// give up
		col = 80
	}
	return col
}

func (p *ANSIMeter) Start(label string, total float64) {
	p.label = []rune(label)
	p.total = total
	p.t0 = time.Now().UTC()
	fmt.Fprint(stdout, cursorInvisible)
}

func norm(col int, msg []rune) []rune {
	if col <= 0 {
		return []rune{}
	}
	out := make([]rune, col)
	copy(out, msg)
	d := col - len(msg)
	if d < 0 {
		out[col-1] = '…'
	} else {
		for i := len(msg); i < col; i++ {
			out[i] = ' '
		}
	}
	return out
}

func (p *ANSIMeter) SetTotal(total float64) {
	p.total = total
}

func (p *ANSIMeter) percent() string {
	if p.total == 0. {
		return "---%"
	}
	q := p.written * 100 / p.total
	if q > 999.4 || q < 0. {
		return "???%"
	}
	return fmt.Sprintf("%3.0f%%", q)
}

func (p *ANSIMeter) Set(current float64) {
	if current < 0 {
		current = 0
	}
	if current > p.total {
		current = p.total
	}

	p.written = current
	col := termWidth()
	// time left: 5
	//    gutter: 1
	//     speed: 8
	//    gutter: 1
	//   percent: 4
	//    gutter: 1
	//          =====
	//           20
	// and we want to leave at least 10 for the label, so:
	//  * if      width <= 15, don't show any of this (progress bar is good enough)
	//  * if 15 < width <= 20, only show time left (time left + gutter = 6)
	//  * if 20 < width <= 29, also show percentage (percent + gutter = 5
	//  * if 29 < width      , also show speed (speed+gutter = 9)
	var percent, speed, timeleft string
	if col > 15 {
		since := time.Now().UTC().Sub(p.t0).Seconds()
		per := since / p.written
		left := (p.total - p.written) * per
		timeleft = " " + quantity.FormatDuration(left)
		if col > 20 {
			percent = " " + p.percent()
			if col > 29 {
				speed = " " + quantity.FormatBPS(p.written, since, -1)
			}
		}
	}

	msg := make([]rune, 0, col)
	msg = append(msg, norm(col-len(percent)-len(speed)-len(timeleft), p.label)...)
	msg = append(msg, []rune(percent)...)
	msg = append(msg, []rune(speed)...)
	msg = append(msg, []rune(timeleft)...)
	i := int(current * float64(col) / p.total)
	fmt.Fprint(stdout, "\r", enterReverseMode, string(msg[:i]), exitAttributeMode, string(msg[i:]))
}

var spinner = []string{"/", "-", "\\", "|"}

func (p *ANSIMeter) Spin(msgstr string) {
	msg := []rune(msgstr)
	col := termWidth()
	if col-2 >= len(msg) {
		fmt.Fprint(stdout, "\r", string(norm(col-2, msg)), " ", spinner[p.spin])
		p.spin++
		if p.spin >= len(spinner) {
			p.spin = 0
		}
	} else {
		fmt.Fprint(stdout, "\r", string(norm(col, msg)))
	}
}

func (*ANSIMeter) Finished() {
	fmt.Fprint(stdout, "\r", exitAttributeMode, cursorVisible, clrEOL)
}

func (*ANSIMeter) Notify(msgstr string) {
	col := termWidth()
	fmt.Fprint(stdout, "\r", exitAttributeMode, clrEOL)

	msg := []rune(msgstr)
	var i int
	for len(msg) > col {
		for i = col; i >= 0; i-- {
			if unicode.IsSpace(msg[i]) {
				break
			}
		}
		if i < 1 {
			// didn't find anything; print the whole thing and try again
			fmt.Fprintln(stdout, string(msg[:col]))
			msg = msg[col:]
		} else {
			// found a space; print up to but not including it, and skip it
			fmt.Fprintln(stdout, string(msg[:i]))
			msg = msg[i+1:]
		}
	}
	fmt.Fprintln(stdout, string(msg))
}

func (p *ANSIMeter) Write(bs []byte) (n int, err error) {
	n = len(bs)
	p.Set(p.written + float64(n))

	return
}
