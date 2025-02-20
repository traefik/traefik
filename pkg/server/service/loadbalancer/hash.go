package loadbalancer

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"strconv"
)

// FnvHash returns the FNV-64 hash of the input string.
func FnvHash(input string) string {
	hasher := fnv.New64()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hasher.Write([]byte(input))

	return strconv.FormatUint(hasher.Sum64(), 16)
}

// Sha256Hash returns the SHA-256 hash (truncated to 16 characters) of the input string.
func Sha256Hash(input string) string {
	hash := sha256.New()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hash.Write([]byte(input))

	hashedInput := hex.EncodeToString(hash.Sum(nil))
	if len(hashedInput) < 16 {
		return hashedInput
	}
	return hashedInput[:16]
}
