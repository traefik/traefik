// Copyright (C) 2013-2018 by Maxim Bublis <b@codemonkey.ru>
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package uuid provides implementations of the Universally Unique Identifier (UUID), as specified in RFC-4122 and DCE 1.1.
//
// RFC-4122[1] provides the specification for versions 1, 3, 4, and 5.
//
// DCE 1.1[2] provides the specification for version 2.
//
// [1] https://tools.ietf.org/html/rfc4122
// [2] http://pubs.opengroup.org/onlinepubs/9696989899/chap5.htm#tagcjh_08_02_01_01
package uuid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

// Size of a UUID in bytes.
const Size = 16

// UUID is an array type to represent the value of a UUID, as defined in RFC-4122.
type UUID [Size]byte

// UUID versions.
const (
	_  byte = iota
	V1      // Version 1 (date-time and MAC address)
	V2      // Version 2 (date-time and MAC address, DCE security version)
	V3      // Version 3 (namespace name-based)
	V4      // Version 4 (random)
	V5      // Version 5 (namespace name-based)
)

// UUID layout variants.
const (
	VariantNCS byte = iota
	VariantRFC4122
	VariantMicrosoft
	VariantFuture
)

// UUID DCE domains.
const (
	DomainPerson = iota
	DomainGroup
	DomainOrg
)

// Timestamp is the count of 100-nanosecond intervals since 00:00:00.00,
// 15 October 1582 within a V1 UUID. This type has no meaning for V2-V5
// UUIDs since they don't have an embedded timestamp.
type Timestamp uint64

const _100nsPerSecond = 10000000

// Time returns the UTC time.Time representation of a Timestamp
func (t Timestamp) Time() (time.Time, error) {
	secs := uint64(t) / _100nsPerSecond
	nsecs := 100 * (uint64(t) % _100nsPerSecond)
	return time.Unix(int64(secs)-(epochStart/_100nsPerSecond), int64(nsecs)), nil
}

// TimestampFromV1 returns the Timestamp embedded within a V1 UUID.
// Returns an error if the UUID is any version other than 1.
func TimestampFromV1(u UUID) (Timestamp, error) {
	if u.Version() != 1 {
		err := fmt.Errorf("uuid: %s is version %d, not version 1", u, u.Version())
		return 0, err
	}
	low := binary.BigEndian.Uint32(u[0:4])
	mid := binary.BigEndian.Uint16(u[4:6])
	hi := binary.BigEndian.Uint16(u[6:8]) & 0xfff
	return Timestamp(uint64(low) + (uint64(mid) << 32) + (uint64(hi) << 48)), nil
}

// String parse helpers.
var (
	urnPrefix  = []byte("urn:uuid:")
	byteGroups = []int{8, 4, 4, 4, 12}
)

// Nil is the nil UUID, as specified in RFC-4122, that has all 128 bits set to
// zero.
var Nil = UUID{}

// Predefined namespace UUIDs.
var (
	NamespaceDNS  = Must(FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	NamespaceURL  = Must(FromString("6ba7b811-9dad-11d1-80b4-00c04fd430c8"))
	NamespaceOID  = Must(FromString("6ba7b812-9dad-11d1-80b4-00c04fd430c8"))
	NamespaceX500 = Must(FromString("6ba7b814-9dad-11d1-80b4-00c04fd430c8"))
)

// Version returns the algorithm version used to generate the UUID.
func (u UUID) Version() byte {
	return u[6] >> 4
}

// Variant returns the UUID layout variant.
func (u UUID) Variant() byte {
	switch {
	case (u[8] >> 7) == 0x00:
		return VariantNCS
	case (u[8] >> 6) == 0x02:
		return VariantRFC4122
	case (u[8] >> 5) == 0x06:
		return VariantMicrosoft
	case (u[8] >> 5) == 0x07:
		fallthrough
	default:
		return VariantFuture
	}
}

// Bytes returns a byte slice representation of the UUID.
func (u UUID) Bytes() []byte {
	return u[:]
}

// String returns a canonical RFC-4122 string representation of the UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

// SetVersion sets the version bits.
func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

// SetVariant sets the variant bits.
func (u *UUID) SetVariant(v byte) {
	switch v {
	case VariantNCS:
		u[8] = (u[8]&(0xff>>1) | (0x00 << 7))
	case VariantRFC4122:
		u[8] = (u[8]&(0xff>>2) | (0x02 << 6))
	case VariantMicrosoft:
		u[8] = (u[8]&(0xff>>3) | (0x06 << 5))
	case VariantFuture:
		fallthrough
	default:
		u[8] = (u[8]&(0xff>>3) | (0x07 << 5))
	}
}

// Must is a helper that wraps a call to a function returning (UUID, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations such as
//  var packageUUID = uuid.Must(uuid.FromString("123e4567-e89b-12d3-a456-426655440000"))
func Must(u UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return u
}
