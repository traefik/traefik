package types

import (
	"fmt"
	"strconv"
)

// TraceID is a 128 bit number internally stored as 2x uint64 (high & low).
type TraceID struct {
	High uint64
	Low  uint64
}

// TraceIDFromHex returns the TraceID from a Hex string.
func TraceIDFromHex(h string) (t TraceID, err error) {
	if len(h) > 16 {
		if t.High, err = strconv.ParseUint(h[0:len(h)-16], 16, 64); err != nil {
			return
		}
		t.Low, err = strconv.ParseUint(h[len(h)-16:], 16, 64)
		return
	}
	t.Low, err = strconv.ParseUint(h, 16, 64)
	return
}

// ToHex outputs the 128-bit traceID as hex string.
func (t TraceID) ToHex() string {
	if t.High == 0 {
		return strconv.FormatUint(t.Low, 16)
	}
	return fmt.Sprintf(
		"%016s%016s", strconv.FormatUint(t.High, 16), strconv.FormatUint(t.Low, 16),
	)
}

// Empty returns if TraceID has zero value
func (t TraceID) Empty() bool {
	return t.Low == 0 && t.High == 0
}
