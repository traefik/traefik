package zbase32_test

import (
	"bytes"
	"flag"
	"testing"

	"github.com/tv42/zbase32"
)

func TestFlagString(t *testing.T) {
	v := zbase32.Value{0x34, 0x5a}
	if g, e := v.String(), "gtpy"; g != e {
		t.Errorf("wrong String: %q != %q", g, e)
	}
}

func TestFlagSet(t *testing.T) {
	var f flag.FlagSet
	var v zbase32.Value
	f.Var(&v, "frob", "input for frobnication")
	if err := f.Parse([]string{"-frob=pb1sa5dxfoo8q551pt1yw"}); err != nil {
		t.Fatalf("parsing flags: %v", err)
	}
	if g, e := "hello, world\n", string(v); g != e {
		t.Errorf("wrong decode: %q != %q", g, e)
	}
}

func TestFlagSetError(t *testing.T) {
	var v zbase32.Value
	switch err := v.Set("bad input!"); err := err.(type) {
	case zbase32.CorruptInputError:
		// ok
	default:
		t.Fatalf("wrong error: %T: %v", err, err)
	}
}

func TestFlagGet(t *testing.T) {
	v := zbase32.Value{0x34, 0x5a}
	switch b := v.Get().(type) {
	case []byte:
		if g, e := b, []byte{0x34, 0x5a}; !bytes.Equal(g, e) {
			t.Errorf("wrong Get: [% x] != [% x]", g, e)
		}
	default:
		t.Fatalf("wrong Get result: %T: %v", b, b)
	}
}
