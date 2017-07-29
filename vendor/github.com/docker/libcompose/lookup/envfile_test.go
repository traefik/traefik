package lookup

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvfileLookupReturnsEmptyIfError(t *testing.T) {
	envfileLookup := &EnvfileLookup{
		Path: "anything/file.env",
	}
	actuals := envfileLookup.Lookup("any", nil)
	if len(actuals) != 0 {
		t.Fatalf("expected an empty slice, got %v", actuals)
	}
}

func TestEnvfileLookupWithGoodFile(t *testing.T) {
	content := `foo=bar
    baz=quux
# comment

_foobar=foobaz
with.dots=working
and_underscore=working too
`
	tmpFolder, err := ioutil.TempDir("", "test-envfile")
	if err != nil {
		t.Fatal(err)
	}
	envfile := filepath.Join(tmpFolder, ".env")
	if err := ioutil.WriteFile(envfile, []byte(content), 0700); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpFolder)

	envfileLookup := &EnvfileLookup{
		Path: envfile,
	}

	validateLookup(t, "baz=quux", envfileLookup.Lookup("baz", nil))
	validateLookup(t, "foo=bar", envfileLookup.Lookup("foo", nil))
	validateLookup(t, "_foobar=foobaz", envfileLookup.Lookup("_foobar", nil))
	validateLookup(t, "with.dots=working", envfileLookup.Lookup("with.dots", nil))
	validateLookup(t, "and_underscore=working too", envfileLookup.Lookup("and_underscore", nil))
}

func validateLookup(t *testing.T, expected string, actuals []string) {
	if len(actuals) != 1 {
		t.Fatalf("expected 1 result, got %v", actuals)
	}
	if actuals[0] != expected {
		t.Fatalf("expected %s, got %s", expected, actuals[0])
	}
}
