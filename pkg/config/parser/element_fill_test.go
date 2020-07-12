package parser

import (
	"reflect"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFill(t *testing.T) {
	type expected struct {
		element interface{}
		error   bool
	}

	testCases := []struct {
		desc     string
		node     *Node
		element  interface{}
		expected expected
	}{
		{
			desc:     "empty node",
			node:     &Node{},
			element:  &struct{ Foo string }{},
			expected: expected{error: true},
		},
		{
			desc:     "empty element",
			node:     &Node{Name: "traefik", Kind: reflect.Struct},
			element:  &struct{}{},
			expected: expected{element: &struct{}{}},
		},
		{
			desc:     "type struct as root",
			node:     &Node{Name: "traefik", Kind: reflect.Struct},
			element:  struct{}{},
			expected: expected{error: true},
		},
		{
			desc:     "nil node",
			node:     nil,
			element:  &struct{ Foo string }{},
			expected: expected{element: &struct{ Foo string }{}},
		},
		{
			desc:     "nil element",
			node:     &Node{Name: "traefik", Kind: reflect.Struct},
			element:  nil,
			expected: expected{element: nil},
		},
		{
			desc: "string",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.String},
				},
			},
			element:  &struct{ Foo string }{},
			expected: expected{element: &struct{ Foo string }{Foo: "bar"}},
		},
		{
			desc: "field not found",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Fii", Value: "bar", Kind: reflect.String},
				},
			},
			element:  &struct{ Foo string }{},
			expected: expected{error: true},
		},
		{
			desc: "2 children",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Fii", FieldName: "Fii", Value: "bir", Kind: reflect.String},
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int},
				},
			},
			element: &struct {
				Fii string
				Foo int
			}{},
			expected: expected{element: &struct {
				Fii string
				Foo int
			}{Fii: "bir", Foo: 4}},
		},
		{
			desc: "case insensitive",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "foo", FieldName: "Foo", Value: "bir", Kind: reflect.String},
				},
			},
			element: &struct {
				Foo string
				foo int
			}{},
			expected: expected{element: &struct {
				Foo string
				foo int
			}{Foo: "bir"}},
		},
		{
			desc: "func",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Kind: reflect.Func},
				},
			},
			element:  &struct{ Foo func() }{},
			expected: expected{element: &struct{ Foo func() }{}},
		},
		{
			desc: "int",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int},
				},
			},
			element:  &struct{ Foo int }{},
			expected: expected{element: &struct{ Foo int }{Foo: 4}},
		},
		{
			desc: "invalid int",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Int},
				},
			},
			element:  &struct{ Foo int }{},
			expected: expected{error: true},
		},
		{
			desc: "int8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int8},
				},
			},
			element:  &struct{ Foo int8 }{},
			expected: expected{element: &struct{ Foo int8 }{Foo: 4}},
		},
		{
			desc: "invalid int8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Int8},
				},
			},
			element:  &struct{ Foo int8 }{},
			expected: expected{error: true},
		},
		{
			desc: "int16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int16},
				},
			},
			element:  &struct{ Foo int16 }{},
			expected: expected{element: &struct{ Foo int16 }{Foo: 4}},
		},
		{
			desc: "invalid int16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Int16},
				},
			},
			element:  &struct{ Foo int16 }{},
			expected: expected{error: true},
		},
		{
			desc: "int32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int32},
				},
			},
			element:  &struct{ Foo int32 }{},
			expected: expected{element: &struct{ Foo int32 }{Foo: 4}},
		},
		{
			desc: "invalid int32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Int32},
				},
			},
			element:  &struct{ Foo int32 }{},
			expected: expected{error: true},
		},
		{
			desc: "int64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo int64 }{},
			expected: expected{element: &struct{ Foo int64 }{Foo: 4}},
		},
		{
			desc: "invalid int64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo int64 }{},
			expected: expected{error: true},
		},
		{
			desc: "uint",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Uint},
				},
			},
			element:  &struct{ Foo uint }{},
			expected: expected{element: &struct{ Foo uint }{Foo: 4}},
		},
		{
			desc: "invalid uint",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Uint},
				},
			},
			element:  &struct{ Foo uint }{},
			expected: expected{error: true},
		},
		{
			desc: "uint8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Uint8},
				},
			},
			element:  &struct{ Foo uint8 }{},
			expected: expected{element: &struct{ Foo uint8 }{Foo: 4}},
		},
		{
			desc: "invalid uint8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Uint8},
				},
			},
			element:  &struct{ Foo uint8 }{},
			expected: expected{error: true},
		},
		{
			desc: "uint16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Uint16},
				},
			},
			element:  &struct{ Foo uint16 }{},
			expected: expected{element: &struct{ Foo uint16 }{Foo: 4}},
		},
		{
			desc: "invalid uint16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Uint16},
				},
			},
			element:  &struct{ Foo uint16 }{},
			expected: expected{error: true},
		},
		{
			desc: "uint32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Uint32},
				},
			},
			element:  &struct{ Foo uint32 }{},
			expected: expected{element: &struct{ Foo uint32 }{Foo: 4}},
		},
		{
			desc: "invalid uint32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Uint32},
				},
			},
			element:  &struct{ Foo uint32 }{},
			expected: expected{error: true},
		},
		{
			desc: "uint64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Uint64},
				},
			},
			element:  &struct{ Foo uint64 }{},
			expected: expected{element: &struct{ Foo uint64 }{Foo: 4}},
		},
		{
			desc: "invalid uint64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four", Kind: reflect.Uint64},
				},
			},
			element:  &struct{ Foo uint64 }{},
			expected: expected{error: true},
		},
		{
			desc: "time.Duration with unit",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4s", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo time.Duration }{},
			expected: expected{element: &struct{ Foo time.Duration }{Foo: 4 * time.Second}},
		},
		{
			desc: "time.Duration without unit",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo time.Duration }{},
			expected: expected{element: &struct{ Foo time.Duration }{Foo: 4 * time.Nanosecond}},
		},
		{
			desc: "types.Duration with unit",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4s", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo types.Duration }{},
			expected: expected{element: &struct{ Foo types.Duration }{Foo: types.Duration(4 * time.Second)}},
		},
		{
			desc: "types.Duration without unit",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Int64},
				},
			},
			element:  &struct{ Foo types.Duration }{},
			expected: expected{element: &struct{ Foo types.Duration }{Foo: types.Duration(4 * time.Second)}},
		},
		{
			desc: "bool",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "true", Kind: reflect.Bool},
				},
			},
			element:  &struct{ Foo bool }{},
			expected: expected{element: &struct{ Foo bool }{Foo: true}},
		},
		{
			desc: "invalid bool",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bool", Kind: reflect.Bool},
				},
			},
			element:  &struct{ Foo bool }{},
			expected: expected{error: true},
		},
		{
			desc: "float32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2.1", Kind: reflect.Float32},
				},
			},
			element:  &struct{ Foo float32 }{},
			expected: expected{element: &struct{ Foo float32 }{Foo: 2.1}},
		},
		{
			desc: "invalid float32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "two dot one", Kind: reflect.Float32},
				},
			},
			element:  &struct{ Foo float32 }{},
			expected: expected{error: true},
		},
		{
			desc: "float64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "2.1", Kind: reflect.Float64},
				},
			},
			element:  &struct{ Foo float64 }{},
			expected: expected{element: &struct{ Foo float64 }{Foo: 2.1}},
		},
		{
			desc: "invalid float64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "two dot one", Kind: reflect.Float64},
				},
			},
			element:  &struct{ Foo float64 }{},
			expected: expected{error: true},
		},
		{
			desc: "struct",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Struct,
						Children: []*Node{
							{Name: "Fii", FieldName: "Fii", Value: "huu", Kind: reflect.String},
							{Name: "Fuu", FieldName: "Fuu", Value: "6", Kind: reflect.Int},
						},
					},
				},
			},
			element: &struct {
				Foo struct {
					Fii string
					Fuu int
				}
			}{},
			expected: expected{
				element: &struct {
					Foo struct {
						Fii string
						Fuu int
					}
				}{
					Foo: struct {
						Fii string
						Fuu int
					}{
						Fii: "huu",
						Fuu: 6,
					},
				},
			},
		},
		{
			desc: "pointer",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Struct,
						Children: []*Node{
							{Name: "Fii", FieldName: "Fii", Value: "huu", Kind: reflect.String},
							{Name: "Fuu", FieldName: "Fuu", Value: "6", Kind: reflect.Int},
						},
					},
				},
			},
			element: &struct {
				Foo *struct {
					Fii string
					Fuu int
				}
			}{},
			expected: expected{
				element: &struct {
					Foo *struct {
						Fii string
						Fuu int
					}
				}{
					Foo: &struct {
						Fii string
						Fuu int
					}{
						Fii: "huu",
						Fuu: 6,
					},
				},
			},
		},
		{
			desc: "pointer disabled false without children",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Ptr,
					},
				},
			},
			element: &struct {
				Foo *struct {
					Fii string
					Fuu int
				} `label:"allowEmpty"`
			}{},
			expected: expected{
				element: &struct {
					Foo *struct {
						Fii string
						Fuu int
					} `label:"allowEmpty"`
				}{
					Foo: &struct {
						Fii string
						Fuu int
					}{},
				},
			},
		},
		{
			desc: "pointer disabled true without children",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Ptr,
						Disabled:  true,
					},
				},
			},
			element: &struct {
				Foo *struct {
					Fii string
					Fuu int
				} `label:"allowEmpty"`
			}{},
			expected: expected{
				element: &struct {
					Foo *struct {
						Fii string
						Fuu int
					} `label:"allowEmpty"`
				}{},
			},
		},
		{
			desc: "pointer disabled true with children",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Disabled:  true,
						Kind:      reflect.Ptr,
						Children: []*Node{
							{Name: "Fii", FieldName: "Fii", Value: "huu", Kind: reflect.String},
							{Name: "Fuu", FieldName: "Fuu", Value: "6", Kind: reflect.Int},
						},
					},
				},
			},
			element: &struct {
				Foo *struct {
					Fii string
					Fuu int
				} `label:"allowEmpty"`
			}{},
			expected: expected{
				element: &struct {
					Foo *struct {
						Fii string
						Fuu int
					} `label:"allowEmpty"`
				}{},
			},
		},
		{
			desc: "map string",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Map,
						Children: []*Node{
							{Name: "name1", Value: "hii", Kind: reflect.String},
							{Name: "name2", Value: "huu", Kind: reflect.String},
						},
					},
				},
			},
			element: &struct {
				Foo map[string]string
			}{},
			expected: expected{
				element: &struct {
					Foo map[string]string
				}{
					Foo: map[string]string{
						"name1": "hii",
						"name2": "huu",
					},
				},
			},
		},
		{
			desc: "map struct",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Map,
						Children: []*Node{
							{
								Name: "name1",
								Kind: reflect.Struct,
								Children: []*Node{
									{Name: "Fii", FieldName: "Fii", Kind: reflect.String, Value: "hii"},
								},
							},
							{
								Name: "name2",
								Kind: reflect.Struct,
								Children: []*Node{
									{Name: "Fii", FieldName: "Fii", Kind: reflect.String, Value: "huu"},
								},
							},
						},
					},
				},
			},
			element: &struct {
				Foo map[string]struct{ Fii string }
			}{},
			expected: expected{
				element: &struct {
					Foo map[string]struct{ Fii string }
				}{
					Foo: map[string]struct{ Fii string }{
						"name1": {Fii: "hii"},
						"name2": {Fii: "huu"},
					},
				},
			},
		},
		{
			desc: "slice string",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "huu,hii,hoo", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []string }{},
			expected: expected{element: &struct{ Foo []string }{Foo: []string{"huu", "hii", "hoo"}}},
		},
		{
			desc: "slice named type",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "huu,hii,hoo", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []NamedType }{},
			expected: expected{element: &struct{ Foo []NamedType }{Foo: []NamedType{"huu", "hii", "hoo"}}},
		},
		{
			desc: "slice named type int",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "1,2,3", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []NamedTypeInt }{},
			expected: expected{element: &struct{ Foo []NamedTypeInt }{Foo: []NamedTypeInt{1, 2, 3}}},
		},
		{
			desc: "empty slice",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []string }{},
			expected: expected{element: &struct{ Foo []string }{Foo: nil}},
		},
		{
			desc: "slice int",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int }{},
			expected: expected{element: &struct{ Foo []int }{Foo: []int{4, 3, 6}}},
		},
		{
			desc: "slice invalid int",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int }{},
			expected: expected{error: true},
		},
		{
			desc: "slice int8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Slice,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int8 }{},
			expected: expected{element: &struct{ Foo []int8 }{Foo: []int8{4, 3, 6}}},
		},
		{
			desc: "slice invalid int8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int8 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice int16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int16 }{},
			expected: expected{element: &struct{ Foo []int16 }{Foo: []int16{4, 3, 6}}},
		},
		{
			desc: "slice invalid int16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int16 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice int32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int32 }{},
			expected: expected{element: &struct{ Foo []int32 }{Foo: []int32{4, 3, 6}}},
		},
		{
			desc: "slice invalid int32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int32 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice int64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int64 }{},
			expected: expected{element: &struct{ Foo []int64 }{Foo: []int64{4, 3, 6}}},
		},
		{
			desc: "slice invalid int64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []int64 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice uint",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint }{},
			expected: expected{element: &struct{ Foo []uint }{Foo: []uint{4, 3, 6}}},
		},
		{
			desc: "slice invalid uint",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint }{},
			expected: expected{error: true},
		},
		{
			desc: "slice uint8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint8 }{},
			expected: expected{element: &struct{ Foo []uint8 }{Foo: []uint8{4, 3, 6}}},
		},
		{
			desc: "slice invalid uint8",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint8 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice uint16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint16 }{},
			expected: expected{element: &struct{ Foo []uint16 }{Foo: []uint16{4, 3, 6}}},
		},
		{
			desc: "slice invalid uint16",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint16 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice uint32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint32 }{},
			expected: expected{element: &struct{ Foo []uint32 }{Foo: []uint32{4, 3, 6}}},
		},
		{
			desc: "slice invalid uint32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint32 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice uint64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint64 }{},
			expected: expected{element: &struct{ Foo []uint64 }{Foo: []uint64{4, 3, 6}}},
		},
		{
			desc: "slice invalid uint64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []uint64 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice float32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []float32 }{},
			expected: expected{element: &struct{ Foo []float32 }{Foo: []float32{4, 3, 6}}},
		},
		{
			desc: "slice invalid float32",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []float32 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice float64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4,3,6", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []float64 }{},
			expected: expected{element: &struct{ Foo []float64 }{Foo: []float64{4, 3, 6}}},
		},
		{
			desc: "slice invalid float64",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "four,three,six", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []float64 }{},
			expected: expected{error: true},
		},
		{
			desc: "slice bool",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "true, false, true", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []bool }{},
			expected: expected{element: &struct{ Foo []bool }{Foo: []bool{true, false, true}}},
		},
		{
			desc: "slice invalid bool",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bool, false, true", Kind: reflect.Slice},
				},
			},
			element:  &struct{ Foo []bool }{},
			expected: expected{error: true},
		},
		{
			desc: "slice slice-as-struct",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Fii",
						FieldName: "Foo",
						Kind:      reflect.Slice,
						Tag:       `label-slice-as-struct:"Fii"`,
						Children: []*Node{
							{Name: "bar", FieldName: "Bar", Kind: reflect.String, Value: "haa"},
							{Name: "bir", FieldName: "Bir", Kind: reflect.String, Value: "hii"},
						},
					},
				},
			},
			element: &struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{},
			expected: expected{element: &struct {
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
			}},
		},
		{
			desc: "slice slice-as-struct pointer",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Fii",
						FieldName: "Foo",
						Kind:      reflect.Slice,
						Tag:       `label-slice-as-struct:"Fii"`,
						Children: []*Node{
							{Name: "bar", FieldName: "Bar", Kind: reflect.String, Value: "haa"},
							{Name: "bir", FieldName: "Bir", Kind: reflect.String, Value: "hii"},
						},
					},
				},
			},
			element: &struct {
				Foo []*struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{},
			expected: expected{element: &struct {
				Foo []*struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{
				Foo: []*struct {
					Bar string
					Bir string
				}{
					{
						Bar: "haa",
						Bir: "hii",
					},
				},
			}},
		},
		{
			desc: "slice slice-as-struct without children",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Fii",
						FieldName: "Foo",
						Tag:       `label-slice-as-struct:"Fii"`,
						Kind:      reflect.Slice,
					},
				},
			},
			element: &struct {
				Foo []struct {
					Bar string
					Bir string
				} `label-slice-as-struct:"Fii"`
			}{},
			expected: expected{error: true},
		},
		{
			desc: "pointer SetDefaults method",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Struct,
						Children: []*Node{
							{Name: "Fuu", FieldName: "Fuu", Value: "huu", Kind: reflect.String},
						},
					},
				},
			},
			element: &struct {
				Foo *InitializedFoo
			}{},
			expected: expected{element: &struct {
				Foo *InitializedFoo
			}{
				Foo: &InitializedFoo{
					Fii: "default",
					Fuu: "huu",
				},
			}},
		},
		{
			desc: "pointer wrong SetDefaults method",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Struct,
						Children: []*Node{
							{Name: "Fuu", FieldName: "Fuu", Value: "huu", Kind: reflect.String},
						},
					},
				},
			},
			element: &struct {
				Foo *wrongInitialledFoo
			}{},
			expected: expected{element: &struct {
				Foo *wrongInitialledFoo
			}{
				Foo: &wrongInitialledFoo{
					Fuu: "huu",
				},
			}},
		},
		{
			desc: "int pointer",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "4", Kind: reflect.Ptr},
				},
			},
			element:  &struct{ Foo *int }{},
			expected: expected{element: &struct{ Foo *int }{Foo: func(v int) *int { return &v }(4)}},
		},
		{
			desc: "bool pointer",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "true", Kind: reflect.Ptr},
				},
			},
			element:  &struct{ Foo *bool }{},
			expected: expected{element: &struct{ Foo *bool }{Foo: func(v bool) *bool { return &v }(true)}},
		},
		{
			desc: "string pointer",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{Name: "Foo", FieldName: "Foo", Value: "bar", Kind: reflect.Ptr},
				},
			},
			element:  &struct{ Foo *string }{},
			expected: expected{element: &struct{ Foo *string }{Foo: func(v string) *string { return &v }("bar")}},
		},
		{
			desc: "embedded",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Struct,
				Children: []*Node{
					{
						Name:      "Foo",
						FieldName: "Foo",
						Kind:      reflect.Struct,
						Children: []*Node{
							{Name: "Fuu", FieldName: "Fuu", Value: "huu", Kind: reflect.String},
						},
					},
				},
			},
			element: &struct {
				Foo struct {
					FiiFoo
				}
			}{},
			expected: expected{element: &struct {
				Foo struct {
					FiiFoo
				}
			}{
				Foo: struct {
					FiiFoo
				}{
					FiiFoo: FiiFoo{
						Fii: "",
						Fuu: "huu",
					},
				},
			}},
		},
		{
			desc: "slice struct",
			node: &Node{
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
			},
			element: &struct {
				Foo []struct {
					Field1 string
					Field2 string
				}
			}{},
			expected: expected{element: &struct {
				Foo []struct {
					Field1 string
					Field2 string
				}
			}{
				Foo: []struct {
					Field1 string
					Field2 string
				}{
					{Field1: "A", Field2: "A"},
					{Field1: "B", Field2: "B"},
					{Field1: "C", Field2: "C"},
				},
			}},
		},
		{
			desc: "slice pointer struct",
			node: &Node{
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
			},
			element: &struct {
				Foo []*struct {
					Field1 string
					Field2 string
				}
			}{},
			expected: expected{element: &struct {
				Foo []*struct {
					Field1 string
					Field2 string
				}
			}{
				Foo: []*struct {
					Field1 string
					Field2 string
				}{
					{Field1: "A", Field2: "A"},
					{Field1: "B", Field2: "B"},
					{Field1: "C", Field2: "C"},
				},
			}},
		},
		{
			desc: "raw value",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Ptr,
				Children: []*Node{
					{Name: "meta", FieldName: "Meta", Kind: reflect.Map, RawValue: map[string]interface{}{
						"aaa": "test",
						"bbb": map[string]interface{}{
							"ccc": "test",
							"ddd": map[string]interface{}{
								"eee": "test",
							},
						},
					}},
					{Name: "name", FieldName: "Name", Value: "test", Kind: reflect.String},
				},
			},
			element: &struct {
				Name string
				Meta map[string]interface{}
			}{},
			expected: expected{element: &struct {
				Name string
				Meta map[string]interface{}
			}{
				Name: "test",
				Meta: map[string]interface{}{
					"aaa": "test",
					"bbb": map[string]interface{}{
						"ccc": "test",
						"ddd": map[string]interface{}{
							"eee": "test",
						},
					},
				},
			}},
		},
		{
			desc: "explicit map of map, raw value",
			node: &Node{
				Name: "traefik",
				Kind: reflect.Ptr,
				Children: []*Node{
					{Name: "meta", FieldName: "Meta", Kind: reflect.Map, Children: []*Node{
						{Name: "aaa", Kind: reflect.Map, Children: []*Node{
							{Name: "bbb", RawValue: map[string]interface{}{
								"ccc": "test1",
								"ddd": "test2",
							}},
							{Name: "eee", Value: "test3", RawValue: map[string]interface{}{
								"eee": "test3",
							}},
						}},
					}},
					{Name: "name", FieldName: "Name", Value: "test", Kind: reflect.String},
				},
			},
			element: &struct {
				Name string
				Meta map[string]map[string]interface{}
			}{},
			expected: expected{element: &struct {
				Name string
				Meta map[string]map[string]interface{}
			}{
				Name: "test",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": map[string]interface{}{
							"ccc": "test1",
							"ddd": "test2",
						},
						"eee": "test3",
					},
				},
			}},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := filler{FillerOpts: FillerOpts{AllowSliceAsStruct: true}}.Fill(test.element, test.node)
			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.element, test.element)
			}
		})
	}
}

type (
	NamedType    string
	NamedTypeInt int
)

type InitializedFoo struct {
	Fii string
	Fuu string
}

func (t *InitializedFoo) SetDefaults() {
	t.Fii = "default"
}

type wrongInitialledFoo struct {
	Fii string
	Fuu string
}

func (t *wrongInitialledFoo) SetDefaults() error {
	t.Fii = "default"
	return nil
}

type Bouya string

type FiiFoo struct {
	Fii string
	Fuu Bouya
}
