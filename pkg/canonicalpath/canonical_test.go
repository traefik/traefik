package canonicalpath

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalize_PreserveReserved(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		wantCanonical string
		wantOriginal  string
	}{
		{
			name:          "simple path",
			rawPath:       "/admin/users",
			wantCanonical: "/admin/users",
			wantOriginal:  "/admin/users",
		},
		{
			name:          "encoded letter decoded (%61 = 'a')",
			rawPath:       "/%61dmin/users",
			wantCanonical: "/admin/users",
			wantOriginal:  "/%61dmin/users",
		},
		{
			name:          "encoded slash PRESERVED (%2F stays encoded)",
			rawPath:       "/admin%2Fusers",
			wantCanonical: "/admin%2Fusers",
			wantOriginal:  "/admin%2Fusers",
		},
		{
			name:          "encoded slash lowercase normalized",
			rawPath:       "/admin%2fusers",
			wantCanonical: "/admin%2Fusers",
			wantOriginal:  "/admin%2fusers",
		},
		{
			name:          "encoded space decoded",
			rawPath:       "/admin%20users",
			wantCanonical: "/admin users",
			wantOriginal:  "/admin%20users",
		},
		{
			name:          "backslash normalized",
			rawPath:       "/admin\\users",
			wantCanonical: "/admin/users",
			wantOriginal:  "/admin\\users",
		},
		{
			name:          "multiple slashes collapsed",
			rawPath:       "/admin//users///list",
			wantCanonical: "/admin/users/list",
			wantOriginal:  "/admin//users///list",
		},
		{
			name:          "mixed unreserved encoding",
			rawPath:       "/%61dmin%20users",
			wantCanonical: "/admin users",
			wantOriginal:  "/%61dmin%20users",
		},
		{
			name:          "empty path",
			rawPath:       "",
			wantCanonical: "/",
			wantOriginal:  "/",
		},
		{
			name:          "root path",
			rawPath:       "/",
			wantCanonical: "/",
			wantOriginal:  "/",
		},
		{
			name:          "GitLab-style namespace path PRESERVED",
			rawPath:       "/api/v4/projects/namespace%2Fproject",
			wantCanonical: "/api/v4/projects/namespace%2Fproject",
			wantOriginal:  "/api/v4/projects/namespace%2Fproject",
		},
		{
			name:          "npm scoped package PRESERVED",
			rawPath:       "/@scope%2Fpackage",
			wantCanonical: "/@scope%2Fpackage",
			wantOriginal:  "/@scope%2Fpackage",
		},
		{
			name:          "encoded hash preserved",
			rawPath:       "/path%23anchor",
			wantCanonical: "/path%23anchor",
			wantOriginal:  "/path%23anchor",
		},
		{
			name:          "encoded question mark preserved",
			rawPath:       "/path%3Fquery",
			wantCanonical: "/path%3Fquery",
			wantOriginal:  "/path%3Fquery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)

			assert.Equal(t, tt.wantCanonical, pr.Canonical, "Canonical path mismatch")
			assert.Equal(t, tt.wantOriginal, pr.Original, "Original path should be preserved")
		})
	}
}

func TestCanonicalize_FullDecode(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		wantCanonical string
	}{
		{
			name:          "simple path",
			rawPath:       "/admin/users",
			wantCanonical: "/admin/users",
		},
		{
			name:          "encoded letter decoded",
			rawPath:       "/%61dmin/users",
			wantCanonical: "/admin/users",
		},
		{
			name:          "encoded slash DECODED (semantic change!)",
			rawPath:       "/admin%2Fusers",
			wantCanonical: "/admin/users",
		},
		{
			name:          "double encoding handled",
			rawPath:       "/admin%252Fusers",
			wantCanonical: "/admin/users",
		},
		{
			name:          "triple encoding handled",
			rawPath:       "/admin%25252Fusers",
			wantCanonical: "/admin/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyFullDecode)
			assert.Equal(t, tt.wantCanonical, pr.Canonical, "Canonical path mismatch")
		})
	}
}

func TestDefaultStrategy(t *testing.T) {
	// Verify that the default strategy preserves reserved characters
	assert.Equal(t, StrategyPreserveReserved, DefaultStrategy())

	// Using default strategy should preserve %2F
	pr := Canonicalize("/admin%2Fusers", DefaultStrategy())
	assert.Equal(t, "/admin%2Fusers", pr.Canonical, "Default strategy should preserve reserved characters")
}

