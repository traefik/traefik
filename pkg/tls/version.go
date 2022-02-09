package tls

import "crypto/tls"

// GetVersion returns the normalized TLS version.
// Available TLS versions defined at https://pkg.go.dev/crypto/tls/#pkg-constants
func GetVersion(connState *tls.ConnectionState) string {
	switch connState.Version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	}

	return "unknown"
}
