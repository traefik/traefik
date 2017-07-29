package gax

import "testing"

func TestXGoogHeader(t *testing.T) {
	for _, tst := range []struct {
		kv   []string
		want string
	}{
		{nil, ""},
		{[]string{"abc", "def"}, "abc/def"},
		{[]string{"abc", "def", "xyz", "123", "foo", ""}, "abc/def xyz/123 foo/"},
	} {
		got := XGoogHeader(tst.kv...)
		if got != tst.want {
			t.Errorf("Header(%q) = %q, want %q", tst.kv, got, tst.want)
		}
	}
}
