package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	testCases := []struct {
		desc     string
		rootName string
		pairs    map[string]string
		expected *sample
	}{
		{
			desc:     "simple case",
			rootName: "baqup",
			pairs: map[string]string{
				"baqup/fielda":        "bar",
				"baqup/fieldb":        "1",
				"baqup/fieldc":        "true",
				"baqup/fieldd/0":      "one",
				"baqup/fieldd/1":      "two",
				"baqup/fielde":        "",
				"baqup/fieldf/Test1":  "A",
				"baqup/fieldf/Test2":  "B",
				"baqup/fieldg/0/name": "A",
				"baqup/fieldg/1/name": "B",
				"baqup/fieldh/":       "foo",
			},
			expected: &sample{
				FieldA: "bar",
				FieldB: 1,
				FieldC: true,
				FieldD: []string{"one", "two"},
				FieldE: &struct {
					Name string
				}{},
				FieldF: map[string]string{
					"Test1": "A",
					"Test2": "B",
				},
				FieldG: []sub{
					{Name: "A"},
					{Name: "B"},
				},
				FieldH: "foo",
			},
		},
		{
			desc:     "multi-level root name",
			rootName: "foo/bar/baqup",
			pairs: map[string]string{
				"foo/bar/baqup/fielda":        "bar",
				"foo/bar/baqup/fieldb":        "2",
				"foo/bar/baqup/fieldc":        "true",
				"foo/bar/baqup/fieldd/0":      "one",
				"foo/bar/baqup/fieldd/1":      "two",
				"foo/bar/baqup/fielde":        "",
				"foo/bar/baqup/fieldf/Test1":  "A",
				"foo/bar/baqup/fieldf/Test2":  "B",
				"foo/bar/baqup/fieldg/0/name": "A",
				"foo/bar/baqup/fieldg/1/name": "B",
				"foo/bar/baqup/fieldh/":       "foo",
			},
			expected: &sample{
				FieldA: "bar",
				FieldB: 2,
				FieldC: true,
				FieldD: []string{"one", "two"},
				FieldE: &struct {
					Name string
				}{},
				FieldF: map[string]string{
					"Test1": "A",
					"Test2": "B",
				},
				FieldG: []sub{
					{Name: "A"},
					{Name: "B"},
				},
				FieldH: "foo",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			element := &sample{}

			err := Decode(mapToPairs(test.pairs), element, test.rootName)
			require.NoError(t, err)

			assert.Equal(t, test.expected, element)
		})
	}
}

type sample struct {
	FieldA string
	FieldB int
	FieldC bool
	FieldD []string
	FieldE *struct {
		Name string
	} `kv:"allowEmpty"`
	FieldF map[string]string
	FieldG []sub
	FieldH string
}

type sub struct {
	Name string
}
