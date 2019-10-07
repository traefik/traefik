package flag

import (
	"reflect"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/stretchr/testify/assert"
)

func Test_getFlagTypes(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected map[string]reflect.Kind
	}{
		{
			desc:     "nil",
			element:  nil,
			expected: map[string]reflect.Kind{},
		},
		{
			desc: "no fields",
			element: &struct {
			}{},
			expected: map[string]reflect.Kind{},
		},
		{
			desc: "string field",
			element: &struct {
				Foo string
			}{},
			expected: map[string]reflect.Kind{},
		},
		{
			desc: "bool field level 0",
			element: &struct {
				Foo bool
				fii bool
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Bool,
			},
		},
		{
			desc: "bool field level 1",
			element: &struct {
				Foo struct {
					Field bool
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo.field": reflect.Bool,
			},
		},
		{
			desc: "bool field level 2",
			element: &struct {
				Foo *struct {
					Fii *struct {
						Field bool
					}
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo":           reflect.Ptr,
				"foo.fii":       reflect.Ptr,
				"foo.fii.field": reflect.Bool,
			},
		},
		{
			desc: "pointer field",
			element: &struct {
				Foo *struct {
					Field string
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Ptr,
			},
		},
		{
			desc: "bool field level 3",
			element: &struct {
				Foo *struct {
					Fii *struct {
						Fuu *struct {
							Field bool
						}
					}
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo":               reflect.Ptr,
				"foo.fii":           reflect.Ptr,
				"foo.fii.fuu":       reflect.Ptr,
				"foo.fii.fuu.field": reflect.Bool,
			},
		},
		{
			desc: "map string",
			element: &struct {
				Foo map[string]string
			}{},
			expected: map[string]reflect.Kind{},
		},
		{
			desc: "map bool",
			element: &struct {
				Foo map[string]bool
				Fii struct{}
			}{},
			expected: map[string]reflect.Kind{
				"foo." + parser.MapNamePlaceholder: reflect.Bool,
			},
		},
		{
			desc: "map struct",
			element: &struct {
				Foo map[string]struct {
					Field bool
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo." + parser.MapNamePlaceholder + ".field": reflect.Bool,
			},
		},
		{
			desc: "map map bool",
			element: &struct {
				Foo map[string]map[string]bool
			}{},
			expected: map[string]reflect.Kind{
				"foo." + parser.MapNamePlaceholder + "." + parser.MapNamePlaceholder: reflect.Bool,
			},
		},
		{
			desc: "map struct map",
			element: &struct {
				Foo map[string]struct {
					Fii map[string]bool
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo." + parser.MapNamePlaceholder + ".fii." + parser.MapNamePlaceholder: reflect.Bool,
			},
		},
		{
			desc: "pointer bool field level 0",
			element: &struct {
				Foo *bool
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Bool,
			},
		},
		{
			desc: "pointer int field level 0",
			element: &struct {
				Foo *int
			}{},
			expected: map[string]reflect.Kind{},
		},
		{
			desc: "bool slice field level 0",
			element: &struct {
				Foo []bool
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Slice,
			},
		},
		{
			desc: "string slice field level 0",
			element: &struct {
				Foo []string
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Slice,
			},
		},
		{
			desc: "slice field level 1",
			element: &struct {
				Foo struct {
					Field []string
				}
			}{},
			expected: map[string]reflect.Kind{
				"foo.field": reflect.Slice,
			},
		},
		{
			desc: "map slice string",
			element: &struct {
				Foo map[string][]string
			}{},
			expected: map[string]reflect.Kind{
				"foo." + parser.MapNamePlaceholder: reflect.Slice,
			},
		},
		{
			desc: "embedded struct",
			element: &struct {
				Yo
			}{},
			expected: map[string]reflect.Kind{
				"foo": reflect.Bool,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFlagTypes(test.element)
			assert.Equal(t, test.expected, actual)
		})
	}
}

type Yo struct {
	Foo bool
}
