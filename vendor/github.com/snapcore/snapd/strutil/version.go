// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2017 Canonical Ltd
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
	"regexp"
	"strconv"
	"strings"
)

const (
	reDigit           = "[0-9]"
	reAlpha           = "[a-zA-Z]"
	reDigitOrNonDigit = "[0-9]+|[^0-9]+"

	reHasEpoch = "^[0-9]+:"
)

var (
	matchDigit = regexp.MustCompile(reDigit).Match
	matchAlpha = regexp.MustCompile(reAlpha).Match
	findFrags  = regexp.MustCompile(reDigitOrNonDigit).FindAllString
	matchEpoch = regexp.MustCompile(reHasEpoch).MatchString
)

// golang: seriously? that's sad!
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// version number compare, inspired by the libapt/python-debian code
func cmpInt(intA, intB int) int {
	if intA < intB {
		return -1
	} else if intA > intB {
		return 1
	}
	return 0
}

func chOrder(ch uint8) int {
	// "~" is lower than everything else
	if ch == '~' {
		return -10
	}
	// empty is higher than "~" but lower than everything else
	if ch == 0 {
		return -5
	}
	if matchAlpha([]byte{ch}) {
		return int(ch)
	}

	// can only happen if cmpString sets '0' because there is no fragment
	if matchDigit([]byte{ch}) {
		return 0
	}

	return int(ch) + 256
}

func cmpString(as, bs string) int {
	for i := 0; i < max(len(as), len(bs)); i++ {
		var a uint8
		var b uint8
		if i < len(as) {
			a = as[i]
		}
		if i < len(bs) {
			b = bs[i]
		}
		if chOrder(a) < chOrder(b) {
			return -1
		}
		if chOrder(a) > chOrder(b) {
			return +1
		}
	}
	return 0
}

func cmpFragment(a, b string) int {
	intA, errA := strconv.Atoi(a)
	intB, errB := strconv.Atoi(b)
	if errA == nil && errB == nil {
		return cmpInt(intA, intB)
	}
	res := cmpString(a, b)
	//fmt.Println(a, b, res)
	return res
}

func getFragments(a string) []string {
	return findFrags(a, -1)
}

// VersionIsValid returns true if the given string is a valid
// version number according to the debian policy
func VersionIsValid(a string) bool {
	if matchEpoch(a) {
		return false
	}
	if strings.Count(a, "-") > 1 {
		return false
	}
	return true
}

func compareSubversion(va, vb string) int {
	fragsA := getFragments(va)
	fragsB := getFragments(vb)

	for i := 0; i < max(len(fragsA), len(fragsB)); i++ {
		a := ""
		b := ""
		if i < len(fragsA) {
			a = fragsA[i]
		}
		if i < len(fragsB) {
			b = fragsB[i]
		}
		res := cmpFragment(a, b)
		//fmt.Println(a, b, res)
		if res != 0 {
			return res
		}
	}
	return 0
}

// VersionCompare compare two version strings that follow the debian
// version policy and
// Returns:
//   -1 if a is smaller than b
//    0 if a equals b
//   +1 if a is bigger than b
func VersionCompare(va, vb string) (res int, err error) {
	// FIXME: return err here instead
	if !VersionIsValid(va) {
		return 0, fmt.Errorf("invalid version %q", va)
	}
	if !VersionIsValid(vb) {
		return 0, fmt.Errorf("invalid version %q", vb)
	}

	if !strings.Contains(va, "-") {
		va += "-0"
	}
	if !strings.Contains(vb, "-") {
		vb += "-0"
	}

	// the main version number (before the "-")
	mainA := strings.Split(va, "-")[0]
	mainB := strings.Split(vb, "-")[0]
	res = compareSubversion(mainA, mainB)
	if res != 0 {
		return res, nil
	}

	// the subversion revision behind the "-"
	revA := strings.Split(va, "-")[1]
	revB := strings.Split(vb, "-")[1]
	return compareSubversion(revA, revB), nil
}
