package generator

import (
	"testing"

	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected interface{}
	}{
		{
			desc: "nil",
		},
		{
			desc:    "simple",
			element: &Ya{},
			expected: &Ya{
				Foo: &Yaa{
					FieldIn1: "",
					FieldIn2: false,
					FieldIn3: 0,
					FieldIn4: map[string]string{
						parser.MapNamePlaceholder: "",
					},
					FieldIn5: map[string]int{
						parser.MapNamePlaceholder: 0,
					},
					FieldIn6: map[string]struct{ Field string }{
						parser.MapNamePlaceholder: {},
					},
					FieldIn7: map[string]struct{ Field map[string]string }{
						parser.MapNamePlaceholder: {
							Field: map[string]string{
								parser.MapNamePlaceholder: "",
							},
						},
					},
					FieldIn8: map[string]*struct{ Field string }{
						parser.MapNamePlaceholder: {},
					},
					FieldIn9: map[string]*struct{ Field map[string]string }{
						parser.MapNamePlaceholder: {
							Field: map[string]string{
								parser.MapNamePlaceholder: "",
							},
						},
					},
					FieldIn10: struct{ Field string }{},
					FieldIn11: &struct{ Field string }{},
					FieldIn12: func(v string) *string { return &v }(""),
					FieldIn13: func(v bool) *bool { return &v }(false),
					FieldIn14: func(v int) *int { return &v }(0),
				},
				Field1: "",
				Field2: false,
				Field3: 0,
				Field4: map[string]string{
					parser.MapNamePlaceholder: "",
				},
				Field5: map[string]int{
					parser.MapNamePlaceholder: 0,
				},
				Field6: map[string]struct{ Field string }{
					parser.MapNamePlaceholder: {},
				},
				Field7: map[string]struct{ Field map[string]string }{
					parser.MapNamePlaceholder: {
						Field: map[string]string{
							parser.MapNamePlaceholder: "",
						},
					},
				},
				Field8: map[string]*struct{ Field string }{
					parser.MapNamePlaceholder: {},
				},
				Field9: map[string]*struct{ Field map[string]string }{
					parser.MapNamePlaceholder: {
						Field: map[string]string{
							parser.MapNamePlaceholder: "",
						},
					},
				},
				Field10: struct{ Field string }{},
				Field11: &struct{ Field string }{},
				Field12: func(v string) *string { return &v }(""),
				Field13: func(v bool) *bool { return &v }(false),
				Field14: func(v int) *int { return &v }(0),
				Field15: []int{},
			},
		},
		{
			desc: "with initial state",
			element: &Ya{
				Foo: &Yaa{
					FieldIn1:  "bar",
					FieldIn2:  false,
					FieldIn3:  1,
					FieldIn4:  nil,
					FieldIn5:  nil,
					FieldIn6:  nil,
					FieldIn7:  nil,
					FieldIn8:  nil,
					FieldIn9:  nil,
					FieldIn10: struct{ Field string }{},
					FieldIn11: nil,
					FieldIn12: nil,
					FieldIn13: nil,
					FieldIn14: nil,
				},
				Field1:  "bir",
				Field2:  true,
				Field3:  0,
				Field4:  nil,
				Field5:  nil,
				Field6:  nil,
				Field7:  nil,
				Field8:  nil,
				Field9:  nil,
				Field10: struct{ Field string }{},
				Field11: nil,
				Field12: nil,
				Field13: nil,
				Field14: nil,
				Field15: []int{7},
			},
			expected: &Ya{
				Foo: &Yaa{
					FieldIn1: "bar",
					FieldIn2: false,
					FieldIn3: 1,
					FieldIn4: map[string]string{
						parser.MapNamePlaceholder: "",
					},
					FieldIn5: map[string]int{
						parser.MapNamePlaceholder: 0,
					},
					FieldIn6: map[string]struct{ Field string }{
						parser.MapNamePlaceholder: {},
					},
					FieldIn7: map[string]struct{ Field map[string]string }{
						parser.MapNamePlaceholder: {
							Field: map[string]string{
								parser.MapNamePlaceholder: "",
							},
						},
					},
					FieldIn8: map[string]*struct{ Field string }{
						parser.MapNamePlaceholder: {},
					},
					FieldIn9: map[string]*struct{ Field map[string]string }{
						parser.MapNamePlaceholder: {
							Field: map[string]string{
								parser.MapNamePlaceholder: "",
							},
						},
					},
					FieldIn10: struct{ Field string }{},
					FieldIn11: &struct{ Field string }{},
					FieldIn12: func(v string) *string { return &v }(""),
					FieldIn13: func(v bool) *bool { return &v }(false),
					FieldIn14: func(v int) *int { return &v }(0),
				},
				Field1: "bir",
				Field2: true,
				Field3: 0,
				Field4: map[string]string{
					parser.MapNamePlaceholder: "",
				},
				Field5: map[string]int{
					parser.MapNamePlaceholder: 0,
				},
				Field6: map[string]struct{ Field string }{
					parser.MapNamePlaceholder: {},
				},
				Field7: map[string]struct{ Field map[string]string }{
					parser.MapNamePlaceholder: {
						Field: map[string]string{
							parser.MapNamePlaceholder: "",
						},
					},
				},
				Field8: map[string]*struct{ Field string }{
					parser.MapNamePlaceholder: {},
				},
				Field9: map[string]*struct{ Field map[string]string }{
					parser.MapNamePlaceholder: {
						Field: map[string]string{
							parser.MapNamePlaceholder: "",
						},
					},
				},
				Field10: struct{ Field string }{},
				Field11: &struct{ Field string }{},
				Field12: func(v string) *string { return &v }(""),
				Field13: func(v bool) *bool { return &v }(false),
				Field14: func(v int) *int { return &v }(0),
				Field15: []int{7},
			},
		},
		{
			desc:    "setDefault",
			element: &Hu{},
			expected: &Hu{
				Foo: "hu",
				Fii: &Hi{
					Field: "hi",
				},
				Fuu: map[string]string{"<name>": ""},
				Fee: map[string]Hi{"<name>": {Field: "hi"}},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			Generate(test.element)

			assert.Equal(t, test.expected, test.element)
		})
	}
}

