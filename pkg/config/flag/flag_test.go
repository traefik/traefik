package flag

import (
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/generator"
	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		desc     string
		args     []string
		element  interface{}
		expected interface{}
	}{
		{
			desc:     "no args",
			args:     nil,
			expected: nil,
		},
		{
			desc: "types.Duration value",
			args: []string{"--foo=1"},
			element: &struct {
				Foo types.Duration
			}{},
			expected: &struct {
				Foo types.Duration
			}{
				Foo: types.Duration(1 * time.Second),
			},
		},
		{
			desc: "time.Duration value",
			args: []string{"--foo=1"},
			element: &struct {
				Foo time.Duration
			}{},
			expected: &struct {
				Foo time.Duration
			}{
				Foo: 1 * time.Nanosecond,
			},
		},
		{
			desc: "bool value",
			args: []string{"--foo"},
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
			desc: "equal",
			args: []string{"--foo=bar"},
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
			desc: "space separated",
			args: []string{"--foo", "bar"},
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
			desc: "space separated with end of parameter",
			args: []string{"--foo=bir", "--", "--bar"},
			element: &struct {
				Foo string
			}{},
			expected: &struct {
				Foo string
			}{
				Foo: "bir",
			},
		},
		{
			desc: "multiple bool flags without value",
			args: []string{"--foo", "--bar"},
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
			desc: "slice with several flags",
			args: []string{"--foo=bar", "--foo=baz"},
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
			desc: "map string",
			args: []string{"--foo.name=bar"},
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
			desc: "map string case sensitive",
			args: []string{"--foo.caseSensitiveName=barBoo"},
			element: &struct {
				Foo map[string]string
			}{},
			expected: &struct {
				Foo map[string]string
			}{
				Foo: map[string]string{
					"caseSensitiveName": "barBoo",
				},
			},
		},
		{
			desc: "map struct",
			args: []string{"--foo.name.value=bar"},
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
			desc: "map struct with sub-struct",
			args: []string{"--foo.name.bar.value=bar"},
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
			desc: "map struct with sub-map",
			args: []string{"--foo.name1.bar.name2.value=bar"},
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
			desc: "slice with several flags 2",
			args: []string{"--foo", "bar", "--foo", "baz"},
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
			desc: "slice with several flags 3",
			args: []string{"--foo", "bar", "--foo=", "--baz"},
			element: &struct {
				Foo []string
				Baz bool
			}{},
			expected: &struct {
				Foo []string
				Baz bool
			}{
				Foo: []string{"bar", ""},
				Baz: true,
			},
		},
		{
			desc: "slice with several flags 4",
			args: []string{"--foo", "bar", "--foo", "--baz"},
			element: &struct {
				Foo []string
				Baz bool
			}{},
			expected: &struct {
				Foo []string
				Baz bool
			}{
				Foo: []string{"bar", "--baz"},
			},
		},
		{
			desc: "slice of struct",
			args: []string{
				"--foo[0].Field1", "bar", "--foo[0].Field2", "6",
				"--foo[1].Field1", "bur", "--foo[1].Field2", "2",
			},
			element: &struct {
				Foo []struct {
					Field1 string
					Field2 int
				}
			}{},
			expected: &struct {
				Foo []struct {
					Field1 string
					Field2 int
				}
			}{
				Foo: []struct {
					Field1 string
					Field2 int
				}{
					{
						Field1: "bar",
						Field2: 6,
					},
					{
						Field1: "bur",
						Field2: 2,
					},
				},
			},
		},
		{
			desc: "slice of pointer of struct",
			args: []string{
				"--foo[0].Field1", "bar", "--foo[0].Field2", "6",
				"--foo[1].Field1", "bur", "--foo[1].Field2", "2",
			},
			element: &struct {
				Foo []*struct {
					Field1 string
					Field2 int
				}
			}{},
			expected: &struct {
				Foo []*struct {
					Field1 string
					Field2 int
				}
			}{
				Foo: []*struct {
					Field1 string
					Field2 int
				}{
					{
						Field1: "bar",
						Field2: 6,
					},
					{
						Field1: "bur",
						Field2: 2,
					},
				},
			},
		},
		{
			desc: "multiple string flag",
			element: &struct {
				Foo string
			}{},
			args: []string{"--foo=bar", "--foo=baz"},
			expected: &struct {
				Foo string
			}{
				Foo: "baz",
			},
		},
		{
			desc: "multiple string flag 2",
			element: &struct {
				Foo string
			}{},
			args: []string{"--foo", "bar", "--foo", "baz"},
			expected: &struct {
				Foo string
			}{
				Foo: "baz",
			},
		},
		{
			desc: "string without value",
			element: &struct {
				Foo string
				Bar bool
			}{},
			args: []string{"--foo", "--bar"},
			expected: &struct {
				Foo string
				Bar bool
			}{
				Foo: "--bar",
			},
		},
		{
			desc: "struct pointer value",
			args: []string{"--foo"},
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

			err := Decode(test.args, test.element)
			require.NoError(t, err)

			assert.Equal(t, test.expected, test.element)
		})
	}
}

