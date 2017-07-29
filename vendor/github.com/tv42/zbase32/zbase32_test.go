package zbase32_test

import (
	"bytes"
	"testing"

	"github.com/tv42/zbase32"
)

type bitTestCase struct {
	bits    int
	decoded []byte
	encoded string
}

var bitTests = []bitTestCase{
	// Test cases from the spec
	{0, []byte{}, ""},
	{1, []byte{0}, "y"},
	{1, []byte{128}, "o"},
	{2, []byte{64}, "e"},
	{2, []byte{192}, "a"},
	{10, []byte{0, 0}, "yy"},
	{10, []byte{128, 128}, "on"},
	{20, []byte{139, 136, 128}, "tqre"},
	{24, []byte{240, 191, 199}, "6n9hq"},
	{24, []byte{212, 122, 4}, "4t7ye"},
	// Note: this test varies from what's in the spec by one character!
	{30, []byte{245, 87, 189, 12}, "6im54d"},

	// Edge cases we stumbled on, that are not covered above.
	{8, []byte{0xff}, "9h"},
	{11, []byte{0xff, 0xE0}, "99o"},
	{40, []byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{48, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999h"},
	{192, []byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "ab3sr1ix8fhfnuzaeo75fkn3a7xh8udk6jsiiko"},

	// Used in the docs.
	{20, []byte{0x10, 0x11, 0x10}, "nyet"},
	{24, []byte{0x10, 0x11, 0x10}, "nyety"},
}

type byteTestCase struct {
	decoded []byte
	encoded string
}

var byteTests = []byteTestCase{
	// Byte-aligned test cases from the spec
	{[]byte{240, 191, 199}, "6n9hq"},
	{[]byte{212, 122, 4}, "4t7ye"},

	// Edge cases we stumbled on, that are not covered above.
	{[]byte{0xff}, "9h"},
	{[]byte{0xb5}, "sw"},
	{[]byte{0x34, 0x5a}, "gtpy"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, "99999999"},
	{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "999999999h"},
	{[]byte{
		0xc0, 0x73, 0x62, 0x4a, 0xaf, 0x39, 0x78, 0x51,
		0x4e, 0xf8, 0x44, 0x3b, 0xb2, 0xa8, 0x59, 0xc7,
		0x5f, 0xc3, 0xcc, 0x6a, 0xf2, 0x6d, 0x5a, 0xaa,
	}, "ab3sr1ix8fhfnuzaeo75fkn3a7xh8udk6jsiiko"},
}

func TestEncodeBits(t *testing.T) {
	for _, tc := range bitTests {
		dst := make([]byte, zbase32.EncodedLen(len(tc.decoded)))
		n := zbase32.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsString(t *testing.T) {
	for _, tc := range bitTests {
		s := zbase32.EncodeBitsToString(tc.decoded, tc.bits)
		if g, e := s, tc.encoded; g != e {
			t.Errorf("EncodeBitsToString %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBytes(t *testing.T) {
	for _, tc := range byteTests {
		dst := make([]byte, zbase32.EncodedLen(len(tc.decoded)))
		n := zbase32.Encode(dst, tc.decoded)
		dst = dst[:n]

		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("Encode %x wrong result: %q != %q", tc.decoded, g, e)
			continue
		}
	}
}

func TestEncodeBitsMasksExcess(t *testing.T) {
	for _, tc := range []bitTestCase{
		{0, []byte{255, 255}, ""},
		{1, []byte{255, 255}, "o"},
		{2, []byte{255, 255}, "a"},
		{3, []byte{255, 255}, "h"},
		{4, []byte{255, 255}, "6"},
		{5, []byte{255, 255}, "9"},
		{6, []byte{255, 255}, "9o"},
		{7, []byte{255, 255}, "9a"},
		{8, []byte{255, 255}, "9h"},
		{9, []byte{255, 255}, "96"},
		{10, []byte{255, 255}, "99"},
		{11, []byte{255, 255}, "99o"},
		{12, []byte{255, 255}, "99a"},
		{13, []byte{255, 255}, "99h"},
		{14, []byte{255, 255}, "996"},
		{15, []byte{255, 255}, "999"},
		{16, []byte{255, 255}, "999o"},
	} {
		dst := make([]byte, zbase32.EncodedLen(len(tc.decoded)))
		n := zbase32.EncodeBits(dst, tc.decoded, tc.bits)
		dst = dst[:n]
		if g, e := string(dst), tc.encoded; g != e {
			t.Errorf("EncodeBits %d bits of %x wrong result: %q != %q", tc.bits, tc.decoded, g, e)
		}
	}
}

func TestDecodeBits(t *testing.T) {
	for _, tc := range bitTests {
		dst := make([]byte, zbase32.DecodedLen(len(tc.encoded)))
		n, err := zbase32.DecodeBits(dst, []byte(tc.encoded), tc.bits)
		dst = dst[:n]
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBitsString(t *testing.T) {
	for _, tc := range bitTests {
		dec, err := zbase32.DecodeBitsString(tc.encoded, tc.bits)
		if err != nil {
			t.Errorf("DecodeBits %d bits from %q: error: %v", tc.bits, tc.encoded, err)
			continue
		}
		if g, e := dec, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("DecodeBits %d bits from %q, %x != %x", tc.bits, tc.encoded, g, e)
		}
	}
}

func TestDecodeBytes(t *testing.T) {
	for _, tc := range byteTests {
		dst := make([]byte, zbase32.DecodedLen(len(tc.encoded)))
		n, err := zbase32.Decode(dst, []byte(tc.encoded))
		dst = dst[:n]
		if err != nil {
			t.Errorf("Decode %q: error: %v", tc.encoded, err)
			continue
		}
		if g, e := dst, tc.decoded; !bytes.Equal(g, e) {
			t.Errorf("Decode %q, %x != %x", tc.encoded, g, e)
		}
	}
}

func TestDecodeBad(t *testing.T) {
	input := `foo!bar`
	_, err := zbase32.DecodeString(input)
	switch err := err.(type) {
	case nil:
		t.Fatalf("expected error from bad decode")
	case zbase32.CorruptInputError:
		if g, e := err.Error(), `illegal z-base-32 data at input byte 3`; g != e {
			t.Fatalf("wrong error: %q != %q", g, e)
		}
	default:
		t.Fatalf("wrong error from bad decode: %T: %v", err, err)
	}
}
