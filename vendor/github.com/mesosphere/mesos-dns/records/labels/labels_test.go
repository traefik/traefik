package labels

import (
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

func BenchmarkDomainFrag_RFC952(b *testing.B) {
	benchmarkDomainFrag(b, RFC952)
}

func TestDomainFrag_RFC952(t *testing.T) {
	t.Parallel()
	testDomainFrag(t, "RFC952", RFC952, cases{
		"pod_123$abc.marathon-0.6.0-dev.mesos": "pod-123abc.marathon-0.dev.mesos",
	})
}

func BenchmarkDomainFrag_RFC1123(b *testing.B) {
	benchmarkDomainFrag(b, RFC1123)
}

func TestDomainFrag_RFC1123(t *testing.T) {
	t.Parallel()
	testDomainFrag(t, "RFC1123", RFC1123, cases{
		"pod_123$abc.marathon-0.6.0-dev.mesos": "pod-123abc.marathon-0.6.0-dev.mesos",
	})
}

func BenchmarkRFC952(b *testing.B) {
	const name = "89f.gsf---g_7-fgs--d7fddg-"
	for i := 0; i < b.N; i++ {
		RFC952(name)
	}
}

func TestRFC952(t *testing.T) {
	t.Parallel()

	testFunc(t, "RFC952", RFC952, cases{
		"4abc123":                                   "abc123",
		"-4abc123":                                  "abc123",
		"fd%gsf---gs7-f$gs--d7fddg-123":             "fdgsf---gs7-fgs--d7fddg1",
		"89fdgsf---gs7-fgs--d7fddg-123":             "fdgsf---gs7-fgs--d7fddg1",
		"89fdgsf---gs7-fgs--d7fddg---123":           "fdgsf---gs7-fgs--d7fddg1",
		"89fdgsf---gs7-fgs--d7fddg-":                "fdgsf---gs7-fgs--d7fddg",
		"chronos with a space AND MIXED CASE-2.0.1": "chronoswithaspaceandmixe",
		"chronos with a space AND----------MIXED--": "chronoswithaspaceandmixe",
	})

	quickCheckFunc(t, "RFC952", RFC952, cases{
		"doesn't start with numbers or dashes": func(s string) bool {
			return 0 != strings.IndexFunc(RFC952(s), func(r rune) bool {
				return r == '-' || (r >= '0' && r <= '9')
			})
		},
		"isn't longer than 24 chars": func(s string) bool {
			return len(RFC952(s)) <= 24
		},
	})
}

func BenchmarkRFC1123(b *testing.B) {
	const name = "##fdgsf---gs7-fgs--d7fddg123456789012345678901234567890123456789-"
	for i := 0; i < b.N; i++ {
		RFC1123(name)
	}
}

func TestRFC1123(t *testing.T) {
	t.Parallel()

	testFunc(t, "RFC1123", RFC1123, cases{
		"4abc123":                                                                "4abc123",
		"-4abc123":                                                               "4abc123",
		"89fdgsf---gs7-fgs--d7fddg-123":                                          "89fdgsf---gs7-fgs--d7fddg-123",
		"89fdgsf---gs7-fgs--d7fddg---123":                                        "89fdgsf---gs7-fgs--d7fddg---123",
		"89fdgsf---gs7-fgs--d7fddg-":                                             "89fdgsf---gs7-fgs--d7fddg",
		"fd%gsf---gs7-f$gs--d7fddg-123":                                          "fdgsf---gs7-fgs--d7fddg-123",
		"fd%gsf---gs7-f$gs--d7fddg123456789012345678901234567890123456789-123":   "fdgsf---gs7-fgs--d7fddg1234567890123456789012345678901234567891",
		"$$fdgsf---gs7-fgs--d7fddg123456789012345678901234567890123456789-123":   "fdgsf---gs7-fgs--d7fddg1234567890123456789012345678901234567891",
		"%%fdgsf---gs7-fgs--d7fddg123456789012345678901234567890123456789---123": "fdgsf---gs7-fgs--d7fddg1234567890123456789012345678901234567891",
		"##fdgsf---gs7-fgs--d7fddg123456789012345678901234567890123456789-":      "fdgsf---gs7-fgs--d7fddg123456789012345678901234567890123456789",
	})

	quickCheckFunc(t, "RFC1123", RFC1123, cases{
		"doesn't start with dashes": func(s string) bool {
			return strings.IndexRune(RFC1123(s), '-') != 0
		},
		"isn't longer than 63 chars": func(s string) bool {
			return len(RFC952(s)) <= 63
		},
	})
}

func testDomainFrag(t *testing.T, id string, label Func, special cases) {
	domfrag := func(s string) string { return DomainFrag(s, Sep, label) }
	testTransform(t, "DomainFrag."+id, domfrag, special, cases{
		"":                "",
		".":               "",
		"..":              "",
		"...":             "",
		"a":               "a",
		"abc":             "abc",
		"abc.":            "abc",
		".abc":            "abc",
		"a.c":             "a.c",
		".a.c":            "a.c",
		"a.c.":            "a.c",
		"a..c":            "a.c",
		"ab.c":            "ab.c",
		"ab.cd":           "ab.cd",
		"ab.cd.efg":       "ab.cd.efg",
		"a.c.e":           "a.c.e",
		"a..c.e":          "a.c.e",
		"a.c..e":          "a.c.e",
		"host.com":        "host.com",
		"space space.com": "spacespace.com",
		"blah-dash.com":   "blah-dash.com",
		"not$1234.com":    "not1234.com",
		"(@ host . com":   "host.com",
		"MiXeDcase.CoM":   "mixedcase.com",
	})
}

func testFunc(t *testing.T, id string, label Func, special cases) {
	testTransform(t, id, label, special, cases{
		"":                   "",
		"a":                  "a",
		"-":                  "",
		"a---":               "a",
		"---a---":            "a",
		"---a---b":           "a---b",
		"a.b.c.d.e":          "a-b-c-d-e",
		"a.c.d_de.":          "a-c-d-de",
		"abc123":             "abc123",
		"-abc123":            "abc123",
		"abc123-":            "abc123",
		"abc-123":            "abc-123",
		"abc--123":           "abc--123",
		"r29f.dev.angrypigs": "r29f-dev-angrypigs",
	})
}

func benchmarkDomainFrag(b *testing.B, label Func) {
	const name = "1www.pod_123$abc.marathon-0.6.0-dev.mesos"
	for i := 0; i < b.N; i++ {
		DomainFrag(name, Sep, label)
	}
}

func quickCheckFunc(t *testing.T, id string, label Func, special cases) {
	quickCheck(t, id, special, cases{
		"charset [0-9a-z-]": func(s string) bool {
			return -1 == strings.IndexFunc(label(s), func(r rune) bool {
				return !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'z') && r != '-'
			})
		},
	})
}

type cases map[string]interface{}

func quickCheck(t *testing.T, id string, special, base cases) {
	config := &quick.Config{MaxCount: 1e6}
	for prop, fn := range merge(special, base) {
		if testing.Short() {
			t.Logf("Skipping quick check of %s %q", id, prop)
		} else if err := quick.Check(fn, config); err != nil {
			t.Errorf("%s: property %q violated: %s", id, prop, err)
		}
	}
}

func testTransform(t *testing.T, id string, tr func(string) string, special, base cases) {
	for name, want := range merge(special, base) {
		if got := tr(name); !reflect.DeepEqual(got, want) {
			t.Errorf("%s(%q): got %q, want %q", id, name, got, want)
		}
	}
}

func merge(a, b cases) cases {
	c := make(cases, len(a)+len(b))
	for k, v := range b {
		c[k] = v
	}
	for k, v := range a {
		c[k] = v
	}
	return c
}
