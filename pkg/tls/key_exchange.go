package tls

import "crypto/tls"

// GetKeyExchangeName returns the name of the key exchange mechanism used in the TLS handshake
// (e.g. "X25519", "CurveP256"). For TLS 1.3 this is always set. For TLS 1.2 it is only set
// when an elliptic curve key exchange was used; an empty string is returned otherwise.
func GetKeyExchangeName(connState *tls.ConnectionState) string {
	if connState.CurveID == 0 {
		return ""
	}

	return connState.CurveID.String()
}
