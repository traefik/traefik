package gettext

import "testing"

func assert_equal(t *testing.T, expected string, got string) {
	if expected != got {
		t.Logf("%s != %s", expected, got)
		t.Fail()
	}
}
