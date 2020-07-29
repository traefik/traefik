package file

import (
	"testing"

	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_decodeRawToNode(t *testing.T) {
	testCases := []struct {
		desc     string
		data     map[string]interface{}
		expected *parser.Node
	}{
		{
			desc: "empty",
			data: map[string]interface{}{},
			expected: &parser.Node{
				Name: "traefik",
			},
		},
		{
			desc: "string",
			data: map[string]interface{}{
				"foo": "bar",
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "bar"},
				},
			},
		},
		{
			desc: "string named type",
			data: map[string]interface{}{
				"foo": bar("bar"),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "bar"},
				},
			},
		},
		{
			desc: "bool",
			data: map[string]interface{}{
				"foo": true,
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "true"},
				},
			},
		},
		{
			desc: "int",
			data: map[string]interface{}{
				"foo": 1,
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "int8",
			data: map[string]interface{}{
				"foo": int8(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "int16",
			data: map[string]interface{}{
				"foo": int16(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "int32",
			data: map[string]interface{}{
				"foo": int32(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "int64",
			data: map[string]interface{}{
				"foo": int64(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "uint",
			data: map[string]interface{}{
				"foo": uint(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "uint8",
			data: map[string]interface{}{
				"foo": uint8(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "uint16",
			data: map[string]interface{}{
				"foo": uint16(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "uint32",
			data: map[string]interface{}{
				"foo": uint32(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "uint64",
			data: map[string]interface{}{
				"foo": uint64(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "float32",
			data: map[string]interface{}{
				"foo": float32(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "float64",
			data: map[string]interface{}{
				"foo": float64(1),
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1"},
				},
			},
		},
		{
			desc: "string slice",
			data: map[string]interface{}{
				"foo": []string{"A", "B"},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "A,B"},
				},
			},
		},
		{
			desc: "int slice",
			data: map[string]interface{}{
				"foo": []int{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "int8 slice",
			data: map[string]interface{}{
				"foo": []int8{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "int16 slice",
			data: map[string]interface{}{
				"foo": []int16{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "int32 slice",
			data: map[string]interface{}{
				"foo": []int32{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "int64 slice",
			data: map[string]interface{}{
				"foo": []int64{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "bool slice",
			data: map[string]interface{}{
				"foo": []bool{true, false},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "true,false"},
				},
			},
		},
		{
			desc: "interface (string) slice",
			data: map[string]interface{}{
				"foo": []interface{}{"A", "B"},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "A,B"},
				},
			},
		},
		{
			desc: "interface (int) slice",
			data: map[string]interface{}{
				"foo": []interface{}{1, 2},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "1,2"},
				},
			},
		},
		{
			desc: "2 strings",
			data: map[string]interface{}{
				"foo": "bar",
				"fii": "bir",
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Value: "bir"},
					{Name: "foo", Value: "bar"},
				},
			},
		},
		{
			desc: "string, level 2",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": "bur",
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "bur"}}},
				},
			},
		},
		{
			desc: "int, level 2",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": 1,
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "1"}}},
				},
			},
		},
		{
			desc: "uint, level 2",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": uint(1),
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "1"}}},
				},
			},
		},
		{
			desc: "bool, level 2",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": true,
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "true"}}},
				},
			},
		},
		{
			desc: "string, level 3",
			data: map[string]interface{}{
				"foo": map[interface{}]interface{}{
					"fii": map[interface{}]interface{}{
						"fuu": "bur",
					},
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "bur"}}},
					}},
				},
			},
		},
		{
			desc: "int, level 3",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": 1,
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "1"}}},
				},
			},
		},
		{
			desc: "uint, level 3",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": uint(1),
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "1"}}},
				},
			},
		},
		{
			desc: "bool, level 3",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": true,
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Children: []*parser.Node{{Name: "fuu", Value: "true"}}},
				},
			},
		},
		{
			desc: "struct",
			data: map[string]interface{}{
				"foo": map[interface{}]interface{}{
					"field1": "C",
					"field2": "C",
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "field1", Value: "C"},
						{Name: "field2", Value: "C"},
					}},
				},
			},
		},
		{
			desc: "slice struct 1",
			data: map[string]interface{}{
				"foo": []map[string]interface{}{
					{"field1": "A", "field2": "A"},
					{"field1": "B", "field2": "B"},
					{"field2": "C", "field1": "C"},
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "field1", Value: "A"},
							{Name: "field2", Value: "A"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "field1", Value: "B"},
							{Name: "field2", Value: "B"},
						}},
						{Name: "[2]", Children: []*parser.Node{
							{Name: "field1", Value: "C"},
							{Name: "field2", Value: "C"},
						}},
					}},
				},
			},
		},
		{
			desc: "slice struct 2",
			data: map[string]interface{}{
				"foo": []interface{}{
					map[interface{}]interface{}{
						"field2": "A",
						"field1": "A",
					},
					map[interface{}]interface{}{
						"field1": "B",
						"field2": "B",
					},
					map[interface{}]interface{}{
						"field1": "C",
						"field2": "C",
					},
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "field1", Value: "A"},
							{Name: "field2", Value: "A"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "field1", Value: "B"},
							{Name: "field2", Value: "B"},
						}},
						{Name: "[2]", Children: []*parser.Node{
							{Name: "field1", Value: "C"},
							{Name: "field2", Value: "C"},
						}},
					}},
				},
			},
		},
		{
			desc: "nil value",
			data: map[string]interface{}{
				"fii": map[interface{}]interface{}{
					"fuu": nil,
				},
			},
			expected: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii"},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			node, err := decodeRawToNode(test.data, parser.DefaultRootName)
			require.NoError(t, err)

			assert.Equal(t, test.expected, node)
		})
	}
}

func Test_decodeRawToNode_errors(t *testing.T) {
	testCases := []struct {
		desc string
		data map[string]interface{}
	}{
		{
			desc: "invalid type",
			data: map[string]interface{}{
				"foo": struct{}{},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := decodeRawToNode(test.data, parser.DefaultRootName)
			require.Error(t, err)
		})
	}
}