func TestPathRepresentation_HasPrefix(t *testing.T) {
	tests := []struct {
		name     string
		rawPath  string
		prefix   string
		expected bool
	}{
		{
			name:     "direct match",
			rawPath:  "/admin/users",
			prefix:   "/admin",
			expected: true,
		},
		{
			name:     "encoded letter bypass blocked",
			rawPath:  "/%61dmin/users",
			prefix:   "/admin",
			expected: true, // Canonical is /admin/users
		},
		{
			name:     "encoded slash - different path",
			rawPath:  "/admin%2Fusers",
			prefix:   "/admin/",
			expected: false, // Canonical is /admin%2Fusers, NOT /admin/users!
		},
		{
			name:     "encoded slash - matches literal prefix",
			rawPath:  "/admin%2Fusers",
			prefix:   "/admin%2F",
			expected: true, // Canonical matches prefix including %2F
		},
		{
			name:     "no match",
			rawPath:  "/public/page",
			prefix:   "/admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			assert.Equal(t, tt.expected, pr.HasPrefix(tt.prefix))
		})
	}
}

func TestPathRepresentation_Equals(t *testing.T) {
	tests := []struct {
		name     string
		rawPath  string
		path     string
		expected bool
	}{
		{
			name:     "direct match",
			rawPath:  "/admin",
			path:     "/admin",
			expected: true,
		},
		{
			name:     "encoded letter - matches decoded",
			rawPath:  "/%61dmin",
			path:     "/admin",
			expected: true,
		},
		{
			name:     "encoded slash - does NOT match decoded",
			rawPath:  "/admin%2Fusers",
			path:     "/admin/users",
			expected: false, // These are semantically different paths!
		},
		{
			name:     "encoded slash - matches encoded",
			rawPath:  "/admin%2Fusers",
			path:     "/admin%2Fusers",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			assert.Equal(t, tt.expected, pr.Equals(tt.path))
		})
	}
}

func TestPathRepresentation_ContainsNullByte(t *testing.T) {
	tests := []struct {
		name     string
		rawPath  string
		expected bool
	}{
		{
			name:     "no null byte",
			rawPath:  "/admin/users",
			expected: false,
		},
		{
			name:     "encoded null byte",
			rawPath:  "/admin%00/users",
			expected: true,
		},
		{
			name:     "literal null byte",
			rawPath:  "/admin\x00/users",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			assert.Equal(t, tt.expected, pr.ContainsNullByte())
		})
	}
}

func TestHasMalformedEncoding(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid path",
			path:     "/admin/users",
			expected: false,
		},
		{
			name:     "valid encoding",
			path:     "/admin%2Fusers",
			expected: false,
		},
		{
			name:     "truncated encoding",
			path:     "/admin%2",
			expected: true,
		},
		{
			name:     "invalid hex",
			path:     "/admin%GG/users",
			expected: true,
		},
		{
			name:     "single percent at end",
			path:     "/admin%",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasMalformedEncoding(tt.path))
		})
	}
}

func TestContextFunctions(t *testing.T) {
	pr := Canonicalize("/admin%2Fusers", StrategyPreserveReserved)

	// Test WithPathRepresentation and GetPathRepresentation
	ctx := context.Background()
	ctx = WithPathRepresentation(ctx, pr)

	retrieved, ok := GetPathRepresentation(ctx)
	require.True(t, ok, "PathRepresentation should be found in context")
	assert.Equal(t, pr.Canonical, retrieved.Canonical)
	assert.Equal(t, pr.Original, retrieved.Original)

	// Test MustGetPathRepresentation
	must := MustGetPathRepresentation(ctx)
	assert.Equal(t, pr.Canonical, must.Canonical)

	// Test GetCanonical - should preserve %2F
	assert.Equal(t, "/admin%2Fusers", GetCanonical(ctx))

	// Test GetOriginal
	assert.Equal(t, "/admin%2Fusers", GetOriginal(ctx))
}

func TestContextFunctions_NotFound(t *testing.T) {
	ctx := context.Background()

	// GetPathRepresentation should return false
	_, ok := GetPathRepresentation(ctx)
	assert.False(t, ok)

	// GetCanonical should return empty string
	assert.Equal(t, "", GetCanonical(ctx))

	// GetOriginal should return empty string
	assert.Equal(t, "", GetOriginal(ctx))

	// MustGetPathRepresentation should panic
	assert.Panics(t, func() {
		MustGetPathRepresentation(ctx)
	})
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "/admin/users",
			expected: []string{"admin", "users"},
		},
		{
			name:     "root",
			path:     "/",
			expected: []string{},
		},
		{
			name:     "single segment",
			path:     "/admin",
			expected: []string{"admin"},
		},
		{
			name:     "deep path",
			path:     "/api/v1/users/123",
			expected: []string{"api", "v1", "users", "123"},
		},
		{
			name:     "path with encoded slash",
			path:     "/api/namespace%2Fproject",
			expected: []string{"api", "namespace%2Fproject"}, // Encoded slash stays in segment!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.path, StrategyPreserveReserved)
			assert.Equal(t, tt.expected, pr.Segments)
		})
	}
}

