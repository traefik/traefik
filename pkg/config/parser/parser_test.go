package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Tomato struct {
	Name string
	Meta map[string]interface{}
}

type Potato struct {
	Name string
	Meta map[string]map[string]interface{}
}

func TestDecode_RawValue(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		elt      interface{}
		expected interface{}
	}{
		{
			desc: "level 1",
			elt:  &Tomato{},
			labels: map[string]string{
				"traefik.name":     "test",
				"traefik.meta.aaa": "test",
			},
			expected: &Tomato{
				Name: "test",
				Meta: map[string]interface{}{
					"aaa": "test",
				},
			},
		},
		{
			desc: "level 2",
			labels: map[string]string{
				"traefik.name":         "test",
				"traefik.meta.aaa":     "test",
				"traefik.meta.bbb.ccc": "test",
			},
			elt: &Tomato{},
			expected: &Tomato{
				Name: "test",
				Meta: map[string]interface{}{
					"aaa": "test",
					"bbb": map[string]interface{}{
						"ccc": "test",
					},
				},
			},
		},
		{
			desc: "level 3",
			labels: map[string]string{
				"traefik.name":             "test",
				"traefik.meta.aaa":         "test",
				"traefik.meta.bbb.ccc":     "test",
				"traefik.meta.bbb.ddd.eee": "test",
			},
			elt: &Tomato{},
			expected: &Tomato{
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
			},
		},
		{
			desc: "struct slice, one entry",
			elt:  &Tomato{},
			labels: map[string]string{
				"traefik.name":            "test1",
				"traefik.meta.aaa[0].bbb": "test2",
				"traefik.meta.aaa[0].ccc": "test3",
			},
			expected: &Tomato{
				Name: "test1",
				Meta: map[string]interface{}{
					"aaa": []interface{}{
						map[string]interface{}{
							"bbb": "test2",
							"ccc": "test3",
						},
					},
				},
			},
		},
		{
			desc: "struct slice, multiple entries",
			elt:  &Tomato{},
			labels: map[string]string{
				"traefik.name":            "test1",
				"traefik.meta.aaa[0].bbb": "test2",
				"traefik.meta.aaa[0].ccc": "test3",
				"traefik.meta.aaa[1].bbb": "test4",
				"traefik.meta.aaa[1].ccc": "test5",
				"traefik.meta.aaa[2].bbb": "test6",
				"traefik.meta.aaa[2].ccc": "test7",
			},
			expected: &Tomato{
				Name: "test1",
				Meta: map[string]interface{}{
					"aaa": []interface{}{
						map[string]interface{}{
							"bbb": "test2",
							"ccc": "test3",
						},
						map[string]interface{}{
							"bbb": "test4",
							"ccc": "test5",
						},
						map[string]interface{}{
							"bbb": "test6",
							"ccc": "test7",
						},
					},
				},
			},
		},
		{
			desc: "explicit map of map, level 1",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":         "test",
				"traefik.meta.aaa.bbb": "test1",
			},
			expected: &Potato{
				Name: "test",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": "test1",
					},
				},
			},
		},
		{
			desc: "explicit map of map, level 2",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":         "test",
				"traefik.meta.aaa.bbb": "test1",
				"traefik.meta.aaa.ccc": "test2",
			},
			expected: &Potato{
				Name: "test",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": "test1",
						"ccc": "test2",
					},
				},
			},
		},
		{
			desc: "explicit map of map, level 3",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":             "test",
				"traefik.meta.aaa.bbb.ccc": "test1",
				"traefik.meta.aaa.bbb.ddd": "test2",
				"traefik.meta.aaa.eee":     "test3",
			},
			expected: &Potato{
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
			},
		},
		{
			desc: "explicit map of map, level 4",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":                 "test",
				"traefik.meta.aaa.bbb.ccc.ddd": "test1",
				"traefik.meta.aaa.bbb.ccc.eee": "test2",
				"traefik.meta.aaa.bbb.fff":     "test3",
				"traefik.meta.aaa.ggg":         "test4",
			},
			expected: &Potato{
				Name: "test",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": map[string]interface{}{
							"ccc": map[string]interface{}{
								"ddd": "test1",
								"eee": "test2",
							},
							"fff": "test3",
						},
						"ggg": "test4",
					},
				},
			},
		},
		{
			desc: "explicit map of map, struct slice, level 1, one entry",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":                "test1",
				"traefik.meta.aaa.bbb[0].ccc": "test2",
				"traefik.meta.aaa.bbb[0].ddd": "test3",
			},
			expected: &Potato{
				Name: "test1",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": []interface{}{
							map[string]interface{}{
								"ccc": "test2",
								"ddd": "test3",
							},
						},
					},
				},
			},
		},
		{
			desc: "explicit map of map, struct slice, level 1, multiple entries",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":                "test1",
				"traefik.meta.aaa.bbb[0].ccc": "test2",
				"traefik.meta.aaa.bbb[0].ddd": "test3",
				"traefik.meta.aaa.bbb[1].ccc": "test4",
				"traefik.meta.aaa.bbb[1].ddd": "test5",
				"traefik.meta.aaa.bbb[2].ccc": "test6",
				"traefik.meta.aaa.bbb[2].ddd": "test7",
			},
			expected: &Potato{
				Name: "test1",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": []interface{}{
							map[string]interface{}{
								"ccc": "test2",
								"ddd": "test3",
							},
							map[string]interface{}{
								"ccc": "test4",
								"ddd": "test5",
							},
							map[string]interface{}{
								"ccc": "test6",
								"ddd": "test7",
							},
						},
					},
				},
			},
		},
		{
			desc: "explicit map of map, struct slice, level 2, one entry",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":                    "test1",
				"traefik.meta.aaa.bbb.ccc[0].ddd": "test2",
				"traefik.meta.aaa.bbb.ccc[0].eee": "test3",
			},
			expected: &Potato{
				Name: "test1",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": map[string]interface{}{
							"ccc": []interface{}{
								map[string]interface{}{
									"ddd": "test2",
									"eee": "test3",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "explicit map of map, struct slice, level 2, multiple entries",
			elt:  &Potato{},
			labels: map[string]string{
				"traefik.name":                    "test1",
				"traefik.meta.aaa.bbb.ccc[0].ddd": "test2",
				"traefik.meta.aaa.bbb.ccc[0].eee": "test3",
				"traefik.meta.aaa.bbb.ccc[1].ddd": "test4",
				"traefik.meta.aaa.bbb.ccc[1].eee": "test5",
				"traefik.meta.aaa.bbb.ccc[2].ddd": "test6",
				"traefik.meta.aaa.bbb.ccc[2].eee": "test7",
			},
			expected: &Potato{
				Name: "test1",
				Meta: map[string]map[string]interface{}{
					"aaa": {
						"bbb": map[string]interface{}{
							"ccc": []interface{}{
								map[string]interface{}{
									"ddd": "test2",
									"eee": "test3",
								},
								map[string]interface{}{
									"ddd": "test4",
									"eee": "test5",
								},
								map[string]interface{}{
									"ddd": "test6",
									"eee": "test7",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		if test.desc != "level 3" {
			continue
		}

		test := test
		t.Run(test.desc, func(t *testing.T) {
			err := Decode(test.labels, test.elt, "traefik")
			require.NoError(t, err)

			assert.Equal(t, test.expected, test.elt)
		})
	}
}
