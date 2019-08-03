package env

import (
	"testing"

	"github.com/containous/traefik/v2/pkg/config/generator"
	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		desc     string
		environ  []string
		element  interface{}
		expected interface{}
	}{
		{
			desc:     "no env vars",
			environ:  nil,
			expected: nil,
		},
		{
			desc:    "bool value",
			environ: []string{"TRAEFIK_FOO=true"},
			element: &struct {
				Foo bool
			}{},
			expected: &struct {
				Foo bool
			}{
				Foo: true,
			},
		},
		{
			desc:    "equal",
			environ: []string{"TRAEFIK_FOO=bar"},
			element: &struct {
				Foo string
			}{},
			expected: &struct {
				Foo string
			}{
				Foo: "bar",
			},
		},
		{
			desc:    "multiple bool flags without value",
			environ: []string{"TRAEFIK_FOO=true", "TRAEFIK_BAR=true"},
			element: &struct {
				Foo bool
				Bar bool
			}{},
			expected: &struct {
				Foo bool
				Bar bool
			}{
				Foo: true,
				Bar: true,
			},
		},
		{
			desc:    "map string",
			environ: []string{"TRAEFIK_FOO_NAME=bar"},
			element: &struct {
				Foo map[string]string
			}{},
			expected: &struct {
				Foo map[string]string
			}{
				Foo: map[string]string{
					"name": "bar",
				},
			},
		},
		{
			desc:    "map struct",
			environ: []string{"TRAEFIK_FOO_NAME_VALUE=bar"},
			element: &struct {
				Foo map[string]struct{ Value string }
			}{},
			expected: &struct {
				Foo map[string]struct{ Value string }
			}{
				Foo: map[string]struct{ Value string }{
					"name": {
						Value: "bar",
					},
				},
			},
		},
		{
			desc:    "map struct with sub-struct",
			environ: []string{"TRAEFIK_FOO_NAME_BAR_VALUE=bar"},
			element: &struct {
				Foo map[string]struct {
					Bar *struct{ Value string }
				}
			}{},
			expected: &struct {
				Foo map[string]struct {
					Bar *struct{ Value string }
				}
			}{
				Foo: map[string]struct {
					Bar *struct{ Value string }
				}{
					"name": {
						Bar: &struct {
							Value string
						}{
							Value: "bar",
						},
					},
				},
			},
		},
		{
			desc:    "map struct with sub-map",
			environ: []string{"TRAEFIK_FOO_NAME1_BAR_NAME2_VALUE=bar"},
			element: &struct {
				Foo map[string]struct {
					Bar map[string]struct{ Value string }
				}
			}{},
			expected: &struct {
				Foo map[string]struct {
					Bar map[string]struct{ Value string }
				}
			}{
				Foo: map[string]struct {
					Bar map[string]struct{ Value string }
				}{
					"name1": {
						Bar: map[string]struct{ Value string }{
							"name2": {
								Value: "bar",
							},
						},
					},
				},
			},
		},
		{
			desc:    "slice",
			environ: []string{"TRAEFIK_FOO=bar,baz"},
			element: &struct {
				Foo []string
			}{},
			expected: &struct {
				Foo []string
			}{
				Foo: []string{"bar", "baz"},
			},
		},
		{
			desc:    "struct pointer value",
			environ: []string{"TRAEFIK_FOO=true"},
			element: &struct {
				Foo *struct{ Field string } `label:"allowEmpty"`
			}{},
			expected: &struct {
				Foo *struct{ Field string } `label:"allowEmpty"`
			}{
				Foo: &struct{ Field string }{},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := Decode(test.environ, DefaultNamePrefix, test.element)
			require.NoError(t, err)

			assert.Equal(t, test.expected, test.element)
		})
	}
}

func TestEncode(t *testing.T) {
	element := &Ya{
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
	}
	generator.Generate(element)

	flats, err := Encode(element)
	require.NoError(t, err)

	expected := []parser.Flat{
		{
			Name:        "TRAEFIK_FIELD1",
			Description: "",
			Default:     "bir",
		},
		{
			Name:        "TRAEFIK_FIELD10",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD10_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD11_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD12",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD13",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FIELD14",
			Description: "",
			Default:     "0",
		},
		{
			Name:        "TRAEFIK_FIELD15",
			Description: "",
			Default:     "7",
		},
		{
			Name:        "TRAEFIK_FIELD2",
			Description: "",
			Default:     "true",
		},
		{
			Name:        "TRAEFIK_FIELD3",
			Description: "",
			Default:     "0",
		},
		{
			Name:        "TRAEFIK_FIELD4_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD5_\u003cNAME\u003e",
			Description: "",
			Default:     "0",
		},
		{
			Name:        "TRAEFIK_FIELD6_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FIELD6_\u003cNAME\u003e_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD7_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FIELD7_\u003cNAME\u003e_FIELD_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD8_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FIELD8_\u003cNAME\u003e_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FIELD9_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FIELD9_\u003cNAME\u003e_FIELD_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN1",
			Description: "",
			Default:     "bar",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN10",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN10_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN11_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN12",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN13",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN14",
			Description: "",
			Default:     "0",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN2",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN3",
			Description: "",
			Default:     "1",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN4_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN5_\u003cNAME\u003e",
			Description: "",
			Default:     "0",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN6_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN6_\u003cNAME\u003e_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN7_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN7_\u003cNAME\u003e_FIELD_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN8_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN8_\u003cNAME\u003e_FIELD",
			Description: "",
			Default:     "",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN9_\u003cNAME\u003e",
			Description: "",
			Default:     "false",
		},
		{
			Name:        "TRAEFIK_FOO_FIELDIN9_\u003cNAME\u003e_FIELD_\u003cNAME\u003e",
			Description: "",
			Default:     "",
		},
	}

	assert.Equal(t, expected, flats)
}