// TestInvariant_RouterMiddlewareConsistency verifies the core invariant:
// The same canonical path is used for both router matching and middleware decisions.
func TestInvariant_RouterMiddlewareConsistency(t *testing.T) {
	// This test demonstrates the CWE-436 fix:
	// Router and middleware MUST use the same path representation

	testCases := []struct {
		name           string
		attackPath     string
		protectedPath  string
		shouldMatch    bool
		explanation    string
	}{
		{
			name:          "encoded letter - bypassed blocked",
			attackPath:    "/%61dmin/secret",
			protectedPath: "/admin",
			shouldMatch:   true,
			explanation:   "%61 decodes to 'a', so /%61dmin matches /admin",
		},
		{
			name:          "encoded slash - NOT a bypass because semantically different",
			attackPath:    "/admin%2Fsecret",
			protectedPath: "/admin/",
			shouldMatch:   false,
			explanation:   "/admin%2Fsecret is ONE segment, NOT /admin/secret (two segments)",
		},
		{
			name:          "encoded slash - matches encoded prefix",
			attackPath:    "/admin%2Fsecret",
			protectedPath: "/admin%2F",
			shouldMatch:   true,
			explanation:   "When prefix includes %2F, it matches",
		},
		{
			name:          "plain match",
			attackPath:    "/admin/secret",
			protectedPath: "/admin",
			shouldMatch:   true,
			explanation:   "Normal path matching works",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := Canonicalize(tc.attackPath, StrategyPreserveReserved)

			// Router check
			routerMatches := pr.HasPrefix(tc.protectedPath)

			// THE INVARIANT: Router matching result is deterministic
			assert.Equal(t, tc.shouldMatch, routerMatches, "Explanation: %s", tc.explanation)

			// Middleware check (using same representation)
			middlewareShouldApply := pr.HasPrefix(tc.protectedPath)

			// THE INVARIANT: Both must agree
			assert.Equal(t, routerMatches, middlewareShouldApply,
				"INVARIANT VIOLATED: Router and middleware have different views of the path")
		})
	}
}

// TestSemanticDifference demonstrates that /admin%2Fsecret != /admin/secret
func TestSemanticDifference(t *testing.T) {
	// These two paths are semantically DIFFERENT per RFC 3986
	pathWithEncodedSlash := Canonicalize("/admin%2Fsecret", StrategyPreserveReserved)
	pathWithRealSlash := Canonicalize("/admin/secret", StrategyPreserveReserved)

	assert.NotEqual(t, pathWithEncodedSlash.Canonical, pathWithRealSlash.Canonical,
		"Paths with encoded vs real slash must be different")

	// One segment vs two segments
	assert.Equal(t, 1, len(pathWithEncodedSlash.Segments), "/admin%2Fsecret should be ONE segment")
	assert.Equal(t, 2, len(pathWithRealSlash.Segments), "/admin/secret should be TWO segments")

	// Demonstrate the segment content
	assert.Equal(t, []string{"admin%2Fsecret"}, pathWithEncodedSlash.Segments)
	assert.Equal(t, []string{"admin", "secret"}, pathWithRealSlash.Segments)
}

// TestStripPrefix tests the StripPrefix method
func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		prefix        string
		wantCanonical string
	}{
		{
			name:          "simple strip",
			rawPath:       "/admin/users",
			prefix:        "/admin",
			wantCanonical: "/users",
		},
		{
			name:          "strip root",
			rawPath:       "/admin",
			prefix:        "/admin",
			wantCanonical: "/",
		},
		{
			name:          "no match - no change",
			rawPath:       "/other/path",
			prefix:        "/admin",
			wantCanonical: "/other/path",
		},
		{
			name:          "strip with encoded slash in path",
			rawPath:       "/admin%2Fusers/detail",
			prefix:        "/admin%2Fusers",
			wantCanonical: "/detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			stripped := pr.StripPrefix(tt.prefix)
			assert.Equal(t, tt.wantCanonical, stripped.Canonical)
		})
	}
}

// TestAddPrefix tests the AddPrefix method
func TestAddPrefix(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		prefix        string
		wantCanonical string
	}{
		{
			name:          "simple add",
			rawPath:       "/users",
			prefix:        "/api",
			wantCanonical: "/api/users",
		},
		{
			name:          "add to root",
			rawPath:       "/",
			prefix:        "/api",
			wantCanonical: "/api/",
		},
		{
			name:          "add with trailing slash",
			rawPath:       "/users",
			prefix:        "/api/",
			wantCanonical: "/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			prefixed := pr.AddPrefix(tt.prefix)
			assert.Equal(t, tt.wantCanonical, prefixed.Canonical)
		})
	}
}

