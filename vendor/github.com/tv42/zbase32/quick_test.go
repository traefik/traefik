package zbase32_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"testing/quick"

	"github.com/tv42/zbase32"
)

var python = flag.Bool("python", false, "run comparison tests against Python zbase32 library")

func calcBits(orig []byte, partial uint) int {
	bits := 8 * uint(len(orig))
	partial &= 7
	if bits > partial {
		bits -= partial
		mask := ^(byte(1<<partial) - 1)
		orig[len(orig)-1] &= mask
	}
	return int(bits)
}

func TestQuickRoundtripBits(t *testing.T) {
	fn := func(orig []byte, partial uint) bool {
		bits := calcBits(orig, partial)
		encoded := zbase32.EncodeBitsToString(orig, bits)
		decoded, err := zbase32.DecodeBitsString(encoded, bits)
		if err != nil {
			t.Logf("orig=\t%x", orig)
			t.Logf("bits=\t%d", bits)
			t.Logf("enc=\t%q", encoded)
			t.Fatalf("encode-decode roundtrip gave error: %v", err)
		}
		if !bytes.Equal(orig, decoded) {
			t.Logf("orig=\t%x", orig)
			t.Logf("dec=\t%x", decoded)
			t.Logf("bits=\t%d", bits)
			t.Logf("enc=\t%q", encoded)
			return false
		}
		return true
	}
	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}

func TestQuickRoundtripBytes(t *testing.T) {
	fn := func(orig []byte) bool {
		encoded := zbase32.EncodeToString(orig)
		decoded, err := zbase32.DecodeString(encoded)
		if err != nil {
			t.Logf("orig=\t%x", orig)
			t.Logf("enc=\t%q", encoded)
			t.Fatalf("encode-decode roundtrip gave error: %v", err)
		}
		if !bytes.Equal(orig, decoded) {
			t.Logf("orig=\t%x", orig)
			t.Logf("dec=\t%x", decoded)
			t.Logf("enc=\t%q", encoded)
			return false
		}
		return true
	}
	if err := quick.Check(fn, nil); err != nil {
		t.Fatal(err)
	}
}

func TestQuickPythonEncodeBits(t *testing.T) {
	if !*python {
		t.Skip("Skipping, use -python to enable")
	}
	us := func(orig []byte, partial uint) (string, error) {
		bits := calcBits(orig, partial)
		encoded := zbase32.EncodeBitsToString(orig, bits)
		return encoded, nil
	}
	them := func(orig []byte, partial uint) (string, error) {
		// the python library raises IndexError on zero-length input
		if len(orig) == 0 {
			return "", nil
		}
		bits := calcBits(orig, partial)
		cmd := exec.Command("python", "-c", `
import sys, zbase32
orig = sys.stdin.read()
bits = int(sys.argv[1])
sys.stdout.write(zbase32.b2a_l(orig, bits))
`,
			strconv.Itoa(bits),
		)
		cmd.Stdin = bytes.NewReader(orig)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("cannot run python: %v", err)
		}
		return string(output), nil
	}
	if err := quick.CheckEqual(us, them, nil); err != nil {
		t.Fatal(err)
	}
}

func TestQuickPythonDecodeBits(t *testing.T) {
	if !*python {
		t.Skip("Skipping, use -python to enable")
	}
	us := func(orig []byte, partial uint) ([]byte, error) {
		bits := calcBits(orig, partial)
		encoded := zbase32.EncodeBitsToString(orig, bits)
		return zbase32.DecodeBitsString(encoded, bits)
	}
	them := func(orig []byte, partial uint) ([]byte, error) {
		// the python library raises IndexError on zero-length input
		if len(orig) == 0 {
			return []byte{}, nil
		}
		bits := calcBits(orig, partial)
		encoded := zbase32.EncodeBitsToString(orig, bits)

		cmd := exec.Command("python", "-c", `
import sys, zbase32
enc = sys.stdin.read()
bits = int(sys.argv[1])
sys.stdout.write(zbase32.a2b_l(enc, bits))
`,
			strconv.Itoa(bits),
		)
		cmd.Stdin = strings.NewReader(encoded)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("cannot run python: %v", err)
		}
		return output, nil
	}
	if err := quick.CheckEqual(us, them, nil); err != nil {
		t.Fatal(err)
	}
}

func TestQuickPythonEncodeBytes(t *testing.T) {
	if !*python {
		t.Skip("Skipping, use -python to enable")
	}
	us := func(orig []byte) (string, error) {
		encoded := zbase32.EncodeToString(orig)
		return encoded, nil
	}
	them := func(orig []byte) (string, error) {
		// the python library raises IndexError on zero-length input
		if len(orig) == 0 {
			return "", nil
		}
		cmd := exec.Command("python", "-c", `
import sys, zbase32
orig = sys.stdin.read()
sys.stdout.write(zbase32.b2a(orig))
`)
		cmd.Stdin = bytes.NewReader(orig)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("cannot run python: %v", err)
		}
		return string(output), nil
	}
	if err := quick.CheckEqual(us, them, nil); err != nil {
		t.Fatal(err)
	}
}

func TestQuickPythonDecodeBytes(t *testing.T) {
	if !*python {
		t.Skip("Skipping, use -python to enable")
	}
	us := func(orig []byte) ([]byte, error) {
		encoded := zbase32.EncodeToString(orig)
		return zbase32.DecodeString(encoded)
	}
	them := func(orig []byte) ([]byte, error) {
		// the python library raises IndexError on zero-length input
		if len(orig) == 0 {
			return []byte{}, nil
		}
		encoded := zbase32.EncodeToString(orig)

		cmd := exec.Command("python", "-c", `
import sys, zbase32
enc = sys.stdin.read()
sys.stdout.write(zbase32.a2b(enc))
`)
		cmd.Stdin = strings.NewReader(encoded)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("cannot run python: %v", err)
		}
		return output, nil
	}
	if err := quick.CheckEqual(us, them, nil); err != nil {
		t.Fatal(err)
	}
}
