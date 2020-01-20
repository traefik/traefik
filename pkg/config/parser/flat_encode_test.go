package parser

import (
	"reflect"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeToFlat(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		node     *Node
		opts     *FlatOpts
		expected []Flat
	}{
		{
			desc: "string field",
			element: &struct {
				Field string `description:"field description"`
			}{
				Field: "test",
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						FieldName:   "Field",
						Description: "field description",
						Value:       "test",
						Kind:        reflect.String,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "6",
						Kind:        reflect.Int,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "true",
						Kind:        reflect.Bool,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "test",
						Kind:        reflect.Ptr,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
				Name:        "field",
				Description: "field description",
				Default:     "test",
			}},
		},
		{
			desc: "string pointer field, custom option",
			element: &struct {
				Field *string `description:"field description"`
			}{
				Field: func(v string) *string { return &v }("test"),
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "test",
						Kind:        reflect.Ptr,
						Tag:         `description:"field description"`,
					},
				},
			},
			opts: &FlatOpts{
				Case:      "upper",
				Separator: "_",
				SkipRoot:  false,
				TagName:   TagLabel,
			},
			expected: []Flat{{
				Name:        "TRAEFIK_FIELD",
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "6",
						Kind:        reflect.Ptr,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "true",
						Kind:        reflect.Ptr,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Kind:        reflect.Slice,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "foo, bar",
						Kind:        reflect.Slice,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Kind:        reflect.Slice,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "6, 3",
						Kind:        reflect.Slice,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
					MapNamePlaceholder: "",
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Kind:        reflect.Map,
						Tag:         `description:"field description"`,
						Children: []*Node{
							{
								Name:      "\u003cname\u003e",
								FieldName: "\u003cname\u003e",
								Kind:      reflect.String,
							},
						},
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:        "Field",
								Description: "field description",
								FieldName:   "Field",
								Value:       "test",
								Kind:        reflect.String,
								Tag:         `description:"field description"`,
							},
						},
					},
				},
			},
			expected: []Flat{
				{
					Name:        "foo.field",
					Description: "field description",
					Default:     "test",
				},
			},
		},
		{
			desc: "struct pointer field, hide field",
			element: &struct {
				Foo *struct {
					Field string `description:"-"`
				} `description:"foo description"`
			}{
				Foo: &struct {
					Field string `description:"-"`
				}{
					Field: "test",
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:        "Field",
								Description: "-",
								FieldName:   "Field",
								Value:       "test",
								Kind:        reflect.String,
								Tag:         `description:"-"`,
							},
						},
					},
				},
			},
			expected: nil,
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description" label:"allowEmpty"`,
						Children: []*Node{
							{
								Name:        "Field",
								Description: "field description",
								FieldName:   "Field",
								Value:       "test",
								Kind:        reflect.String,
								Tag:         `description:"field description"`,
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:        "Fii",
								Description: "fii description",
								FieldName:   "Fii",
								Kind:        reflect.Ptr,
								Tag:         `description:"fii description"`,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Value:       "test",
										Kind:        reflect.String,
										Tag:         `description:"field description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description" label:"allowEmpty"`,
						Children: []*Node{
							{
								Name:        "Fii",
								Description: "fii description",
								FieldName:   "Fii",
								Kind:        reflect.Ptr,
								Tag:         `description:"fii description" label:"allowEmpty"`,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Value:       "test",
										Kind:        reflect.String,
										Tag:         `description:"field description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
						MapNamePlaceholder: "",
					},
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:        "Fii",
								Description: "fii description",
								FieldName:   "Fii",
								Kind:        reflect.Map,
								Tag:         `description:"fii description"`,
								Children: []*Node{
									{
										Name:      "\u003cname\u003e",
										FieldName: "\u003cname\u003e",
										Kind:      reflect.String,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
						MapNamePlaceholder: func(v string) *string { return &v }(""),
					},
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Ptr,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:        "Fii",
								Description: "fii description",
								FieldName:   "Fii",
								Kind:        reflect.Map,
								Tag:         `description:"fii description"`,
								Children: []*Node{
									{
										Name:      "\u003cname\u003e",
										FieldName: "\u003cname\u003e",
										Kind:      reflect.Ptr,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Map,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:      "\u003cname\u003e",
								FieldName: "\u003cname\u003e",
								Kind:      reflect.Struct,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Kind:        reflect.String,
										Tag:         `description:"field description"`,
									},
									{
										Name:        "Yo",
										Description: "yo description",
										FieldName:   "Yo",
										Value:       "0",
										Kind:        reflect.Int,
										Tag:         `description:"yo description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Map,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:      "\u003cname\u003e",
								FieldName: "\u003cname\u003e",
								Kind:      reflect.Ptr,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Kind:        reflect.String,
										Tag:         `description:"field description"`,
									},
									{
										Name:        "Yo",
										Description: "yo description",
										FieldName:   "Yo",
										Kind:        reflect.String,
										Tag:         `description:"yo description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "1000000000",
						Kind:        reflect.Int64,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
				}{
					"<name>": {
						Field: 0,
					},
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Map,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:      "\u003cname\u003e",
								FieldName: "\u003cname\u003e",
								Kind:      reflect.Ptr,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Value:       "0",
										Kind:        reflect.Int64,
										Tag:         `description:"field description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
				}{
					"<name>": {
						Fii: &struct {
							Field time.Duration `description:"field description"`
						}{
							Field: 0,
						},
					},
				},
			},
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:        "Foo",
						Description: "foo description",
						FieldName:   "Foo",
						Kind:        reflect.Map,
						Tag:         `description:"foo description"`,
						Children: []*Node{
							{
								Name:      "\u003cname\u003e",
								FieldName: "\u003cname\u003e",
								Kind:      reflect.Ptr,
								Children: []*Node{
									{
										Name:      "Fii",
										FieldName: "Fii",
										Kind:      reflect.Ptr,
										Children: []*Node{
											{
												Name:        "Field",
												Description: "field description",
												FieldName:   "Field",
												Value:       "0",
												Kind:        reflect.Int64,
												Tag:         `description:"field description"`,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Ptr,
						Children: []*Node{
							{
								Name:        "Field",
								Description: "field description",
								FieldName:   "Field",
								Value:       "1000000000",
								Kind:        reflect.Int64,
								Tag:         `description:"field description"`,
							},
						},
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Ptr,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Ptr,
						Children: []*Node{
							{
								Name:      "Fii",
								FieldName: "Fii",
								Kind:      reflect.Ptr,
								Children: []*Node{
									{
										Name:        "Field",
										Description: "field description",
										FieldName:   "Field",
										Value:       "1000000000",
										Kind:        reflect.Int64,
										Tag:         `description:"field description"`,
									},
								},
							},
						},
					},
				},
			},
			expected: []Flat{{
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
			node: &Node{
				Name:      "traefik",
				FieldName: "",
				Kind:      reflect.Struct,
				Children: []*Node{
					{
						Name:        "Field",
						Description: "field description",
						FieldName:   "Field",
						Value:       "180000000000",
						Kind:        reflect.Int64,
						Tag:         `description:"field description"`,
					},
				},
			},
			expected: []Flat{{
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
			}{
				Foo: &struct {
					Fii []struct {
						Field1 string `description:"field1 description"`
						Field2 int    `description:"field2 description"`
					} `description:"fii description"`
				}{
					Fii: []struct {
						Field1 string `description:"field1 description"`
						Field2 int    `description:"field2 description"`
					}{
						{
							Field1: "",
							Field2: 0,
						},
					},
				},
			},
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", Kind: reflect.Ptr, Description: "foo description", Children: []*Node{
						{Name: "Fii", Kind: reflect.Slice, Description: "fii description", Children: []*Node{
							{Name: "[0]", Kind: reflect.Struct, Children: []*Node{
								{Name: "Field1", Value: "", Kind: reflect.String, Description: "field1 description"},
								{Name: "Field2", Value: "0", Kind: reflect.Int, Description: "field2 description"},
							}},
						}},
					}},
				},
			},
			expected: []Flat{
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
		// 				MapNamePlaceholder: {
		// 					MapNamePlaceholder: "test",
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

			var opts FlatOpts
			if test.opts == nil {
				opts = FlatOpts{Separator: ".", SkipRoot: true, TagName: TagLabel}
			} else {
				opts = *test.opts
			}

			entries, err := EncodeToFlat(test.element, test.node, opts)
			require.NoError(t, err)

			assert.Equal(t, test.expected, entries)
		})
	}
}
