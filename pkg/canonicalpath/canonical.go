// Package canonicalpath provides canonical path representation for consistent
// routing decisions. This addresses CWE-436 (Interpretation Conflict) by ensuring
// all routing, middleware, and security decisions use the same path representation.
//
// The core invariant: A reverse proxy must guarantee that the exact same canonical
// path representation is used for router rule matching, middleware chain selection,
// and security boundary decisions.
//
// Key semantic distinction (RFC 3986):
//   - /admin/secret = TWO path segments: "admin" and "secret"
//   - /admin%2Fsecret = ONE path segment: "admin/secret" (literal slash in segment name)
//
// These are DIFFERENT endpoints and must NOT be conflated!
//
// See: https://cwe.mitre.org/data/definitions/436.html
package canonicalpath

import (
	"context"
	"net/url"
	"strings"
)

// contextKey is a private type used for context keys to avoid collisions.
type contextKey struct{}

// pathRepresentationKey is the context key for PathRepresentation.
var pathRepresentationKey = contextKey{}

// PathRepresentation holds both the original and canonical path forms.
// The original path is preserved for backend forwarding (HTTP transparency),
// while the canonical path is used for all internal routing decisions.
type PathRepresentation struct {
	// Original is the exact path as received from the client (escaped form).
	// This is preserved and forwarded to backends unchanged.
	Original string

	// Canonical is the normalized form used for ALL routing and security decisions.
	// - Unreserved characters are decoded (%61 → 'a')
	// - Reserved characters remain encoded (%2F stays as %2F)
	// - Case is normalized for hex digits (%2f → %2F)
	//
	// This ensures:
	// - /admin and /%61dmin canonicalize to the same value
	// - /admin%2Fsecret stays DIFFERENT from /admin/secret (semantic correctness)
	Canonical string

	// Segments contains the canonical path split into segments for efficient matching.
	// Note: For paths with encoded slashes, segments preserve the encoding.
	Segments []string
}

// NormalizationStrategy defines how paths are canonicalized.
type NormalizationStrategy int

const (
	// StrategyPreserveReserved keeps RFC 3986 reserved characters encoded in canonical form.
	// This is the CORRECT DEFAULT that preserves semantic meaning:
	// - /admin%2Fsecret (one segment) != /admin/secret (two segments)
	//
	// Use this strategy (the default) to maintain RFC 3986 semantics.
	// This is the secure choice because it doesn't alter the path's meaning.
	StrategyPreserveReserved NormalizationStrategy = iota

	// StrategyFullDecode fully decodes all percent-encoding for the canonical form.
	// WARNING: This CHANGES SEMANTIC MEANING - /admin%2Fsecret becomes /admin/secret!
	//
	// Only use this when:
	// 1. You know ALL backends decode reserved characters the same way
	// 2. You explicitly want /admin%2Fx and /admin/x to be treated identically
	// 3. You understand this may break legitimate use cases (GitLab, npm scoped packages)
	//
	// This strategy is provided for specific security scenarios where you want to
	// enforce that ALL path variations reach the same security boundaries.
	StrategyFullDecode
)

// DefaultStrategy returns the recommended default strategy.
// StrategyPreserveReserved is the default because it preserves RFC 3986 semantics.
func DefaultStrategy() NormalizationStrategy {
	return StrategyPreserveReserved
}

// Canonicalize creates a PathRepresentation from a raw URL path.
// The strategy determines how the canonical form is computed.
func Canonicalize(rawPath string, strategy NormalizationStrategy) PathRepresentation {
	if rawPath == "" {
		rawPath = "/"
	}

	var canonical string

	switch strategy {
	case StrategyPreserveReserved:
		canonical = preserveReservedCanonical(rawPath)
	case StrategyFullDecode:
		canonical = fullDecodeCanonical(rawPath)
	default:
		canonical = preserveReservedCanonical(rawPath)
	}

	return PathRepresentation{
		Original:  rawPath,
		Canonical: canonical,
		Segments:  splitPath(canonical),
	}
}

