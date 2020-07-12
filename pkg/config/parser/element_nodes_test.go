package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeToNode(t *testing.T) {
	type expected struct {
		node  *Node
		error bool
	}

	testCases := []struct {
		desc     string
		element  interface{}
		expected expected
	}{
		{
			desc: "Description",
			element: struct {
				Foo string `description:"text"`
			}{Foo: "bar"},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar", Description: "text"},
				}},
			},
		},
		{
			desc: "string",
			element: struct {
				Foo string
			}{Foo: "bar"},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar"},
				}},
			},
		},
		{
			desc: "2 string fields",
			element: struct {
				Foo string
				Fii string
			}{Foo: "bar", Fii: "hii"},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar"},
					{Name: "Fii", FieldName: "Fii", Value: "hii"},
				}},
			},
		},
		{
			desc: "int",
			element: struct {
				Foo int
			}{Foo: 1},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "1"},
				}},
			},
		},
		{
			desc: "int8",
			element: struct {
				Foo int8
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "int16",
			element: struct {
				Foo int16
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "int32",
			element: struct {
				Foo int32
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "int64",
			element: struct {
				Foo int64
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "uint",
			element: struct {
				Foo uint
			}{Foo: 1},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "1"},
				}},
			},
		},
		{
			desc: "uint8",
			element: struct {
				Foo uint8
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "uint16",
			element: struct {
				Foo uint16
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "uint32",
			element: struct {
				Foo uint32
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "uint64",
			element: struct {
				Foo uint64
			}{Foo: 2},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2"},
				}},
			},
		},
		{
			desc: "float32",
			element: struct {
				Foo float32
			}{Foo: 1.12},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "1.120000"},
				}},
			},
		},
		{
			desc: "float64",
			element: struct {
				Foo float64
			}{Foo: 1.12},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "1.120000"},
				}},
			},
		},
		{
			desc: "bool",
			element: struct {
				Foo bool
			}{Foo: true},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "true"},
				}},
			},
		},
		{
			desc: "struct",
			element: struct {
				Foo struct {
					Fii string
					Fuu string
				}
			}{
				Foo: struct {
					Fii string
					Fuu string
				}{
					Fii: "hii",
					Fuu: "huu",
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "hii"},
						{Name: "Fuu", FieldName: "Fuu", Value: "huu"},
					}},
				}},
			},
		},
		{
			desc: "struct unexported field",
			element: struct {
				Foo struct {
					Fii string
					fuu string
				}
			}{
				Foo: struct {
					Fii string
					fuu string
				}{
					Fii: "hii",
					fuu: "huu",
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "hii"},
					}},
				}},
			},
		},
		{
			desc: "struct pointer",
			element: struct {
				Foo *struct {
					Fii string
					Fuu string
				}
			}{
				Foo: &struct {
					Fii string
					Fuu string
				}{
					Fii: "hii",
					Fuu: "huu",
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "hii"},
						{Name: "Fuu", FieldName: "Fuu", Value: "huu"},
					}},
				}},
			},
		},
		{
			desc: "string pointer",
			element: struct {
				Foo *struct {
					Fii *string
					Fuu string
				}
			}{
				Foo: &struct {
					Fii *string
					Fuu string
				}{
					Fii: func(v string) *string { return &v }("hii"),
					Fuu: "huu",
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "hii"},
						{Name: "Fuu", FieldName: "Fuu", Value: "huu"},
					}},
				}},
			},
		},
		{
			desc: "string nil pointer",
			element: struct {
				Foo *struct {
					Fii *string
					Fuu string
				}
			}{
				Foo: &struct {
					Fii *string
					Fuu string
				}{
					Fii: nil,
					Fuu: "huu",
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fuu", FieldName: "Fuu", Value: "huu"},
					}},
				}},
			},
		},
		{
			desc: "int pointer",
			element: struct {
				Foo *struct {
					Fii *int
					Fuu int
				}
			}{
				Foo: &struct {
					Fii *int
					Fuu int
				}{
					Fii: func(v int) *int { return &v }(6),
					Fuu: 4,
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "6"},
						{Name: "Fuu", FieldName: "Fuu", Value: "4"},
					}},
				}},
			},
		},
		{
			desc: "bool pointer",
			element: struct {
				Foo *struct {
					Fii *bool
					Fuu bool
				}
			}{
				Foo: &struct {
					Fii *bool
					Fuu bool
				}{
					Fii: func(v bool) *bool { return &v }(true),
					Fuu: true,
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "true"},
						{Name: "Fuu", FieldName: "Fuu", Value: "true"},
					}},
				}},
			},
		},
		{
			desc: "struct nil struct pointer",
			element: struct {
				Foo *struct {
					Fii *string
					Fuu string
				}
			}{
				Foo: nil,
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "struct pointer, not allowEmpty",
			element: struct {
				Foo *struct {
					Fii string
					Fuu string
				}
			}{
				Foo: &struct {
					Fii string
					Fuu string
				}{},
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "struct pointer, allowEmpty",
			element: struct {
				Foo *struct {
					Fii string
					Fuu string
				} `label:"allowEmpty"`
			}{
				Foo: &struct {
					Fii string
					Fuu string
				}{},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "true"},
				}},
			},
		},
		{
			desc: "map",
			element: struct {
				Foo struct {
					Bar map[string]string
				}
			}{
				Foo: struct {
					Bar map[string]string
				}{
					Bar: map[string]string{
						"name1": "huu",
					},
				},
			},
			expected: expected{node: &Node{Name: "traefik", Children: []*Node{
				{Name: "Foo", FieldName: "Foo", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Children: []*Node{
						{Name: "name1", FieldName: "name1", Value: "huu"},
					}},
				}},
			}}},
		},
		{
			desc: "empty map",
			element: struct {
				Bar map[string]string
			}{
				Bar: map[string]string{},
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "map nil",
			element: struct {
				Bar map[string]string
			}{
				Bar: nil,
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "map with non string key",
			element: struct {
				Foo struct {
					Bar map[int]string
				}
			}{
				Foo: struct {
					Bar map[int]string
				}{
					Bar: map[int]string{
						1: "huu",
					},
				},
			},
			expected: expected{error: true},
		},
		{
			desc:    "slice of string",
			element: struct{ Bar []string }{Bar: []string{"huu", "hii"}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "huu, hii"},
				}},
			},
		},
		{
			desc:    "slice of int",
			element: struct{ Bar []int }{Bar: []int{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of int8",
			element: struct{ Bar []int8 }{Bar: []int8{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of int16",
			element: struct{ Bar []int16 }{Bar: []int16{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of int32",
			element: struct{ Bar []int32 }{Bar: []int32{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of int64",
			element: struct{ Bar []int64 }{Bar: []int64{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of uint",
			element: struct{ Bar []uint }{Bar: []uint{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of uint8",
			element: struct{ Bar []uint8 }{Bar: []uint8{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of uint16",
			element: struct{ Bar []uint16 }{Bar: []uint16{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of uint32",
			element: struct{ Bar []uint32 }{Bar: []uint32{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of uint64",
			element: struct{ Bar []uint64 }{Bar: []uint64{4, 2, 3}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4, 2, 3"},
				}},
			},
		},
		{
			desc:    "slice of float32",
			element: struct{ Bar []float32 }{Bar: []float32{4.1, 2, 3.2}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4.100000, 2.000000, 3.200000"},
				}},
			},
		},
		{
			desc:    "slice of float64",
			element: struct{ Bar []float64 }{Bar: []float64{4.1, 2, 3.2}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "4.100000, 2.000000, 3.200000"},
				}},
			},
		},
		{
			desc:    "slice of bool",
			element: struct{ Bar []bool }{Bar: []bool{true, false, true}},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Bar", FieldName: "Bar", Value: "true, false, true"},
				}},
			},
		},
		{
			desc: "slice label-slice-as-struct",
			element: &struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{
				Foo: []struct {
					Bar string
					Bir string
				}{
					{
						Bar: "haa",
						Bir: "hii",
					},
				},
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{{
					Name:      "Fii",
					FieldName: "Foo",
					Children: []*Node{
						{Name: "Bar", FieldName: "Bar", Value: "haa"},
						{Name: "Bir", FieldName: "Bir", Value: "hii"},
					},
				}},
			}},
		},
		{
			desc: "slice label-slice-as-struct several slice entries",
			element: &struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{
				Foo: []struct {
					Bar string
					Bir string
				}{
					{
						Bar: "haa",
						Bir: "hii",
					},
					{
						Bar: "haa",
						Bir: "hii",
					},
				},
			},
			expected: expected{error: true},
		},
		{
			desc: "slice of struct",
			element: struct {
				Foo []struct {
					Field string
				}
			}{
				Foo: []struct {
					Field string
				}{
					{
						Field: "bar",
					},
					{
						Field: "bir",
					},
				},
			},
			expected: expected{node: &Node{Name: "traefik", Children: []*Node{
				{Name: "Foo", FieldName: "Foo", Children: []*Node{
					{Name: "[0]", Children: []*Node{
						{Name: "Field", FieldName: "Field", Value: "bar"},
					}},
					{Name: "[1]", Children: []*Node{
						{Name: "Field", FieldName: "Field", Value: "bir"},
					}},
				}},
			}}},
		},
		{
			desc: "slice of pointer of struct",
			element: struct {
				Foo []*struct {
					Field string
				}
			}{
				Foo: []*struct {
					Field string
				}{
					{Field: "bar"},
					{Field: "bir"},
				},
			},
			expected: expected{node: &Node{Name: "traefik", Children: []*Node{
				{Name: "Foo", FieldName: "Foo", Children: []*Node{
					{Name: "[0]", Children: []*Node{
						{Name: "Field", FieldName: "Field", Value: "bar"},
					}},
					{Name: "[1]", Children: []*Node{
						{Name: "Field", FieldName: "Field", Value: "bir"},
					}},
				}},
			}}},
		},
		{
			desc: "empty slice",
			element: struct {
				Bar []string
			}{
				Bar: []string{},
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "nil slice",
			element: struct {
				Bar []string
			}{
				Bar: nil,
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "ignore slice",
			element: struct {
				Bar []string `label:"-"`
			}{
				Bar: []string{"huu", "hii"},
			},
			expected: expected{node: &Node{Name: "traefik"}},
		},
		{
			desc: "embedded",
			element: struct {
				Foo struct{ FiiFoo }
			}{
				Foo: struct{ FiiFoo }{
					FiiFoo: FiiFoo{
						Fii: "hii",
						Fuu: "huu",
					},
				},
			},
			expected: expected{
				node: &Node{Name: "traefik", Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "hii"},
						{Name: "Fuu", FieldName: "Fuu", Value: "huu"},
					}},
				}},
			},
		},
		{
			desc: "raw value",
			element: struct {
				Foo *struct {
					Bar map[string]interface{}
				}
			}{
				Foo: &struct {
					Bar map[string]interface{}
				}{
					Bar: map[string]interface{}{
						"AAA": "valueA",
						"BBB": map[string]interface{}{
							"CCC": map[string]interface{}{
								"DDD": "valueD",
							},
						},
					},
				},
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Children: []*Node{
						{Name: "Bar", FieldName: "Bar", RawValue: map[string]interface{}{
							"AAA": "valueA",
							"BBB": map[string]interface{}{
								"CCC": map[string]interface{}{
									"DDD": "valueD",
								},
							},
						}},
					}},
				},
			}},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			etnOpts := EncoderToNodeOpts{OmitEmpty: true, TagName: TagLabel, AllowSliceAsStruct: true}
			node, err := EncodeToNode(test.element, DefaultRootName, etnOpts)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				assert.Equal(t, test.expected.node, node)
			}
		})
	}
}
