package tracer

import (
	"log"
	"time"

	"golang.org/x/sys/windows"
)

// This method is more precise than the go1.8 time.Now on Windows
// See https://msdn.microsoft.com/en-us/library/windows/desktop/hh706895(v=vs.85).aspx
// It is however ~10x slower and requires Windows 8+.
func highPrecisionNow() int64 {
	var ft windows.Filetime
	windows.GetSystemTimePreciseAsFileTime(&ft)
	return ft.Nanoseconds()
}

func lowPrecisionNow() int64 {
	return time.Now().UTC().UnixNano()
}

var now func() int64

// If GetSystemTimePreciseAsFileTime is not available we default to the less
// precise implementation based on time.Now()
func init() {
	if err := windows.LoadGetSystemTimePreciseAsFileTime(); err != nil {
		log.Printf("Unable to load high precison timer, defaulting to time.Now()")
		now = lowPrecisionNow
	} else {
		log.Printf("Using high precision timer")
		now = highPrecisionNow
	}
}