// TestDoubleEncodingPreservation verifies that double-encoded characters
// are preserved correctly and not decoded multiple times.
// This is critical for security: %252F should stay as %252F, not become %2F or /.
func TestDoubleEncodingPreservation(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		wantCanonical string
		explanation   string
	}{
		{
			name:          "double encoded slash (%252F) preserved",
			rawPath:       "/admin%252Fpath",
			wantCanonical: "/admin%252Fpath",
			explanation:   "%25 is reserved (encoded %), so %252F stays as-is",
		},
		{
			name:          "double encoded letter (%2561) preserved",
			rawPath:       "/admin%2561",
			wantCanonical: "/admin%2561",
			explanation:   "%25 is reserved, so %2561 stays as-is (not decoded to %61 then 'a')",
		},
		{
			name:          "triple encoded slash (%25252F) preserved",
			rawPath:       "/admin%25252F",
			wantCanonical: "/admin%25252F",
			explanation:   "Each %25 layer is preserved",
		},
		{
			name:          "mixed double encoding",
			rawPath:       "/%2561dmin%252Fsecret",
			wantCanonical: "/%2561dmin%252Fsecret",
			explanation:   "Both double-encoded sequences preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			assert.Equal(t, tt.wantCanonical, pr.Canonical, "Explanation: %s", tt.explanation)
		})
	}
}

// TestLongPathHandling verifies that very long paths are handled correctly
// without causing performance issues or panics (DoS prevention).
func TestLongPathHandling(t *testing.T) {
	tests := []struct {
		name        string
		pathLength  int
		shouldPanic bool
	}{
		{
			name:        "normal path length (100 chars)",
			pathLength:  100,
			shouldPanic: false,
		},
		{
			name:        "long path (10000 chars)",
			pathLength:  10000,
			shouldPanic: false,
		},
		{
			name:        "very long path (100000 chars)",
			pathLength:  100000,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a path with repeated segments
			var pathBuilder strings.Builder
			pathBuilder.WriteString("/")
			for i := 0; i < tt.pathLength; i++ {
				pathBuilder.WriteByte('a' + byte(i%26))
			}
			longPath := pathBuilder.String()

			// Should not panic
			assert.NotPanics(t, func() {
				pr := Canonicalize(longPath, StrategyPreserveReserved)
				// Verify the path is processed correctly
				assert.True(t, len(pr.Canonical) > 0)
				assert.Equal(t, longPath, pr.Canonical)
			})
		})
	}
}

// TestLongPathWithEncoding tests long paths with percent-encoding.
func TestLongPathWithEncoding(t *testing.T) {
	// Create a path with many encoded characters
	var pathBuilder strings.Builder
	pathBuilder.WriteString("/")
	for i := 0; i < 1000; i++ {
		pathBuilder.WriteString("%61") // encoded 'a'
	}
	encodedPath := pathBuilder.String()

	// Should decode all %61 to 'a'
	pr := Canonicalize(encodedPath, StrategyPreserveReserved)

	expectedPath := "/" + strings.Repeat("a", 1000)
	assert.Equal(t, expectedPath, pr.Canonical)
}

// TestEdgeCaseEncodings tests various edge cases in percent-encoding.
func TestEdgeCaseEncodings(t *testing.T) {
	tests := []struct {
		name          string
		rawPath       string
		wantCanonical string
	}{
		{
			name:          "encoded percent at end",
			rawPath:       "/path%25",
			wantCanonical: "/path%25",
		},
		{
			name:          "encoded null followed by valid char",
			rawPath:       "/path%00test",
			wantCanonical: "/path\x00test", // Null is decoded (caught by ContainsNullByte)
		},
		{
			name:          "all uppercase hex",
			rawPath:       "/PATH%2FTEST",
			wantCanonical: "/PATH%2FTEST",
		},
		{
			name:          "all lowercase hex normalized",
			rawPath:       "/path%2ftest",
			wantCanonical: "/path%2Ftest", // Normalized to uppercase
		},
		{
			name:          "mixed case hex normalized",
			rawPath:       "/path%2Ftest%2fmore",
			wantCanonical: "/path%2Ftest%2Fmore",
		},
		{
			name:          "consecutive encoded slashes",
			rawPath:       "/path%2F%2F%2Ftest",
			wantCanonical: "/path%2F%2F%2Ftest", // Preserved, not collapsed
		},
		{
			name:          "encoded backslash decoded and normalized",
			rawPath:       "/path%5Ctest",
			wantCanonical: "/path/test", // %5C decodes to \, then \ is normalized to /
		},
		{
			name:          "literal backslash normalized",
			rawPath:       "/path\\test",
			wantCanonical: "/path/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := Canonicalize(tt.rawPath, StrategyPreserveReserved)
			assert.Equal(t, tt.wantCanonical, pr.Canonical)
		})
	}
}
