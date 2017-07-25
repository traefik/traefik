package dns

import (
	"testing"
)

func TestCmToM(t *testing.T) {
	s := cmToM(0, 0)
	if s != "0.00" {
		t.Error("0, 0")
	}

	s = cmToM(1, 0)
	if s != "0.01" {
		t.Error("1, 0")
	}

	s = cmToM(3, 1)
	if s != "0.30" {
		t.Error("3, 1")
	}

	s = cmToM(4, 2)
	if s != "4" {
		t.Error("4, 2")
	}

	s = cmToM(5, 3)
	if s != "50" {
		t.Error("5, 3")
	}

	s = cmToM(7, 5)
	if s != "7000" {
		t.Error("7, 5")
	}

	s = cmToM(9, 9)
	if s != "90000000" {
		t.Error("9, 9")
	}
}

func TestSplitN(t *testing.T) {
	xs := splitN("abc", 5)
	if len(xs) != 1 && xs[0] != "abc" {
		t.Errorf("Failure to split abc")
	}

	s := ""
	for i := 0; i < 255; i++ {
		s += "a"
	}

	xs = splitN(s, 255)
	if len(xs) != 1 && xs[0] != s {
		t.Errorf("failure to split 255 char long string")
	}

	s += "b"
	xs = splitN(s, 255)
	if len(xs) != 2 || xs[1] != "b" {
		t.Errorf("failure to split 256 char long string: %d", len(xs))
	}

	// Make s longer
	for i := 0; i < 255; i++ {
		s += "a"
	}
	xs = splitN(s, 255)
	if len(xs) != 3 || xs[2] != "a" {
		t.Errorf("failure to split 510 char long string: %d", len(xs))
	}
}
