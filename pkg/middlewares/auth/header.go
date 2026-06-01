package auth

import (
	"net/http"
	"strings"
)

// deleteCanonicalHeaders removes every header in h whose name matches name
// when underscores are treated as dashes (case-insensitively).
func deleteCanonicalHeaders(h http.Header, name string) {
	canonical := http.CanonicalHeaderKey(name)
	for key := range h {
		if key == canonical || strings.EqualFold(strings.ReplaceAll(key, "_", "-"), canonical) {
			delete(h, key)
		}
	}
}
