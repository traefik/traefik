package secure

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
)

type key int

const cspNonceKey key = iota

// CSPNonce returns the nonce value associated with the present request. If no nonce has been generated it returns an empty string.
func CSPNonce(c context.Context) string {
	if val, ok := c.Value(cspNonceKey).(string); ok {
		return val
	}

	return ""
}

// WithCSPNonce returns a context derived from ctx containing the given nonce as a value.
//
// This is intended for testing or more advanced use-cases;
// For ordinary HTTP handlers, clients can rely on this package's middleware to populate the CSP nonce in the context.
func WithCSPNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, cspNonceKey, nonce)
}

func withCSPNonce(r *http.Request, nonce string) *http.Request {
	return r.WithContext(WithCSPNonce(r.Context(), nonce))
}

func cspRandNonce() string {
	var buf [cspNonceSize]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic("CSP Nonce rand.Reader failed" + err.Error())
	}

	return base64.RawStdEncoding.EncodeToString(buf[:])
}
