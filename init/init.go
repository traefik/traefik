package init

import (
	"os"
	"strings"
)

// This makes use of the GODEBUG flag `http2xconnect` to deactivate the connect setting for HTTP2 by default.
// This type of upgrade is yet incompatible with `net/http` http1 reverse proxy.
// Please see https://github.com/golang/go/issues/71128#issuecomment-2574193636.
func init() {
	goDebug := os.Getenv("GODEBUG")
	if strings.Contains(goDebug, "http2xconnect") {
		return
	}

	if len(goDebug) > 0 {
		goDebug += ","
	}
	os.Setenv("GODEBUG", goDebug+"http2xconnect=0")
}
