// +build go1.7

package strutil

// This is a generated file, editing would be silly.
// Go edit github.com/chipaca/term/gen/merged_rangetable.go instead.

import "unicode"

// using golang.org/x/text/unicode/rangetable do
// rangetable.Merge(unicode.Co, unicode.Cf, unicode.Cs, unicode.Noncharacter_Code_Point)
// (we also care about unicode.Cc but that's handled by hand)
// this makes the lookup in escape quite a bit faster (over 5×)
var Ctrl = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x00ad, Hi: 0x0600, Stride: 1363},
		{Lo: 0x0601, Hi: 0x0605, Stride: 1},
		{Lo: 0x061c, Hi: 0x06dd, Stride: 193},
		{Lo: 0x070f, Hi: 0x08e2, Stride: 467},
		{Lo: 0x180e, Hi: 0x200b, Stride: 2045},
		{Lo: 0x200c, Hi: 0x200f, Stride: 1},
		{Lo: 0x202a, Hi: 0x202e, Stride: 1},
		{Lo: 0x2060, Hi: 0x2064, Stride: 1},
		{Lo: 0x2066, Hi: 0x206f, Stride: 1},
		{Lo: 0xd800, Hi: 0xf8ff, Stride: 1},
		{Lo: 0xfdd0, Hi: 0xfdef, Stride: 1},
		{Lo: 0xfeff, Hi: 0xfff9, Stride: 250},
		{Lo: 0xfffa, Hi: 0xfffb, Stride: 1},
		{Lo: 0xfffe, Hi: 0xffff, Stride: 1},
	},
	R32: []unicode.Range32{
		{Lo: 0x110bd, Hi: 0x1bca0, Stride: 44003},
		{Lo: 0x1bca1, Hi: 0x1bca3, Stride: 1},
		{Lo: 0x1d173, Hi: 0x1d17a, Stride: 1},
		{Lo: 0x1fffe, Hi: 0x1ffff, Stride: 1},
		{Lo: 0x2fffe, Hi: 0x2ffff, Stride: 1},
		{Lo: 0x3fffe, Hi: 0x3ffff, Stride: 1},
		{Lo: 0x4fffe, Hi: 0x4ffff, Stride: 1},
		{Lo: 0x5fffe, Hi: 0x5ffff, Stride: 1},
		{Lo: 0x6fffe, Hi: 0x6ffff, Stride: 1},
		{Lo: 0x7fffe, Hi: 0x7ffff, Stride: 1},
		{Lo: 0x8fffe, Hi: 0x8ffff, Stride: 1},
		{Lo: 0x9fffe, Hi: 0x9ffff, Stride: 1},
		{Lo: 0xafffe, Hi: 0xaffff, Stride: 1},
		{Lo: 0xbfffe, Hi: 0xbffff, Stride: 1},
		{Lo: 0xcfffe, Hi: 0xcffff, Stride: 1},
		{Lo: 0xdfffe, Hi: 0xdffff, Stride: 1},
		{Lo: 0xe0001, Hi: 0xe0020, Stride: 31},
		{Lo: 0xe0021, Hi: 0xe007f, Stride: 1},
		{Lo: 0xefffe, Hi: 0x10ffff, Stride: 1},
	},
}
