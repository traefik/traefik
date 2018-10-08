// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func init() {
	// golang does not init Seed() itself
	rand.Seed(time.Now().UTC().UnixNano())
}

const letters = "BCDFGHJKLMNPQRSTVWXYbcdfghjklmnpqrstvwxy0123456789"

// MakeRandomString returns a random string of length length
//
// The vowels are omitted to avoid that words are created by pure
// chance. Numbers are included.
func MakeRandomString(length int) string {
	out := ""
	for i := 0; i < length; i++ {
		out += string(letters[rand.Intn(len(letters))])
	}

	return out
}

// Convert the given size in btes to a readable string
func SizeToStr(size int64) string {
	suffixes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	for _, suf := range suffixes {
		if size < 1000 {
			return fmt.Sprintf("%d%s", size, suf)
		}
		size /= 1000
	}
	panic("SizeToStr got a size bigger than math.MaxInt64")
}

// Quoted formats a slice of strings to a quoted list of
// comma-separated strings, e.g. `"snap1", "snap2"`
func Quoted(names []string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = strconv.Quote(name)
	}

	return strings.Join(quoted, ", ")
}

// ListContains determines whether the given string is contained in the
// given list of strings.
func ListContains(list []string, str string) bool {
	for _, k := range list {
		if k == str {
			return true
		}
	}
	return false
}

// SortedListContains determines whether the given string is contained
// in the given list of strings, which must be sorted.
func SortedListContains(list []string, str string) bool {
	i := sort.SearchStrings(list, str)
	if i >= len(list) {
		return false
	}
	return list[i] == str
}

// TruncateOutput truncates input data by maxLines, imposing maxBytes limit (total) for them.
// The maxLines may be 0 to avoid the constraint on number of lines.
func TruncateOutput(data []byte, maxLines, maxBytes int) []byte {
	if maxBytes > len(data) {
		maxBytes = len(data)
	}
	lines := maxLines
	bytes := maxBytes
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == '\n' {
			lines--
		}
		if lines == 0 || bytes == 0 {
			return data[i+1:]
		}
		bytes--
	}
	return data
}

// splitUnit takes a string of the form "123unit" and splits
// it into the number and non-number parts (123,"unit").
func splitUnit(inp string) (number int64, unit string, err error) {
	// go after the number first, break on first non-digit
	var nonDigit int
	for i, c := range inp {
		if !unicode.IsDigit(c) {
			number, err = strconv.ParseInt(inp[0:i], 10, 64)
			if err != nil {
				return 0, "", err
			}
			nonDigit = i
			break
		}
	}
	if nonDigit == 0 {
		return 0, "", fmt.Errorf("need a number with a unit as input")
	}

	return number, inp[nonDigit:], nil
}

// ParseByteSize parses a value like 500kB and returns the number
// in bytes. The case of the unit will be ignored for user convenience.
func ParseByteSize(inp string) (int64, error) {
	unitMultiplier := map[string]int64{
		"B": 1,
		// strictly speaking this is "kB" but we ignore cases
		"KB": 1000,
		"MB": 1000 * 1000,
		"GB": 1000 * 1000 * 1000,
		"TB": 1000 * 1000 * 1000 * 1000,
		"PB": 1000 * 1000 * 1000 * 1000 * 1000,
		"EB": 1000 * 1000 * 1000 * 1000 * 1000 * 1000,
	}

	errPrefix := fmt.Sprintf("cannot parse %q: ", inp)

	val, unit, err := splitUnit(inp)
	if err != nil {
		return 0, fmt.Errorf(errPrefix+"%s", err)
	}
	if unit == "" {
		return 0, fmt.Errorf(errPrefix + "need a number with a unit as input")
	}

	mul, ok := unitMultiplier[strings.ToUpper(unit)]
	if !ok {
		return 0, fmt.Errorf(errPrefix + "try 'kB' or 'MB'")
	}

	return val * mul, nil
}
