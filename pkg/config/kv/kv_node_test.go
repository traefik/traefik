package kv

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kvtools/valkeyrie/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/parser"
)

func TestDecodeToNode(t *testing.T) {
	type expected struct {
		error bool
		node  *parser.Node
	}

	testCases := []struct {
		desc     string
		in       map[string]string
		filters  []string
		expected expected
	}{
		{
			desc:     "no label",
			in:       map[string]string{},
			expected: expected{node: nil},
		},
		{
			desc: "level 1",
			in: map[string]string{
				"traefik/foo": "bar",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "level 1 empty value",
			in: map[string]string{
				"traefik/foo": "",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: ""},
				},
			}},
		},
		{
			desc: "level 2",
			in: map[string]string{
				"traefik/foo/bar": "bar",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{{
					Name: "foo",
					Children: []*parser.Node{
						{Name: "bar", Value: "bar"},
					},
				}},
			}},
		},
		{
			desc: "several entries, level 0",
			in: map[string]string{
				"traefik": "bar",
				"traefic": "bur",
			},
			expected: expected{error: true},
		},
		{
			desc: "several entries, prefix filter",
			in: map[string]string{
				"traefik/foo": "bar",
				"traefik/fii": "bir",
			},
			filters: []string{"traefik/Foo"},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "several entries, level 1",
			in: map[string]string{
				"traefik/foo": "bar",
				"traefik/fii": "bur",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "fii", Value: "bur"},
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "several entries, level 2",
			in: map[string]string{
				"traefik/foo/aaa": "bar",
				"traefik/foo/bbb": "bur",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "aaa", Value: "bar"},
						{Name: "bbb", Value: "bur"},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 2, case insensitive",
			in: map[string]string{
				"traefik/foo/aaa": "bar",
				"traefik/Foo/bbb": "bur",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "Foo", Children: []*parser.Node{
						{Name: "bbb", Value: "bur"},
						{Name: "aaa", Value: "bar"},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 2, 3 children",
			in: map[string]string{
				"traefik/foo/aaa": "bar",
				"traefik/foo/bbb": "bur",
				"traefik/foo/ccc": "bir",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "aaa", Value: "bar"},
						{Name: "bbb", Value: "bur"},
						{Name: "ccc", Value: "bir"},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 3",
			in: map[string]string{
				"traefik/foo/bar/aaa": "bar",
				"traefik/foo/bar/bbb": "bur",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "bar", Children: []*parser.Node{
							{Name: "aaa", Value: "bar"},
							{Name: "bbb", Value: "bur"},
						}},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 3, 2 children level 1",
			in: map[string]string{
				"traefik/foo/bar/aaa": "bar",
				"traefik/foo/bar/bbb": "bur",
				"traefik/bar/foo/bbb": "bir",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "bar", Children: []*parser.Node{
						{Name: "foo", Children: []*parser.Node{
							{Name: "bbb", Value: "bir"},
						}},
					}},
					{Name: "foo", Children: []*parser.Node{
						{Name: "bar", Children: []*parser.Node{
							{Name: "aaa", Value: "bar"},
							{Name: "bbb", Value: "bur"},
						}},
					}},
				},
			}},
		},
		{
			desc: "several entries, slice syntax",
			in: map[string]string{
				"traefik/foo/0/aaa": "bar0",
				"traefik/foo/0/bbb": "bur0",
				"traefik/foo/1/aaa": "bar1",
				"traefik/foo/1/bbb": "bur1",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "aaa", Value: "bar0"},
							{Name: "bbb", Value: "bur0"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "aaa", Value: "bar1"},
							{Name: "bbb", Value: "bur1"},
						}},
					}},
				},
			}},
		},
		{
			desc: "several entries, slice in slice of struct",
			in: map[string]string{
				"traefik/foo/0/aaa/0": "bar0",
				"traefik/foo/0/aaa/1": "bar1",
				"traefik/foo/1/aaa/0": "bar2",
				"traefik/foo/1/aaa/1": "bar3",
			},
			expected: expected{node: &parser.Node{
				Name: "traefik",
				Children: []*parser.Node{
					{Name: "foo", Children: []*parser.Node{
						{Name: "[0]", Children: []*parser.Node{
							{Name: "aaa", Value: "bar0,bar1"},
						}},
						{Name: "[1]", Children: []*parser.Node{
							{Name: "aaa", Value: "bar2,bar3"},
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

			out, err := DecodeToNode(mapToPairs(test.in), "traefik", test.filters...)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if !assert.Equal(t, test.expected.node, out) {
					bytes, err := json.MarshalIndent(out, "", "  ")
					require.NoError(t, err)
					fmt.Println(string(bytes))
				}
			}
		})
	}
}

func mapToPairs(in map[string]string) []*store.KVPair {
	var out []*store.KVPair
	for k, v := range in {
		out = append(out, &store.KVPair{Key: k, Value: []byte(v)})
	}
	return out
}
