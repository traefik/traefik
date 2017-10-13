package cookie

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/containous/traefik/log"
)

const cookieNameLength = 6

// GetName of a cookie
func GetName(cookieName string, backendName string) string {
	if len(cookieName) != 0 {
		return sanitizeName(cookieName)
	}

	return GenerateName(backendName)
}

// GenerateName Generate a hashed name
func GenerateName(backendName string) string {
	data := []byte("_TRAEFIK_BACKEND_" + backendName)

	hash := sha1.New()
	_, err := hash.Write(data)
	if err != nil {
		// Impossible case
		log.Errorf("Fail to create cookie name: %v", err)
	}

	return fmt.Sprintf("_%x", hash.Sum(nil))[:cookieNameLength]
}

// sanitizeName According to [RFC 2616](https://www.ietf.org/rfc/rfc2616.txt) section 2.2
func sanitizeName(backend string) string {
	sanitizer := func(r rune) rune {
		switch r {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '`', '|', '~':
			return r
		}

		switch {
		case 'a' <= r && r <= 'z':
			fallthrough
		case 'A' <= r && r <= 'Z':
			fallthrough
		case '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}

	return strings.Map(sanitizer, backend)
}
