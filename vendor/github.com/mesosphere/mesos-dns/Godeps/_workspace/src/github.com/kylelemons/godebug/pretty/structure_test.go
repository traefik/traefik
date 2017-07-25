// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pretty

import (
	"bytes"
	"testing"
)

func TestWriteTo(t *testing.T) {
	tests := []struct {
		desc     string
		node     node
		normal   string
		extended string
	}{
		{
			desc:     "string",
			node:     stringVal("zaphod"),
			normal:   `"zaphod"`,
			extended: `"zaphod"`,
		},
		{
			desc:     "raw",
			node:     rawVal("42"),
			normal:   `42`,
			extended: `42`,
		},
		{
			desc: "keyvals",
			node: keyvals{
				{"name", stringVal("zaphod")},
				{"age", rawVal("42")},
			},
			normal: `{name: "zaphod",
 age:  42}`,
			extended: `{
 name: "zaphod",
 age:  42,
}`,
		},
		{
			desc: "list",
			node: list{
				stringVal("zaphod"),
				rawVal("42"),
			},
			normal: `["zaphod",
 42]`,
			extended: `[
 "zaphod",
 42,
]`,
		},
		{
			desc: "nested",
			node: list{
				stringVal("first"),
				list{rawVal("1"), rawVal("2"), rawVal("3")},
				keyvals{
					{"trillian", keyvals{
						{"race", stringVal("human")},
						{"age", rawVal("36")},
					}},
					{"zaphod", keyvals{
						{"occupation", stringVal("president of the galaxy")},
						{"features", stringVal("two heads")},
					}},
				},
				keyvals{},
			},
			normal: `["first",
 [1,
  2,
  3],
 {trillian: {race: "human",
             age:  36},
  zaphod:   {occupation: "president of the galaxy",
             features:   "two heads"}},
 {}]`,
			extended: `[
 "first",
 [
  1,
  2,
  3,
 ],
 {
  trillian: {
             race: "human",
             age:  36,
            },
  zaphod:   {
             occupation: "president of the galaxy",
             features:   "two heads",
            },
 },
 {},
]`,
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		test.node.WriteTo(buf, "", &Config{})
		if got, want := buf.String(), test.normal; got != want {
			t.Errorf("%s: normal rendendered incorrectly\ngot:\n%s\nwant:\n%s", test.desc, got, want)
		}
		buf.Reset()
		test.node.WriteTo(buf, "", &Config{Diffable: true})
		if got, want := buf.String(), test.extended; got != want {
			t.Errorf("%s: extended rendendered incorrectly\ngot:\n%s\nwant:\n%s", test.desc, got, want)
		}
	}
}

func TestCompactString(t *testing.T) {
	tests := []struct {
		node
		compact string
	}{
		{
			stringVal("abc"),
			"abc",
		},
		{
			rawVal("2"),
			"2",
		},
		{
			list{
				rawVal("2"),
				rawVal("3"),
			},
			"[2,3]",
		},
		{
			keyvals{
				{"name", stringVal("zaphod")},
				{"age", rawVal("42")},
			},
			`{name:"zaphod",age:42}`,
		},
		{
			list{
				list{
					rawVal("0"),
					rawVal("1"),
					rawVal("2"),
					rawVal("3"),
				},
				list{
					rawVal("1"),
					rawVal("2"),
					rawVal("3"),
					rawVal("0"),
				},
				list{
					rawVal("2"),
					rawVal("3"),
					rawVal("0"),
					rawVal("1"),
				},
			},
			`[[0,1,2,3],[1,2,3,0],[2,3,0,1]]`,
		},
	}

	for _, test := range tests {
		if got, want := compactString(test.node), test.compact; got != want {
			t.Errorf("%#v: compact = %q, want %q", test.node, got, want)
		}
	}
}

func TestShortList(t *testing.T) {
	cfg := &Config{
		ShortList: 16,
	}

	tests := []struct {
		node
		want string
	}{
		{
			list{
				list{
					rawVal("0"),
					rawVal("1"),
					rawVal("2"),
					rawVal("3"),
				},
				list{
					rawVal("1"),
					rawVal("2"),
					rawVal("3"),
					rawVal("0"),
				},
				list{
					rawVal("2"),
					rawVal("3"),
					rawVal("0"),
					rawVal("1"),
				},
			},
			`[[0,1,2,3],
 [1,2,3,0],
 [2,3,0,1]]`,
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		test.node.WriteTo(buf, "", cfg)
		if got, want := buf.String(), test.want; got != want {
			t.Errorf("%#v: got:\n%s\nwant:\n%s", test.node, got, want)
		}
	}
}

var benchNode = keyvals{
	{"list", list{
		rawVal("0"),
		rawVal("1"),
		rawVal("2"),
		rawVal("3"),
	}},
	{"keyvals", keyvals{
		{"a", stringVal("b")},
		{"c", stringVal("e")},
		{"d", stringVal("f")},
	}},
}

func benchOpts(b *testing.B, cfg *Config) {
	buf := new(bytes.Buffer)
	benchNode.WriteTo(buf, "", cfg)
	b.SetBytes(int64(buf.Len()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		benchNode.WriteTo(buf, "", cfg)
	}
}

func BenchmarkWriteDefault(b *testing.B)   { benchOpts(b, DefaultConfig) }
func BenchmarkWriteShortList(b *testing.B) { benchOpts(b, &Config{ShortList: 16}) }
func BenchmarkWriteCompact(b *testing.B)   { benchOpts(b, &Config{Compact: true}) }
func BenchmarkWriteDiffable(b *testing.B)  { benchOpts(b, &Config{Diffable: true}) }
