// Copyright 2015 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package xstrings

import (
	"bytes"
	"math/rand"
	"unicode"
	"unicode/utf8"
)

// ToCamelCase can convert all lower case characters behind underscores
// to upper case character.
// Underscore character will be removed in result except following cases.
//     * More than 1 underscore.
//           "a__b" => "A_B"
//     * At the beginning of string.
//           "_a" => "_A"
//     * At the end of string.
//           "ab_" => "Ab_"
func ToCamelCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}
	var r0, r1 rune
	var size int

	// leading '_' will appear in output.
	for len(str) > 0 {
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if r0 != '_' {
			break
		}

		buf.WriteRune(r0)
	}

	if len(str) == 0 {
		return buf.String()
	}

	buf.WriteRune(unicode.ToUpper(r0))
	r0, size = utf8.DecodeRuneInString(str)
	str = str[size:]

	for len(str) > 0 {
		r1 = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		if r1 == '_' && r0 != '_' {
			r0 = unicode.ToUpper(r0)
		} else {
			buf.WriteRune(r1)
		}
	}

	buf.WriteRune(r0)
	return buf.String()
}

// ToSnakeCase can convert all upper case characters in a string to
// underscore format.
//
// Some samples.
//     "FirstName"  => "first_name"
//     "HTTPServer" => "http_server"
//     "NoHTTPS"    => "no_https"
//     "GO_PATH"    => "go_path"
//     "GO PATH"    => "go_path"      // space is converted to underscore.
//     "GO-PATH"    => "go_path"      // hyphen is converted to underscore.
func ToSnakeCase(str string) string {
	if len(str) == 0 {
		return ""
	}

	buf := &bytes.Buffer{}
	var prev, r0, r1 rune
	var size int

	r0 = '_'

	for len(str) > 0 {
		prev = r0
		r0, size = utf8.DecodeRuneInString(str)
		str = str[size:]

		switch {
		case r0 == utf8.RuneError:
			buf.WriteByte(byte(str[0]))

		case unicode.IsUpper(r0):
			if prev != '_' {
				buf.WriteRune('_')
			}

			buf.WriteRune(unicode.ToLower(r0))

			if len(str) == 0 {
				break
			}

			r0, size = utf8.DecodeRuneInString(str)
			str = str[size:]

			if !unicode.IsUpper(r0) {
				buf.WriteRune(r0)
				break
			}

			// find next non-upper-case character and insert `_` properly.
			// it's designed to convert `HTTPServer` to `http_server`.
			// if there are more than 2 adjacent upper case characters in a word,
			// treat them as an abbreviation plus a normal word.
			for len(str) > 0 {
				r1 = r0
				r0, size = utf8.DecodeRuneInString(str)
				str = str[size:]

				if r0 == utf8.RuneError {
					buf.WriteRune(unicode.ToLower(r1))
					buf.WriteByte(byte(str[0]))
					break
				}

				if !unicode.IsUpper(r0) {
					if r0 == '_' || r0 == ' ' || r0 == '-' {
						r0 = '_'

						buf.WriteRune(unicode.ToLower(r1))
					} else {
						buf.WriteRune('_')
						buf.WriteRune(unicode.ToLower(r1))
						buf.WriteRune(r0)
					}

					break
				}

				buf.WriteRune(unicode.ToLower(r1))
			}

			if len(str) == 0 || r0 == '_' {
				buf.WriteRune(unicode.ToLower(r0))
				break
			}

		default:
			if r0 == ' ' || r0 == '-' {
				r0 = '_'
			}

			buf.WriteRune(r0)
		}
	}

	return buf.String()
}

// SwapCase will swap characters case from upper to lower or lower to upper.
func SwapCase(str string) string {
	var r rune
	var size int

	buf := &bytes.Buffer{}

	for len(str) > 0 {
		r, size = utf8.DecodeRuneInString(str)

		switch {
		case unicode.IsUpper(r):
			buf.WriteRune(unicode.ToLower(r))

		case unicode.IsLower(r):
			buf.WriteRune(unicode.ToUpper(r))

		default:
			buf.WriteRune(r)
		}

		str = str[size:]
	}

	return buf.String()
}