// reservedCharacters contains RFC 3986 reserved characters that affect path semantics.
// These must remain encoded to preserve the path's meaning.
var reservedCharacters = map[string]bool{
	"%3A": true, // :
	"%2F": true, // /
	"%3F": true, // ?
	"%23": true, // #
	"%5B": true, // [
	"%5D": true, // ]
	"%40": true, // @
	"%21": true, // !
	"%24": true, // $
	"%26": true, // &
	"%27": true, // '
	"%28": true, // (
	"%29": true, // )
	"%2A": true, // *
	"%2B": true, // +
	"%2C": true, // ,
	"%3B": true, // ;
	"%3D": true, // =
	"%25": true, // %
}

// preserveReservedCanonical normalizes the path while keeping reserved characters encoded.
// This is the semantically correct canonicalization per RFC 3986.
//
// Normalization steps:
// 1. Decode unreserved characters (%61 → 'a', %20 → ' ')
// 2. Keep reserved characters encoded (%2F stays as %2F)
// 3. Normalize percent-encoding case (%2f → %2F)
// 4. Normalize backslash to forward slash
// 5. Collapse multiple consecutive slashes
// 6. Ensure leading slash
func preserveReservedCanonical(raw string) string {
	var result strings.Builder
	result.Grow(len(raw))

	i := 0
	for i < len(raw) {
		if raw[i] == '%' && i+2 < len(raw) {
			encoded := strings.ToUpper(raw[i : i+3])
			if reservedCharacters[encoded] {
				// Keep reserved characters encoded (with normalized case)
				result.WriteString(encoded)
				i += 3
				continue
			}
			// Decode non-reserved characters
			decoded, err := url.PathUnescape(raw[i : i+3])
			if err != nil {
				// Invalid encoding - keep as-is
				result.WriteByte(raw[i])
				i++
				continue
			}
			result.WriteString(decoded)
			i += 3
		} else if raw[i] == '\\' {
			// Normalize backslash to forward slash
			result.WriteByte('/')
			i++
		} else {
			result.WriteByte(raw[i])
			i++
		}
	}

	normalized := result.String()

	// Collapse multiple slashes
	for strings.Contains(normalized, "//") {
		normalized = strings.ReplaceAll(normalized, "//", "/")
	}

	// Ensure leading slash
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}

	return normalized
}

// fullDecodeCanonical fully decodes and normalizes the path.
// WARNING: This changes semantic meaning - use with caution!
func fullDecodeCanonical(raw string) string {
	// Iteratively decode until no more percent-encoding remains
	// This handles double-encoding attacks like %252F -> %2F -> /
	decoded := raw
	for range 10 { // Max iterations to prevent infinite loops
		newDecoded, err := url.PathUnescape(decoded)
		if err != nil || newDecoded == decoded {
			break
		}
		decoded = newDecoded
	}

	// Normalize backslash to forward slash
	decoded = strings.ReplaceAll(decoded, "\\", "/")

	// Collapse multiple slashes
	for strings.Contains(decoded, "//") {
		decoded = strings.ReplaceAll(decoded, "//", "/")
	}

	// Handle path traversal (but don't use path.Clean as it has side effects)
	// Instead, just ensure basic normalization
	// Note: Path traversal should be handled separately as a security check

	// Ensure leading slash
	if !strings.HasPrefix(decoded, "/") {
		decoded = "/" + decoded
	}

	return decoded
}

// splitPath splits the canonical path into segments.
// For paths with encoded slashes, the encoded form is preserved in segments.
func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

// WithPathRepresentation adds the PathRepresentation to the request context.
func WithPathRepresentation(ctx context.Context, pr PathRepresentation) context.Context {
	return context.WithValue(ctx, pathRepresentationKey, pr)
}

// GetPathRepresentation retrieves the PathRepresentation from context.
// Returns the representation and true if found, or a default representation and false if not.
func GetPathRepresentation(ctx context.Context) (PathRepresentation, bool) {
	pr, ok := ctx.Value(pathRepresentationKey).(PathRepresentation)
	return pr, ok
}

