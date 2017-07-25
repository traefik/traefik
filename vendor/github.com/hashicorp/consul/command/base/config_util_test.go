package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"path"
	"reflect"
)

func TestConfigUtil_Values(t *testing.T) {
	type config struct {
		B BoolValue     `mapstructure:"bool"`
		D DurationValue `mapstructure:"duration"`
		S StringValue   `mapstructure:"string"`
		U UintValue     `mapstructure:"uint"`
	}

	cases := []struct {
		in      string
		success string
		failure string
	}{
		{
			`{ }`,
			`"false" "0s" "" "0"`,
			"",
		},
		{
			`{ "bool": true, "duration": "2h", "string": "hello", "uint": 23 }`,
			`"true" "2h0m0s" "hello" "23"`,
			"",
		},
		{
			`{ "bool": "nope" }`,
			"",
			"got 'string'",
		},
		{
			`{ "duration": "nope" }`,
			"",
			"invalid duration nope",
		},
		{
			`{ "string": 123 }`,
			"",
			"got 'float64'",
		},
		{
			`{ "uint": -1 }`,
			"",
			"value cannot be negative",
		},
		{
			`{ "uint": 4294967296 }`,
			"",
			"value is too large",
		},
	}
	for i, c := range cases {
		var raw interface{}
		dec := json.NewDecoder(bytes.NewBufferString(c.in))
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("(case %d) err: %v", i, err)
		}

		var r config
		msdec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			DecodeHook:  ConfigDecodeHook,
			Result:      &r,
			ErrorUnused: true,
		})
		if err != nil {
			t.Fatalf("(case %d) err: %v", i, err)
		}

		err = msdec.Decode(raw)
		if c.failure != "" {
			if err == nil || !strings.Contains(err.Error(), c.failure) {
				t.Fatalf("(case %d) err: %v", i, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("(case %d) err: %v", i, err)
		}

		actual := fmt.Sprintf("%q %q %q %q",
			r.B.String(),
			r.D.String(),
			r.S.String(),
			r.U.String())
		if actual != c.success {
			t.Fatalf("(case %d) bad: %s", i, actual)
		}
	}
}

func TestConfigUtil_Visit(t *testing.T) {
	var trail []string
	visitor := func(path string) error {
		trail = append(trail, path)
		return nil
	}

	basePath := "../../test/command/merge"
	if err := Visit(basePath, visitor); err != nil {
		t.Fatalf("err: %v", err)
	}
	if err := Visit(path.Join(basePath, "subdir", "c.json"), visitor); err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := []string{
		path.Join(basePath, "a.json"),
		path.Join(basePath, "b.json"),
		path.Join(basePath, "nope"),
		path.Join(basePath, "zero.json"),
		path.Join(basePath, "subdir", "c.json"),
	}
	if !reflect.DeepEqual(trail, expected) {
		t.Fatalf("bad: %#v", trail)
	}
}
