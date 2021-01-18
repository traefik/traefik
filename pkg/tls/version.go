package tls

import "crypto/tls"

// VersionsReversed Map of TLS versions from crypto/tls
// Available TLS versions defined at https://golang.org/pkg/crypto/tls/#pkg-constants
var VersionsReversed = map[uint16]string{
	tls.VersionTLS10: "1.0",
	tls.VersionTLS11: "1.1",
	tls.VersionTLS12: "1.2",
	tls.VersionTLS13: "1.3",
}
