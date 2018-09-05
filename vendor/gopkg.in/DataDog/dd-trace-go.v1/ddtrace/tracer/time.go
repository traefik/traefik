// +build !windows

package tracer

import "time"

// now returns current UTC time in nanos.
func now() int64 {
	return time.Now().UTC().UnixNano()
}