// FirstRuneToUpper converts first rune to upper case if necessary.
func FirstRuneToUpper(str string) string {
	if str == "" {
		return str
	}

	r, size := utf8.DecodeRuneInString(str)

	if !unicode.IsLower(r) {
		return str
	}

	buf := &bytes.Buffer{}
	buf.WriteRune(unicode.ToUpper(r))
	buf.WriteString(str[size:])
	return buf.String()
}

// FirstRuneToLower converts first rune to lower case if necessary.
func FirstRuneToLower(str string) string {
	if str == "" {
		return str
	}

	r, size := utf8.DecodeRuneInString(str)

	if !unicode.IsUpper(r) {
		return str
	}

	buf := &bytes.Buffer{}
	buf.WriteRune(unicode.ToLower(r))
	buf.WriteString(str[size:])
	return buf.String()
}

// Shuffle randomizes runes in a string and returns the result.
// It uses default random source in `math/rand`.
func Shuffle(str string) string {
	if str == "" {
		return str
	}

	runes := []rune(str)
	index := 0

	for i := len(runes) - 1; i > 0; i-- {
		index = rand.Intn(i + 1)

		if i != index {
			runes[i], runes[index] = runes[index], runes[i]
		}
	}

	return string(runes)
}

// ShuffleSource randomizes runes in a string with given random source.
func ShuffleSource(str string, src rand.Source) string {
	if str == "" {
		return str
	}

	runes := []rune(str)
	index := 0
	r := rand.New(src)

	for i := len(runes) - 1; i > 0; i-- {
		index = r.Intn(i + 1)

		if i != index {
			runes[i], runes[index] = runes[index], runes[i]
		}
	}

	return string(runes)
}

// Successor returns the successor to string.
//
// If there is one alphanumeric rune is found in string, increase the rune by 1.
// If increment generates a "carry", the rune to the left of it is incremented.
// This process repeats until there is no carry, adding an additional rune if necessary.
//
// If there is no alphanumeric rune, the rightmost rune will be increased by 1
// regardless whether the result is a valid rune or not.
//
// Only following characters are alphanumeric.
//     * a - z
//     * A - Z
//     * 0 - 9
//
// Samples (borrowed from ruby's String#succ document):
//     "abcd"      => "abce"
//     "THX1138"   => "THX1139"
//     "<<koala>>" => "<<koalb>>"
//     "1999zzz"   => "2000aaa"
//     "ZZZ9999"   => "AAAA0000"
//     "***"       => "**+"
func Successor(str string) string {
	if str == "" {
		return str
	}

	var r rune
	var i int
	carry := ' '
	runes := []rune(str)
	l := len(runes)
	lastAlphanumeric := l

	for i = l - 1; i >= 0; i-- {
		r = runes[i]

		if ('a' <= r && r <= 'y') ||
			('A' <= r && r <= 'Y') ||
			('0' <= r && r <= '8') {
			runes[i]++
			carry = ' '
			lastAlphanumeric = i
			break
		}

		switch r {
		case 'z':
			runes[i] = 'a'
			carry = 'a'
			lastAlphanumeric = i

		case 'Z':
			runes[i] = 'A'
			carry = 'A'
			lastAlphanumeric = i

		case '9':
			runes[i] = '0'
			carry = '0'
			lastAlphanumeric = i
		}
	}

	// Needs to add one character for carry.
	if i < 0 && carry != ' ' {
		buf := &bytes.Buffer{}
		buf.Grow(l + 4) // Reserve enough space for write.

		if lastAlphanumeric != 0 {
			buf.WriteString(str[:lastAlphanumeric])
		}

		buf.WriteRune(carry)

		for _, r = range runes[lastAlphanumeric:] {
			buf.WriteRune(r)
		}

		return buf.String()
	}

	// No alphanumeric character. Simply increase last rune's value.
	if lastAlphanumeric == l {
		runes[l-1]++
	}

	return string(runes)
}
