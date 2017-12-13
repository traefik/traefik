package lz4

import (
	"encoding/binary"
	"errors"
)

// block represents a frame data block.
// Used when compressing or decompressing frame blocks concurrently.
type block struct {
	compressed bool
	zdata      []byte // compressed data
	data       []byte // decompressed data
	offset     int    // offset within the data as with block dependency the 64Kb window is prepended to it
	checksum   uint32 // compressed data checksum
	err        error  // error while [de]compressing
}

var (
	// ErrInvalidSource is returned by UncompressBlock when a compressed block is corrupted.
	ErrInvalidSource = errors.New("lz4: invalid source")
	// ErrShortBuffer is returned by UncompressBlock, CompressBlock or CompressBlockHC when
	// the supplied buffer for [de]compression is too small.
	ErrShortBuffer = errors.New("lz4: short buffer")
)

// CompressBlockBound returns the maximum size of a given buffer of size n, when not compressible.
func CompressBlockBound(n int) int {
	return n + n/255 + 16
}

// UncompressBlock decompresses the source buffer into the destination one,
// starting at the di index and returning the decompressed size.
//
// The destination buffer must be sized appropriately.
//
// An error is returned if the source data is invalid or the destination buffer is too small.
func UncompressBlock(src, dst []byte, di int) (int, error) {
	si, sn, di0 := 0, len(src), di
	if sn == 0 {
		return 0, nil
	}

	for {
		// literals and match lengths (token)
		lLen := int(src[si] >> 4)
		mLen := int(src[si] & 0xF)
		if si++; si == sn {
			return di, ErrInvalidSource
		}

		// literals
		if lLen > 0 {
			if lLen == 0xF {
				for src[si] == 0xFF {
					lLen += 0xFF
					if si++; si == sn {
						return di - di0, ErrInvalidSource
					}
				}
				lLen += int(src[si])
				if si++; si == sn {
					return di - di0, ErrInvalidSource
				}
			}
			if len(dst)-di < lLen || si+lLen > sn {
				return di - di0, ErrShortBuffer
			}
			di += copy(dst[di:], src[si:si+lLen])

			if si += lLen; si >= sn {
				return di - di0, nil
			}
		}

		if si += 2; si >= sn {
			return di, ErrInvalidSource
		}
		offset := int(src[si-2]) | int(src[si-1])<<8
		if di-offset < 0 || offset == 0 {
			return di - di0, ErrInvalidSource
		}

		// match
		if mLen == 0xF {
			for src[si] == 0xFF {
				mLen += 0xFF
				if si++; si == sn {
					return di - di0, ErrInvalidSource
				}
			}
			mLen += int(src[si])
			if si++; si == sn {
				return di - di0, ErrInvalidSource
			}
		}
		// minimum match length is 4
		mLen += 4
		if len(dst)-di <= mLen {
			return di - di0, ErrShortBuffer
		}

		// copy the match (NB. match is at least 4 bytes long)
		// NB. past di, copy() would write old bytes instead of
		// the ones we just copied, so split the work into the largest chunk.
		for ; mLen >= offset; mLen -= offset {
			di += copy(dst[di:], dst[di-offset:di])
		}
		di += copy(dst[di:], dst[di-offset:di-offset+mLen])
	}
}