func TestEncode(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected []parser.Flat
	}{
		{
			desc: "string field",
			element: &struct {
				Field string `description:"field description"`
			}{
				Field: "test",
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "test",
			}},
		},
		{
			desc: "int field",
			element: &struct {
				Field int `description:"field description"`
			}{
				Field: 6,
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "6",
			}},
		},
		{
			desc: "bool field",
			element: &struct {
				Field bool `description:"field description"`
			}{
				Field: true,
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "true",
			}},
		},
		{
			desc: "string pointer field",
			element: &struct {
				Field *string `description:"field description"`
			}{
				Field: func(v string) *string { return &v }("test"),
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "test",
			}},
		},
		{
			desc: "int pointer field",
			element: &struct {
				Field *int `description:"field description"`
			}{
				Field: func(v int) *int { return &v }(6),
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "6",
			}},
		},
		{
			desc: "bool pointer field",
			element: &struct {
				Field *bool `description:"field description"`
			}{
				Field: func(v bool) *bool { return &v }(true),
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "true",
			}},
		},
		{
			desc: "slice of string field, no initial value",
			element: &struct {
				Field []string `description:"field description"`
			}{},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "",
			}},
		},
		{
			desc: "slice of string field, with initial value",
			element: &struct {
				Field []string `description:"field description"`
			}{
				Field: []string{"foo", "bar"},
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "foo, bar",
			}},
		},
		{
			desc: "slice of int field, no initial value",
			element: &struct {
				Field []int `description:"field description"`
			}{},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "",
			}},
		},
		{
			desc: "slice of int field, with initial value",
			element: &struct {
				Field []int `description:"field description"`
			}{
				Field: []int{6, 3},
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "6, 3",
			}},
		},
		{
			desc: "map string field",
			element: &struct {
				Field map[string]string `description:"field description"`
			}{
				Field: map[string]string{
					parser.MapNamePlaceholder: "",
				},
			},
			expected: []parser.Flat{{
				Name:        "field.<name>",
				Description: "field description",
				Default:     "",
			}},
		},
		{
			desc: "struct pointer field",
			element: &struct {
				Foo *struct {
					Field string `description:"field description"`
				} `description:"foo description"`
			}{
				Foo: &struct {
					Field string `description:"field description"`
				}{
					Field: "test",
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.field",
					Description: "field description",
					Default:     "test",
				},
			},
		},
		{
			desc: "struct pointer field, allow empty",
			element: &struct {
				Foo *struct {
					Field string `description:"field description"`
				} `description:"foo description" label:"allowEmpty"`
			}{
				Foo: &struct {
					Field string `description:"field description"`
				}{
					Field: "test",
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.field",
					Description: "field description",
					Default:     "test",
				},
			},
		},
		{
			desc: "struct pointer field level 2",
			element: &struct {
				Foo *struct {
					Fii *struct {
						Field string `description:"field description"`
					} `description:"fii description"`
				} `description:"foo description"`
			}{
				Foo: &struct {
					Fii *struct {
						Field string `description:"field description"`
					} `description:"fii description"`
				}{
					Fii: &struct {
						Field string `description:"field description"`
					}{
						Field: "test",
					},
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.fii.field",
					Description: "field description",
					Default:     "test",
				},
			},
		},
		{
			desc: "struct pointer field level 2, allow empty",
			element: &struct {
				Foo *struct {
					Fii *struct {
						Field string `description:"field description"`
					} `description:"fii description" label:"allowEmpty"`
				} `description:"foo description" label:"allowEmpty"`
			}{
				Foo: &struct {
					Fii *struct {
						Field string `description:"field description"`
					} `description:"fii description" label:"allowEmpty"`
				}{
					Fii: &struct {
						Field string `description:"field description"`
					}{
						Field: "test",
					},
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.fii",
					Description: "fii description",
					Default:     "false",
				},
				{
					Name:        "foo.fii.field",
					Description: "field description",
					Default:     "test",
				},
			},
		},
		{
			desc: "map string field level 2",
			element: &struct {
				Foo *struct {
					Fii map[string]string `description:"fii description"`
				} `description:"foo description"`
			}{
				Foo: &struct {
					Fii map[string]string `description:"fii description"`
				}{
					Fii: map[string]string{
						parser.MapNamePlaceholder: "",
					},
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.fii.<name>",
					Description: "fii description",
					Default:     "",
				},
			},
		},
		{
			desc: "map string pointer field level 2",
			element: &struct {
				Foo *struct {
					Fii map[string]*string `description:"fii description"`
				} `description:"foo description"`
			}{
				Foo: &struct {
					Fii map[string]*string `description:"fii description"`
				}{
					Fii: map[string]*string{
						parser.MapNamePlaceholder: func(v string) *string { return &v }(""),
					},
				},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.fii.<name>",
					Description: "fii description",
					Default:     "",
				},
			},
		},
		{
			desc: "map struct level 1",
			element: &struct {
				Foo map[string]struct {
					Field string `description:"field description"`
					Yo    int    `description:"yo description"`
				} `description:"foo description"`
			}{},
			expected: []parser.Flat{
				{
					Name:        "foo.<name>",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.<name>.field",
					Description: "field description",
					Default:     "",
				},
				{
					Name:        "foo.<name>.yo",
					Description: "yo description",
					Default:     "0",
				},
			},
		},
		{
			desc: "map struct pointer level 1",
			element: &struct {
				Foo map[string]*struct {
					Field string `description:"field description"`
					Yo    string `description:"yo description"`
				} `description:"foo description"`
			}{},
			expected: []parser.Flat{
				{
					Name:        "foo.<name>",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.<name>.field",
					Description: "field description",
					Default:     "",
				},
				{
					Name:        "foo.<name>.yo",
					Description: "yo description",
					Default:     "",
				},
			},
		},
		{
			desc: "time duration field",
			element: &struct {
				Field time.Duration `description:"field description"`
			}{
				Field: 1 * time.Second,
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "1s",
			}},
		},
		{
			desc: "time duration field map",
			element: &struct {
				Foo map[string]*struct {
					Field time.Duration `description:"field description"`
				} `description:"foo description"`
			}{
				Foo: map[string]*struct {
					Field time.Duration `description:"field description"`
				}{},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.<name>",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.<name>.field",
					Description: "field description",
					Default:     "0s",
				},
			},
		},
		{
			desc: "time duration field map 2",
			element: &struct {
				Foo map[string]*struct {
					Fii *struct {
						Field time.Duration `description:"field description"`
					}
				} `description:"foo description"`
			}{
				Foo: map[string]*struct {
					Fii *struct {
						Field time.Duration `description:"field description"`
					}
				}{},
			},
			expected: []parser.Flat{
				{
					Name:        "foo.<name>",
					Description: "foo description",
					Default:     "false",
				},
				{
					Name:        "foo.<name>.fii.field",
					Description: "field description",
					Default:     "0s",
				},
			},
		},
		{
			desc: "time duration field 2",
			element: &struct {
				Foo *struct {
					Field time.Duration `description:"field description"`
				}
			}{
				Foo: &struct {
					Field time.Duration `description:"field description"`
				}{
					Field: 1 * time.Second,
				},
			},
			expected: []parser.Flat{{
				Name:        "foo.field",
				Description: "field description",
				Default:     "1s",
			}},
		},
		{
			desc: "time duration field 3",
			element: &struct {
				Foo *struct {
					Fii *struct {
						Field time.Duration `description:"field description"`
					}
				}
			}{
				Foo: &struct {
					Fii *struct {
						Field time.Duration `description:"field description"`
					}
				}{
					Fii: &struct {
						Field time.Duration `description:"field description"`
					}{
						Field: 1 * time.Second,
					},
				},
			},
			expected: []parser.Flat{{
				Name:        "foo.fii.field",
				Description: "field description",
				Default:     "1s",
			}},
		},
		{
			desc: "time duration field",
			element: &struct {
				Field types.Duration `description:"field description"`
			}{
				Field: types.Duration(180 * time.Second),
			},
			expected: []parser.Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "180",
			}},
		},
		{
			desc: "slice of struct",
			element: &struct {
				Foo *struct {
					Fii []struct {
						Field1 string `description:"field1 description"`
						Field2 int    `description:"field2 description"`
					} `description:"fii description"`
				} `description:"foo description"`
			}{},
			expected: []parser.Flat{
				{
					Name:        "foo.fii",
					Description: "fii description",
					Default:     "",
				},
				{
					Name:        "foo.fii[0].field1",
					Description: "field1 description",
					Default:     "",
				},
				{
					Name:        "foo.fii[0].field2",
					Description: "field2 description",
					Default:     "0",
				},
			},
		},
		// Skipped: because realistically not needed in Traefik for now.
		// {
		// 	desc: "map of map field level 2",
		// 	element: &struct {
		// 		Foo *struct {
		// 			Fii map[string]map[string]string `description:"fii description"`
		// 		} `description:"foo description"`
		// 	}{
		// 		Foo: &struct {
		// 			Fii map[string]map[string]string `description:"fii description"`
		// 		}{
		// 			Fii: map[string]map[string]string{
		// 				parser.MapNamePlaceholder: {
		// 					parser.MapNamePlaceholder: "test",
		// 				},
		// 			},
		// 		},
		// 	},
		// 	expected: `XXX`,
		// },
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			generator.Generate(test.element)

			entries, err := Encode(test.element)
			require.NoError(t, err)

			assert.Equal(t, test.expected, entries)
		})
	}
}