// MustGetPathRepresentation retrieves the PathRepresentation from context.
// Panics if not found - use only when the middleware ordering guarantees presence.
func MustGetPathRepresentation(ctx context.Context) PathRepresentation {
	pr, ok := GetPathRepresentation(ctx)
	if !ok {
		panic("canonicalpath: PathRepresentation not in context - ensure CanonicalPathMiddleware is in the handler chain")
	}
	return pr
}

// GetCanonical is a convenience function to get just the canonical path string.
func GetCanonical(ctx context.Context) string {
	pr, ok := GetPathRepresentation(ctx)
	if !ok {
		return ""
	}
	return pr.Canonical
}

// GetOriginal is a convenience function to get just the original path string.
func GetOriginal(ctx context.Context) string {
	pr, ok := GetPathRepresentation(ctx)
	if !ok {
		return ""
	}
	return pr.Original
}

// HasPrefix checks if the canonical path has the given prefix.
func (pr PathRepresentation) HasPrefix(prefix string) bool {
	return strings.HasPrefix(pr.Canonical, prefix)
}

// Equals checks if the canonical path equals the given path.
func (pr PathRepresentation) Equals(p string) bool {
	return pr.Canonical == p
}

// ContainsNullByte checks if the path contains a null byte.
// Null bytes are never acceptable in paths and indicate an attack.
func (pr PathRepresentation) ContainsNullByte() bool {
	return strings.Contains(pr.Original, "%00") ||
		strings.Contains(pr.Original, "\x00") ||
		strings.Contains(pr.Canonical, "\x00")
}

// HasMalformedEncoding checks if the path has malformed percent-encoding.
func HasMalformedEncoding(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+2 >= len(s) {
				return true // truncated
			}
			if !isHex(s[i+1]) || !isHex(s[i+2]) {
				return true // invalid hex
			}
			i += 2
		}
	}
	return false
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

// StripPrefix removes a prefix from the canonical path representation.
// Returns a new PathRepresentation with the prefix stripped from both
// Original and Canonical forms, maintaining consistency.
//
// This should be used by middlewares instead of directly manipulating req.URL.Path
// to maintain the CWE-436 invariant.
func (pr PathRepresentation) StripPrefix(prefix string) PathRepresentation {
	if !pr.HasPrefix(prefix) {
		return pr // No change if prefix doesn't match
	}

	newCanonical := strings.TrimPrefix(pr.Canonical, prefix)
	if newCanonical == "" {
		newCanonical = "/"
	} else if !strings.HasPrefix(newCanonical, "/") {
		newCanonical = "/" + newCanonical
	}

	// Also strip from original, but need to handle encoding differences
	// For now, strip the same prefix (works when prefix doesn't contain encoded chars)
	newOriginal := strings.TrimPrefix(pr.Original, prefix)
	if newOriginal == "" {
		newOriginal = "/"
	} else if !strings.HasPrefix(newOriginal, "/") {
		newOriginal = "/" + newOriginal
	}

	return PathRepresentation{
		Original:  newOriginal,
		Canonical: newCanonical,
		Segments:  splitPath(newCanonical),
	}
}

// AddPrefix adds a prefix to the canonical path representation.
// Returns a new PathRepresentation with the prefix added to both forms.
func (pr PathRepresentation) AddPrefix(prefix string) PathRepresentation {
	newCanonical := prefix + pr.Canonical
	// Collapse any double slashes that might result
	for strings.Contains(newCanonical, "//") {
		newCanonical = strings.ReplaceAll(newCanonical, "//", "/")
	}

	newOriginal := prefix + pr.Original
	for strings.Contains(newOriginal, "//") {
		newOriginal = strings.ReplaceAll(newOriginal, "//", "/")
	}

	return PathRepresentation{
		Original:  newOriginal,
		Canonical: newCanonical,
		Segments:  splitPath(newCanonical),
	}
}