func Test_generate(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected interface{}
	}{
		{
			desc: "struct pointer",
			element: &struct {
				Foo string
				Fii *struct{ Field string }
			}{},
			expected: &struct {
				Foo string
				Fii *struct{ Field string }
			}{
				Foo: "",
				Fii: &struct{ Field string }{
					Field: "",
				},
			},
		},
		{
			desc: "string slice",
			element: &struct {
				Foo []string
			}{},
			expected: &struct {
				Foo []string
			}{
				Foo: []string{},
			},
		},
		{
			desc: "int slice",
			element: &struct {
				Foo []int
			}{},
			expected: &struct {
				Foo []int
			}{
				Foo: []int{},
			},
		},
		{
			desc: "struct slice",
			element: &struct {
				Foo []struct {
					Field string
				}
			}{},
			expected: &struct {
				Foo []struct {
					Field string
				}
			}{
				Foo: []struct {
					Field string
				}{
					{Field: ""},
				},
			},
		},
		{
			desc: "map string",
			element: &struct {
				Foo string
				Fii map[string]string
			}{},
			expected: &struct {
				Foo string
				Fii map[string]string
			}{
				Foo: "",
				Fii: map[string]string{
					parser.MapNamePlaceholder: "",
				},
			},
		},
		{
			desc: "map struct",
			element: &struct {
				Foo string
				Fii map[string]struct{ Field string }
			}{},
			expected: &struct {
				Foo string
				Fii map[string]struct{ Field string }
			}{
				Foo: "",
				Fii: map[string]struct{ Field string }{
					parser.MapNamePlaceholder: {},
				},
			},
		},
		{
			desc: "map struct pointer level 2",
			element: &struct {
				Foo string
				Fuu *struct {
					Fii map[string]*struct{ Field string }
				}
			}{},
			expected: &struct {
				Foo string
				Fuu *struct {
					Fii map[string]*struct{ Field string }
				}
			}{
				Foo: "",
				Fuu: &struct {
					Fii map[string]*struct {
						Field string
					}
				}{
					Fii: map[string]*struct{ Field string }{
						parser.MapNamePlaceholder: {
							Field: "",
						},
					},
				},
			},
		},
		{
			desc:    "SetDefaults",
			element: &Hu{},
			expected: &Hu{
				Foo: "hu",
				Fii: &Hi{
					Field: "hi",
				},
				Fuu: map[string]string{
					parser.MapNamePlaceholder: "",
				},
				Fee: map[string]Hi{
					parser.MapNamePlaceholder: {
						Field: "hi",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			generate(test.element)

			assert.Equal(t, test.expected, test.element)
		})
	}
}

type Hu struct {
	Foo string
	Fii *Hi
	Fuu map[string]string
	Fee map[string]Hi
}

func (h *Hu) SetDefaults() {
	h.Foo = "hu"
}

type Hi struct {
	Field string
}

func (h *Hi) SetDefaults() {
	h.Field = "hi"
}

type Ya struct {
	Foo     *Yaa
	Field1  string
	Field2  bool
	Field3  int
	Field4  map[string]string
	Field5  map[string]int
	Field6  map[string]struct{ Field string }
	Field7  map[string]struct{ Field map[string]string }
	Field8  map[string]*struct{ Field string }
	Field9  map[string]*struct{ Field map[string]string }
	Field10 struct{ Field string }
	Field11 *struct{ Field string }
	Field12 *string
	Field13 *bool
	Field14 *int
	Field15 []int
}

type Yaa struct {
	FieldIn1  string
	FieldIn2  bool
	FieldIn3  int
	FieldIn4  map[string]string
	FieldIn5  map[string]int
	FieldIn6  map[string]struct{ Field string }
	FieldIn7  map[string]struct{ Field map[string]string }
	FieldIn8  map[string]*struct{ Field string }
	FieldIn9  map[string]*struct{ Field map[string]string }
	FieldIn10 struct{ Field string }
	FieldIn11 *struct{ Field string }
	FieldIn12 *string
	FieldIn13 *bool
	FieldIn14 *int
}
