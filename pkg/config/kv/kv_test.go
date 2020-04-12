package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	pairs := mapToPairs(map[string]string{
		"test/traefik/fielda":        "bar",
		"test/traefik/fieldb":        "1",
		"test/traefik/fieldc":        "true",
		"test/traefik/fieldd/0":      "one",
		"test/traefik/fieldd/1":      "two",
		"test/traefik/fielde":        "",
		"test/traefik/fieldf/Test1":  "A",
		"test/traefik/fieldf/Test2":  "B",
		"test/traefik/fieldg/0/name": "A",
		"test/traefik/fieldg/1/name": "B",
	})

	element := &sample{}

	err := Decode(pairs, element, "test/traefik")
	require.NoError(t, err)

	expected := &sample{
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
	}
	assert.Equal(t, expected, element)
}

type sample struct {
	FieldA string
	FieldB int
	FieldC bool
	FieldD []string
	FieldE *struct {
		Name string
	} `label:"allowEmpty"`
	FieldF map[string]string
	FieldG []sub
}

type sub struct {
	Name string
}
