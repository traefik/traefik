package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeToNode(t *testing.T) {
	type expected struct {
		error bool
		node  *Node
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
			desc: "invalid label, ending by a dot",
			in: map[string]string{
				"traefik.http.": "bar",
			},
			expected: expected{
				error: true,
			},
		},
		{
			desc: "level 1",
			in: map[string]string{
				"traefik.foo": "bar",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "level 1 empty value",
			in: map[string]string{
				"traefik.foo": "",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Value: ""},
				},
			}},
		},
		{
			desc: "level 2",
			in: map[string]string{
				"traefik.foo.bar": "bar",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{{
					Name: "foo",
					Children: []*Node{
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
				"traefik.foo": "bar",
				"traefik.fii": "bir",
			},
			filters: []string{"traefik.Foo"},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "several entries, level 1",
			in: map[string]string{
				"traefik.foo": "bar",
				"traefik.fii": "bur",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "fii", Value: "bur"},
					{Name: "foo", Value: "bar"},
				},
			}},
		},
		{
			desc: "several entries, level 2",
			in: map[string]string{
				"traefik.foo.aaa": "bar",
				"traefik.foo.bbb": "bur",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "aaa", Value: "bar"},
						{Name: "bbb", Value: "bur"},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 2, case insensitive",
			in: map[string]string{
				"traefik.foo.aaa": "bar",
				"traefik.Foo.bbb": "bur",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "Foo", Children: []*Node{
						{Name: "bbb", Value: "bur"},
						{Name: "aaa", Value: "bar"},
					}},
				},
			}},
		},
		{
			desc: "several entries, level 2, 3 children",
			in: map[string]string{
				"traefik.foo.aaa": "bar",
				"traefik.foo.bbb": "bur",
				"traefik.foo.ccc": "bir",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
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
				"traefik.foo.bar.aaa": "bar",
				"traefik.foo.bar.bbb": "bur",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
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
				"traefik.foo.bar.aaa": "bar",
				"traefik.foo.bar.bbb": "bur",
				"traefik.bar.foo.bbb": "bir",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "bar", Children: []*Node{
						{Name: "foo", Children: []*Node{
							{Name: "bbb", Value: "bir"},
						}},
					}},
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
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
				"traefik.foo[0].aaa": "bar0",
				"traefik.foo[0].bbb": "bur0",
				"traefik.foo[1].aaa": "bar1",
				"traefik.foo[1].bbb": "bur1",
			},
			expected: expected{node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "[0]", Children: []*Node{
							{Name: "aaa", Value: "bar0"},
							{Name: "bbb", Value: "bur0"},
						}},
						{Name: "[1]", Children: []*Node{
							{Name: "aaa", Value: "bar1"},
							{Name: "bbb", Value: "bur1"},
						}},
					}},
				},
			}},
		},
		{
			desc: "several entries, invalid slice syntax",
			in: map[string]string{
				"traefik.foo.[0].aaa": "bar0",
				"traefik.foo.[0].bbb": "bur0",
				"traefik.foo.[1].aaa": "bar1",
				"traefik.foo.[1].bbb": "bur1",
			},
			expected: expected{error: true},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			out, err := DecodeToNode(test.in, DefaultRootName, test.filters...)

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
