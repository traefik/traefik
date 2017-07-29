package glob

import (
	"strings"
	"testing"
)

func testGlobMatch(t *testing.T, pattern, subj string) {
	if !Glob(pattern, subj) {
		t.Fatalf("%s should match %s", pattern, subj)
	}
}

func testGlobNoMatch(t *testing.T, pattern, subj string) {
	if Glob(pattern, subj) {
		t.Fatalf("%s should not match %s", pattern, subj)
	}
}

func TestEmptyPattern(t *testing.T) {
	testGlobMatch(t, "", "")
	testGlobNoMatch(t, "", "test")
}

func TestEmptySubject(t *testing.T) {
	for _, pattern := range []string{
		"",
		"*",
		"**",
		"***",
		"****************",
		strings.Repeat("*", 1000000),
	} {
		testGlobMatch(t, pattern, "")
	}

	for _, pattern := range []string{
		// No globs/non-glob characters
		"test",
		"*test*",

		// Trailing characters
		"*x",
		"*****************x",
		strings.Repeat("*", 1000000) + "x",

		// Leading characters
		"x*",
		"x*****************",
		"x" + strings.Repeat("*", 1000000),

		// Mixed leading/trailing characters
		"x*x",
		"x****************x",
		"x" + strings.Repeat("*", 1000000) + "x",
	} {
		testGlobNoMatch(t, pattern, "")
	}
}

func TestPatternWithoutGlobs(t *testing.T) {
	testGlobMatch(t, "test", "test")
}

func TestGlob(t *testing.T) {
	// Matches
	for _, pattern := range []string{
		"*test",           // Leading glob
		"this*",           // Trailing glob
		"this*test",       // Middle glob
		"*is *",           // String in between two globs
		"*is*a*",          // Lots of globs
		"**test**",        // Double glob characters
		"**is**a***test*", // Varying number of globs
		"* *",             // White space between globs
		"*",               // Lone glob
		"**********",      // Nothing but globs
		"*Ѿ*",             // Unicode with globs
		"*is a ϗѾ *",      // Mixed ASCII/unicode
	} {
		testGlobMatch(t, pattern, "this is a ϗѾ test")
	}

	// Non-matches
	for _, pattern := range []string{
		"test*",               // Implicit substring match
		"*is",                 // Partial match
		"*no*",                // Globs without a match between them
		" ",                   // Plain white space
		"* ",                  // Trailing white space
		" *",                  // Leading white space
		"*ʤ*",                 // Non-matching unicode
		"this*this is a test", // Repeated prefix
	} {
		testGlobNoMatch(t, pattern, "this is a test")
	}
}

func BenchmarkGlob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !Glob("*quick*fox*dog", "The quick brown fox jumped over the lazy dog") {
			b.Fatalf("should match")
		}
	}
}
