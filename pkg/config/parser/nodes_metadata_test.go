package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddMetadata(t *testing.T) {
	type expected struct {
		node  *Node
		error bool
	}

	type interf interface{}

	testCases := []struct {
		desc      string
		tree      *Node
		structure interface{}
		expected  expected
	}{
		{
			desc:      "Node Nil",
			tree:      nil,
			structure: nil,
			expected:  expected{node: nil},
		},
		{
			desc:      "Empty Node",
			tree:      &Node{},
			structure: nil,
			expected:  expected{error: true},
		},
		{
			desc: "Nil structure",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "bar"},
				},
			},
			structure: nil,
			expected:  expected{error: true},
		},
		{
			desc:     "level 0",
			tree:     &Node{Name: "traefik", Value: "bar"},
			expected: expected{error: true},
		},
		{
			desc: "level 1",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar"},
				},
			},
			structure: struct{ Foo string }{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.String},
					},
				},
			},
		},
		{
			desc: "level 1, pointer",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "bar"},
				},
			},
			structure: &struct{ Foo string }{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Ptr,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.String},
					},
				},
			},
		},
		{
			desc: "level 1, slice",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "bar,bur"},
				},
			},
			structure: struct{ Foo []string }{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "bar,bur", Kind: reflect.Slice},
					},
				},
			},
		},
		{
			desc: "level 1, interface",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "", Children: []*Node{
						{Name: "Fii", Value: "hii"},
					}},
				},
			},
			structure: struct{ Foo interf }{},
			expected:  expected{error: true},
		},
		{
			desc: "level 1, map string",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "name1", Value: "bar"},
						{Name: "name2", Value: "bur"},
					}},
				},
			},
			structure: struct{ Foo map[string]string }{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Kind: reflect.Map, Children: []*Node{
							{Name: "name1", Value: "bar", Kind: reflect.String},
							{Name: "name2", Value: "bur", Kind: reflect.String},
						}},
					},
				},
			},
		},
		{
			desc: "level 1, map struct",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "name1", Children: []*Node{
							{Name: "Fii", Value: "bar"},
						}},
					}},
				},
			},
			structure: struct {
				Foo map[string]struct{ Fii string }
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Kind: reflect.Map, Children: []*Node{
							{Name: "name1", Kind: reflect.Struct, Children: []*Node{
								{Name: "Fii", FieldName: "Fii", Value: "bar", Kind: reflect.String},
							}},
						}},
					},
				},
			},
		},
		{
			desc: "level 1, map int as key",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "name1", Children: []*Node{
							{Name: "Fii", Value: "bar"},
						}},
					}},
				},
			},
			structure: struct {
				Foo map[int]struct{ Fii string }
			}{},
			expected: expected{error: true},
		},
		{
			desc: "level 1, int pointer",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "0"},
				},
			},
			structure: struct {
				Foo *int
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "0", Kind: reflect.Ptr},
					},
				},
			},
		},
		{
			desc: "level 1, bool pointer",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "0"},
				},
			},
			structure: struct {
				Foo *bool
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "0", Kind: reflect.Ptr},
					},
				},
			},
		},
		{
			desc: "level 1, string pointer",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "0"},
				},
			},
			structure: struct {
				Foo *string
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "0", Kind: reflect.Ptr},
					},
				},
			},
		},
		{
			desc: "level 1, 2 children with different types",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "bar"},
					{Name: "Fii", Value: "1"},
				},
			},
			structure: struct {
				Foo string
				Fii int
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.String},
						{Name: "Fii", FieldName: "Fii", Value: "1", Kind: reflect.Int},
					},
				},
			},
		},
		{
			desc: "level 1, use exported instead of unexported",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Value: "bar"},
				},
			},
			structure: struct {
				foo int
				Foo string
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "foo", Value: "bar", FieldName: "Foo", Kind: reflect.String},
					},
				},
			},
		},
		{
			desc: "level 1, unexported",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Value: "bar"},
				},
			},
			structure: struct {
				foo string
			}{},
			expected: expected{error: true},
		},
		{
			desc: "level 1, 3 children with different types",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "bar"},
					{Name: "Fii", Value: "1"},
					{Name: "Fuu", Value: "true"},
				},
			},
			structure: struct {
				Foo string
				Fii int
				Fuu bool
			}{},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.String},
						{Name: "Fii", FieldName: "Fii", Value: "1", Kind: reflect.Int},
						{Name: "Fuu", FieldName: "Fuu", Value: "true", Kind: reflect.Bool},
					},
				},
			},
		},
		{
			desc: "level 2",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Bar", Value: "bir"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				}
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Kind: reflect.Struct, Children: []*Node{
							{Name: "Bar", FieldName: "Bar", Value: "bir", Kind: reflect.String},
						}},
					},
				},
			},
		},
		{
			desc: "level 2, struct without children",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo"},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				}
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{error: true},
		},
		{
			desc: "level 2, slice-as-struct",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Fii", Children: []*Node{
						{Name: "bar", Value: "haa"},
						{Name: "bir", Value: "hii"},
					}},
				},
			},
			structure: struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{
				Foo: []struct {
					Bar string
					Bir string
				}{},
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Fii",
						FieldName: "Foo",
						Kind:      reflect.Slice,
						Tag:       reflect.StructTag(`label-slice-as-struct:"Fii"`),
						Children: []*Node{
							{Name: "bar", FieldName: "Bar", Kind: reflect.String, Value: "haa"},
							{Name: "bir", FieldName: "Bir", Kind: reflect.String, Value: "hii"},
						},
					},
				},
			}},
		},
		{
			desc: "level 2, slice-as-struct without children",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Fii"},
				},
			},
			structure: struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{
				Foo: []struct {
					Bar string
					Bir string
				}{},
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Fii",
						FieldName: "Foo",
						Kind:      reflect.Slice,
						Tag:       reflect.StructTag(`label-slice-as-struct:"Fii"`),
					},
				},
			}},
		},
		{
			desc: "level 2, struct with allowEmpty, value true",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "true"},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				} `label:"allowEmpty"`
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "true", Kind: reflect.Struct, Tag: reflect.StructTag(`label:"allowEmpty"`)},
					},
				},
			},
		},
		{
			desc: "level 2, struct with allowEmpty, value true with case variation",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "TruE"},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				} `label:"allowEmpty"`
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "TruE", Kind: reflect.Struct, Tag: reflect.StructTag(`label:"allowEmpty"`)},
					},
				},
			},
		},
		{
			desc: "level 2, struct with allowEmpty, value false",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "false"},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				} `label:"allowEmpty"`
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Value: "false", Disabled: true, Kind: reflect.Struct, Tag: reflect.StructTag(`label:"allowEmpty"`)},
					},
				},
			},
		},
		{
			desc: "level 2, struct with allowEmpty with children, value false",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Value: "false", Children: []*Node{
						{Name: "Bar", Value: "hii"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
				} `label:"allowEmpty"`
			}{
				Foo: struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{
							Name:      "Foo",
							FieldName: "Foo",
							Value:     "false",
							Disabled:  true,
							Kind:      reflect.Struct,
							Tag:       reflect.StructTag(`label:"allowEmpty"`),
							Children: []*Node{
								{Name: "Bar", FieldName: "Bar", Value: "hii", Kind: reflect.String},
							},
						},
					},
				},
			},
		},
		{
			desc: "level 2, struct pointer without children",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo"},
				},
			},
			structure: struct {
				Foo *struct {
					Bar string
				}
			}{
				Foo: &struct {
					Bar string
				}{},
			},
			expected: expected{error: true},
		},
		{
			desc: "level 2, map without children",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo"},
				},
			},
			structure: struct {
				Foo map[string]string
			}{
				Foo: map[string]string{},
			},
			expected: expected{error: true},
		},
		{
			desc: "level 2, pointer",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Bar", Value: "bir"},
					}},
				},
			},
			structure: struct {
				Foo *struct {
					Bar string
				}
			}{
				Foo: &struct {
					Bar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{Name: "Foo", FieldName: "Foo", Kind: reflect.Ptr, Children: []*Node{
							{Name: "Bar", FieldName: "Bar", Value: "bir", Kind: reflect.String},
						}},
					},
				},
			},
		},
		{
			desc: "level 2, 2 children",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Bar", Value: "bir"},
						{Name: "Bur", Value: "fuu"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					Bar string
					Bur string
				}
			}{
				Foo: struct {
					Bar string
					Bur string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{
							Name:      "Foo",
							FieldName: "Foo",
							Kind:      reflect.Struct,
							Children: []*Node{
								{Name: "Bar", FieldName: "Bar", Value: "bir", Kind: reflect.String},
								{Name: "Bur", FieldName: "Bur", Value: "fuu", Kind: reflect.String},
							}},
					},
				},
			},
		},
		{
			desc: "level 3",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Bar", Children: []*Node{
							{Name: "Bur", Value: "fuu"},
						}},
					}},
				},
			},
			structure: struct {
				Foo struct {
					Bar struct {
						Bur string
					}
				}
			}{
				Foo: struct {
					Bar struct {
						Bur string
					}
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{
							Name:      "Foo",
							FieldName: "Foo",
							Kind:      reflect.Struct,
							Children: []*Node{
								{
									Name:      "Bar",
									FieldName: "Bar",
									Kind:      reflect.Struct,
									Children: []*Node{
										{Name: "Bur", FieldName: "Bur", Value: "fuu", Kind: reflect.String},
									}},
							}},
					},
				},
			},
		},
		{
			desc: "level 3, 2 children level 1, 2 children level 2, 2 children level 3",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Bar", Children: []*Node{
							{Name: "Fii", Value: "fii"},
							{Name: "Fee", Value: "1"},
						}},
						{Name: "Bur", Children: []*Node{
							{Name: "Faa", Value: "faa"},
						}},
					}},
					{Name: "Fii", Children: []*Node{
						{Name: "FiiBar", Value: "fiiBar"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					Bar struct {
						Fii string
						Fee int
					}
					Bur struct {
						Faa string
					}
				}
				Fii struct {
					FiiBar string
				}
			}{
				Foo: struct {
					Bar struct {
						Fii string
						Fee int
					}
					Bur struct {
						Faa string
					}
				}{},
				Fii: struct {
					FiiBar string
				}{},
			},
			expected: expected{
				node: &Node{
					Name: "traefik",
					Kind: reflect.Struct,
					Children: []*Node{
						{
							Name:      "Foo",
							FieldName: "Foo",
							Kind:      reflect.Struct,
							Children: []*Node{
								{
									Name:      "Bar",
									FieldName: "Bar",
									Kind:      reflect.Struct,
									Children: []*Node{
										{Name: "Fii", FieldName: "Fii", Kind: reflect.String, Value: "fii"},
										{Name: "Fee", FieldName: "Fee", Kind: reflect.Int, Value: "1"},
									}},
								{
									Name:      "Bur",
									FieldName: "Bur",
									Kind:      reflect.Struct,
									Children: []*Node{
										{Name: "Faa", FieldName: "Faa", Kind: reflect.String, Value: "faa"},
									}},
							}},
						{
							Name:      "Fii",
							FieldName: "Fii",
							Kind:      reflect.Struct,
							Children: []*Node{
								{Name: "FiiBar", FieldName: "FiiBar", Kind: reflect.String, Value: "fiiBar"},
							}},
					},
				},
			},
		},
		{
			desc: "Slice struct",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "[0]", Children: []*Node{
							{Name: "Field1", Value: "A"},
							{Name: "Field2", Value: "A"},
						}},
						{Name: "[1]", Children: []*Node{
							{Name: "Field1", Value: "B"},
							{Name: "Field2", Value: "B"},
						}},
						{Name: "[2]", Children: []*Node{
							{Name: "Field1", Value: "C"},
							{Name: "Field2", Value: "C"},
						}},
					}},
				},
			},
			structure: struct {
				Foo []struct {
					Field1 string
					Field2 string
				}
			}{},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Kind: reflect.Slice, Children: []*Node{
						{Name: "[0]", Kind: reflect.Struct, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "A", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "A", Kind: reflect.String},
						}},
						{Name: "[1]", Kind: reflect.Struct, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "B", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "B", Kind: reflect.String},
						}},
						{Name: "[2]", Kind: reflect.Struct, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "C", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "C", Kind: reflect.String},
						}},
					}},
				},
			}},
		},
		{
			desc: "Slice pointer struct",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "[0]", Children: []*Node{
							{Name: "Field1", Value: "A"},
							{Name: "Field2", Value: "A"},
						}},
						{Name: "[1]", Children: []*Node{
							{Name: "Field1", Value: "B"},
							{Name: "Field2", Value: "B"},
						}},
						{Name: "[2]", Children: []*Node{
							{Name: "Field1", Value: "C"},
							{Name: "Field2", Value: "C"},
						}},
					}},
				},
			},
			structure: struct {
				Foo []*struct {
					Field1 string
					Field2 string
				}
			}{},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Kind: reflect.Slice, Children: []*Node{
						{Name: "[0]", Kind: reflect.Ptr, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "A", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "A", Kind: reflect.String},
						}},
						{Name: "[1]", Kind: reflect.Ptr, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "B", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "B", Kind: reflect.String},
						}},
						{Name: "[2]", Kind: reflect.Ptr, Children: []*Node{
							{Name: "Field1", FieldName: "Field1", Value: "C", Kind: reflect.String},
							{Name: "Field2", FieldName: "Field2", Value: "C", Kind: reflect.String},
						}},
					}},
				},
			}},
		},
		{
			desc: "embedded",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "Fii", Value: "bir"},
						{Name: "Fuu", Value: "bur"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					FiiFoo
				}
			}{},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Kind: reflect.Struct, Children: []*Node{
						{Name: "Fii", FieldName: "Fii", Value: "bir", Kind: reflect.String},
						{Name: "Fuu", FieldName: "Fuu", Value: "bur", Kind: reflect.String},
					}},
				},
			}},
		},
		{
			desc: "embedded slice",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "MySliceType", Value: "foo,fii"},
				},
			},
			structure: struct {
				MySliceType
			}{},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "MySliceType", FieldName: "MySliceType", Value: "foo,fii", Kind: reflect.Slice},
				},
			}},
		},
		{
			desc: "embedded slice 2",
			tree: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "MySliceType", Value: "foo,fii"},
					}},
				},
			},
			structure: struct {
				Foo struct {
					MySliceType
				}
			}{},
			expected: expected{node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Kind: reflect.Struct, Children: []*Node{
						{Name: "MySliceType", FieldName: "MySliceType", Value: "foo,fii", Kind: reflect.Slice},
					}},
				},
			}},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := metadata{MetadataOpts{TagName: TagLabel, AllowSliceAsStruct: true}}.Add(test.structure, test.tree)

			if test.expected.error {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if !assert.Equal(t, test.expected.node, test.tree) {
					bytes, errM := json.MarshalIndent(test.tree, "", "  ")
					require.NoError(t, errM)
					fmt.Println(string(bytes))
				}
			}
		})
	}
}

type MySliceType []string
