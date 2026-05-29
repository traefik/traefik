// Package confighash computes a fast, allocation-light fingerprint of a
// Traefik configuration value.
//
// It replaces github.com/mitchellh/hashstructure on the Kubernetes-provider
// reload hot-path. The two differ in three ways:
//
//   - maps are combined order-independently with XOR over per-entry hashes
//     instead of sorting keys before hashing (no allocation per call),
//   - primitives are written straight into hash/maphash via encoding/binary
//     rather than boxed through reflect.Value.Interface,
//   - the seed is process-random (hash/maphash.MakeSeed), so the value is
//     stable within a single Traefik process but not across processes.
//
// Callers (provider reload-dedup, anonymous-stats payload) only compare the
// value to its previous self in the same process, so the per-process seed is
// the right trade-off: it gives us hash/maphash's AES-NI-accelerated mixing
// and DoS resistance without any portability concern.
package confighash

import (
	"encoding/binary"
	"hash/maphash"
	"math"
	"reflect"
	"sync"
)

// seed is shared by every hash computed in this process. Generated once at
// package init via hash/maphash.MakeSeed.
var seed = maphash.MakeSeed()

var hashPool = sync.Pool{
	New: func() any {
		return &maphash.Hash{}
	},
}

func acquire() *maphash.Hash {
	h := hashPool.Get().(*maphash.Hash)
	h.SetSeed(seed)
	h.Reset()
	return h
}

func release(h *maphash.Hash) { hashPool.Put(h) }

// Type tags ensure two values of different kinds with identical byte content
// never collide (e.g. uint64(0) vs nil pointer, []byte{} vs string "").
const (
	tagNil byte = iota + 1
	tagBool
	tagInt
	tagUint
	tagFloat
	tagComplex
	tagString
	tagBytes
	tagSlice
	tagArray
	tagStruct
	tagMap
	tagPtr
	tagInterface
	tagUnsupported
)

// Hash returns a 64-bit fingerprint of v.
//
// Two structurally-equal values produce the same hash within a single Traefik
// process. Map iteration order does not affect the result. The function never
// panics and never returns an error; unsupported kinds (chan, func, unsafe
// pointer) hash to a constant sentinel so they can appear inside an otherwise
// hashable struct.
func Hash(v any) uint64 {
	h := acquire()
	defer release(h)
	hashValue(h, reflect.ValueOf(v))
	return h.Sum64()
}

func hashValue(h *maphash.Hash, v reflect.Value) {
	if !v.IsValid() {
		_ = h.WriteByte(tagNil)
		return
	}

	switch v.Kind() {
	case reflect.Bool:
		var buf [2]byte
		buf[0] = tagBool
		if v.Bool() {
			buf[1] = 1
		}
		_, _ = h.Write(buf[:])

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		writeTagUint64(h, tagInt, uint64(v.Int()))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		writeTagUint64(h, tagUint, v.Uint())

	case reflect.Float32, reflect.Float64:
		writeTagUint64(h, tagFloat, math.Float64bits(v.Float()))

	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		var buf [17]byte
		buf[0] = tagComplex
		binary.LittleEndian.PutUint64(buf[1:9], math.Float64bits(real(c)))
		binary.LittleEndian.PutUint64(buf[9:17], math.Float64bits(imag(c)))
		_, _ = h.Write(buf[:])

	case reflect.String:
		writeTagUint64(h, tagString, uint64(v.Len()))
		_, _ = h.WriteString(v.String())

	case reflect.Ptr:
		if v.IsNil() {
			_ = h.WriteByte(tagNil)
			return
		}
		_ = h.WriteByte(tagPtr)
		hashValue(h, v.Elem())

	case reflect.Interface:
		if v.IsNil() {
			_ = h.WriteByte(tagNil)
			return
		}
		_ = h.WriteByte(tagInterface)
		hashValue(h, v.Elem())

	case reflect.Slice:
		if v.IsNil() {
			_ = h.WriteByte(tagNil)
			return
		}
		// Fast path for []byte: avoids per-element reflection.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			writeTagUint64(h, tagBytes, uint64(v.Len()))
			_, _ = h.Write(v.Bytes())
			return
		}
		writeTagUint64(h, tagSlice, uint64(v.Len()))
		for i := 0; i < v.Len(); i++ {
			hashValue(h, v.Index(i))
		}

	case reflect.Array:
		writeTagUint64(h, tagArray, uint64(v.Len()))
		for i := 0; i < v.Len(); i++ {
			hashValue(h, v.Index(i))
		}

	case reflect.Struct:
		_ = h.WriteByte(tagStruct)
		t := v.Type()
		n := v.NumField()
		for i := 0; i < n; i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			// Field name is mixed in so renaming or reordering a field
			// invalidates the hash even if the wire types stay the same.
			_, _ = h.WriteString(f.Name)
			hashValue(h, v.Field(i))
		}

	case reflect.Map:
		if v.IsNil() {
			_ = h.WriteByte(tagNil)
			return
		}
		// Combine per-entry hashes with XOR -- commutative and associative,
		// so iteration order does not matter and we never allocate a
		// sorted key list. Per-entry hashing borrows a sub-hasher from
		// the same pool.
		var combined uint64
		sub := acquire()
		iter := v.MapRange()
		for iter.Next() {
			sub.Reset()
			hashValue(sub, iter.Key())
			hashValue(sub, iter.Value())
			combined ^= sub.Sum64()
		}
		release(sub)
		// Mixing the map length distinguishes nil / empty / single-nil-entry
		// maps after the XOR collapses identical sub-hashes.
		var buf [17]byte
		buf[0] = tagMap
		binary.LittleEndian.PutUint64(buf[1:9], uint64(v.Len()))
		binary.LittleEndian.PutUint64(buf[9:17], combined)
		_, _ = h.Write(buf[:])

	default:
		// chan, func, unsafe.Pointer: not meaningful to hash structurally.
		_ = h.WriteByte(tagUnsupported)
	}
}

func writeTagUint64(h *maphash.Hash, tag byte, u uint64) {
	var buf [9]byte
	buf[0] = tag
	binary.LittleEndian.PutUint64(buf[1:], u)
	_, _ = h.Write(buf[:])
}