// CompressBlock compresses the source buffer starting at soffet into the destination one.
// This is the fast version of LZ4 compression and also the default one.
//
// The size of the compressed data is returned. If it is 0 and no error, then the data is incompressible.
//
// An error is returned if the destination buffer is too small.
func CompressBlock(src, dst []byte, soffset int) (int, error) {
	sn, dn := len(src)-mfLimit, len(dst)
	if sn <= 0 || dn == 0 || soffset >= sn {
		return 0, nil
	}
	var si, di int

	// fast scan strategy:
	// we only need a hash table to store the last sequences (4 bytes)
	var hashTable [1 << hashLog]int
	var hashShift = uint((minMatch * 8) - hashLog)

	// Initialise the hash table with the first 64Kb of the input buffer
	// (used when compressing dependent blocks)
	for si < soffset {
		h := binary.LittleEndian.Uint32(src[si:]) * hasher >> hashShift
		si++
		hashTable[h] = si
	}

	anchor := si
	fma := 1 << skipStrength
	for si < sn-minMatch {
		// hash the next 4 bytes (sequence)...
		h := binary.LittleEndian.Uint32(src[si:]) * hasher >> hashShift
		// -1 to separate existing entries from new ones
		ref := hashTable[h] - 1
		// ...and store the position of the hash in the hash table (+1 to compensate the -1 upon saving)
		hashTable[h] = si + 1
		// no need to check the last 3 bytes in the first literal 4 bytes as
		// this guarantees that the next match, if any, is compressed with
		// a lower size, since to have some compression we must have:
		// ll+ml-overlap > 1 + (ll-15)/255 + (ml-4-15)/255 + 2 (uncompressed size>compressed size)
		// => ll+ml>3+2*overlap => ll+ml>= 4+2*overlap
		// and by definition we do have:
		// ll >= 1, ml >= 4
		// => ll+ml >= 5
		// => so overlap must be 0

		// the sequence is new, out of bound (64kb) or not valid: try next sequence
		if ref < 0 || fma&(1<<skipStrength-1) < 4 ||
			(si-ref)>>winSizeLog > 0 ||
			src[ref] != src[si] ||
			src[ref+1] != src[si+1] ||
			src[ref+2] != src[si+2] ||
			src[ref+3] != src[si+3] {
			// variable step: improves performance on non-compressible data
			si += fma >> skipStrength
			fma++
			continue
		}
		// match found
		fma = 1 << skipStrength
		lLen := si - anchor
		offset := si - ref

		// encode match length part 1
		si += minMatch
		mLen := si // match length has minMatch already
		for si <= sn && src[si] == src[si-offset] {
			si++
		}
		mLen = si - mLen
		if mLen < 0xF {
			dst[di] = byte(mLen)
		} else {
			dst[di] = 0xF
		}

		// encode literals length
		if lLen < 0xF {
			dst[di] |= byte(lLen << 4)
		} else {
			dst[di] |= 0xF0
			if di++; di == dn {
				return di, ErrShortBuffer
			}
			l := lLen - 0xF
			for ; l >= 0xFF; l -= 0xFF {
				dst[di] = 0xFF
				if di++; di == dn {
					return di, ErrShortBuffer
				}
			}
			dst[di] = byte(l)
		}
		if di++; di == dn {
			return di, ErrShortBuffer
		}

		// literals
		if di+lLen >= dn {
			return di, ErrShortBuffer
		}
		di += copy(dst[di:], src[anchor:anchor+lLen])
		anchor = si

		// encode offset
		if di += 2; di >= dn {
			return di, ErrShortBuffer
		}
		dst[di-2], dst[di-1] = byte(offset), byte(offset>>8)

		// encode match length part 2
		if mLen >= 0xF {
			for mLen -= 0xF; mLen >= 0xFF; mLen -= 0xFF {
				dst[di] = 0xFF
				if di++; di == dn {
					return di, ErrShortBuffer
				}
			}
			dst[di] = byte(mLen)
			if di++; di == dn {
				return di, ErrShortBuffer
			}
		}
	}

	if anchor == 0 {
		// incompressible
		return 0, nil
	}

	// last literals
	lLen := len(src) - anchor
	if lLen < 0xF {
		dst[di] = byte(lLen << 4)
	} else {
		dst[di] = 0xF0
		if di++; di == dn {
			return di, ErrShortBuffer
		}
		lLen -= 0xF
		for ; lLen >= 0xFF; lLen -= 0xFF {
			dst[di] = 0xFF
			if di++; di == dn {
				return di, ErrShortBuffer
			}
		}
		dst[di] = byte(lLen)
	}
	if di++; di == dn {
		return di, ErrShortBuffer
	}

	// write literals
	src = src[anchor:]
	switch n := di + len(src); {
	case n > dn:
		return di, ErrShortBuffer
	case n >= sn:
		// incompressible
		return 0, nil
	}
	di += copy(dst[di:], src)
	return di, nil
}

// CompressBlockHC compresses the source buffer starting at soffet into the destination one.
// CompressBlockHC compression ratio is better than CompressBlock but it is also slower.
//
// The size of the compressed data is returned. If it is 0 and no error, then the data is not compressible.
//
// An error is returned if the destination buffer is too small.
func CompressBlockHC(src, dst []byte, soffset int) (int, error) {
	sn, dn := len(src)-mfLimit, len(dst)
	if sn <= 0 || dn == 0 || soffset >= sn {
		return 0, nil
	}
	var si, di int

	// Hash Chain strategy:
	// we need a hash table and a chain table
	// the chain table cannot contain more entries than the window size (64Kb entries)
	var hashTable [1 << hashLog]int
	var chainTable [winSize]int
	var hashShift = uint((minMatch * 8) - hashLog)

	// Initialise the hash table with the first 64Kb of the input buffer
	// (used when compressing dependent blocks)
	for si < soffset {
		h := binary.LittleEndian.Uint32(src[si:]) * hasher >> hashShift
		chainTable[si&winMask] = hashTable[h]
		si++
		hashTable[h] = si
	}

	anchor := si
	for si < sn-minMatch {
		// hash the next 4 bytes (sequence)...
		h := binary.LittleEndian.Uint32(src[si:]) * hasher >> hashShift

		// follow the chain until out of window and give the longest match
		mLen := 0
		offset := 0
		for next := hashTable[h] - 1; next > 0 && next > si-winSize; next = chainTable[next&winMask] - 1 {
			// the first (mLen==0) or next byte (mLen>=minMatch) at current match length must match to improve on the match length
			if src[next+mLen] == src[si+mLen] {
				for ml := 0; ; ml++ {
					if src[next+ml] != src[si+ml] || si+ml > sn {
						// found a longer match, keep its position and length
						if mLen < ml && ml >= minMatch {
							mLen = ml
							offset = si - next
						}
						break
					}
				}
			}
		}
		chainTable[si&winMask] = hashTable[h]
		hashTable[h] = si + 1

		// no match found
		if mLen == 0 {
			si++
			continue
		}

		// match found
		// update hash/chain tables with overlaping bytes:
		// si already hashed, add everything from si+1 up to the match length
		for si, ml := si+1, si+mLen; si < ml; {
			h := binary.LittleEndian.Uint32(src[si:]) * hasher >> hashShift
			chainTable[si&winMask] = hashTable[h]
			si++
			hashTable[h] = si
		}

		lLen := si - anchor
		si += mLen
		mLen -= minMatch // match length does not include minMatch

		if mLen < 0xF {
			dst[di] = byte(mLen)
		} else {
			dst[di] = 0xF
		}

		// encode literals length
		if lLen < 0xF {
			dst[di] |= byte(lLen << 4)
		} else {
			dst[di] |= 0xF0
			if di++; di == dn {
				return di, ErrShortBuffer
			}
			l := lLen - 0xF
			for ; l >= 0xFF; l -= 0xFF {
				dst[di] = 0xFF
				if di++; di == dn {
					return di, ErrShortBuffer
				}
			}
			dst[di] = byte(l)
		}
		if di++; di == dn {
			return di, ErrShortBuffer
		}

		// literals
		if di+lLen >= dn {
			return di, ErrShortBuffer
		}
		di += copy(dst[di:], src[anchor:anchor+lLen])
		anchor = si

		// encode offset
		if di += 2; di >= dn {
			return di, ErrShortBuffer
		}
		dst[di-2], dst[di-1] = byte(offset), byte(offset>>8)

		// encode match length part 2
		if mLen >= 0xF {
			for mLen -= 0xF; mLen >= 0xFF; mLen -= 0xFF {
				dst[di] = 0xFF
				if di++; di == dn {
					return di, ErrShortBuffer
				}
			}
			dst[di] = byte(mLen)
			if di++; di == dn {
				return di, ErrShortBuffer
			}
		}
	}

	if anchor == 0 {
		// incompressible
		return 0, nil
	}

	// last literals
	lLen := len(src) - anchor
	if lLen < 0xF {
		dst[di] = byte(lLen << 4)
	} else {
		dst[di] = 0xF0
		if di++; di == dn {
			return di, ErrShortBuffer
		}
		lLen -= 0xF
		for ; lLen >= 0xFF; lLen -= 0xFF {
			dst[di] = 0xFF
			if di++; di == dn {
				return di, ErrShortBuffer
			}
		}
		dst[di] = byte(lLen)
	}
	if di++; di == dn {
		return di, ErrShortBuffer
	}

	// write literals
	src = src[anchor:]
	switch n := di + len(src); {
	case n > dn:
		return di, ErrShortBuffer
	case n >= sn:
		// incompressible
		return 0, nil
	}
	di += copy(dst[di:], src)
	return di, nil
}
