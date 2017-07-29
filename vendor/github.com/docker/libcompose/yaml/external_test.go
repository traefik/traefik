package yaml

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestMarshalExternal(t *testing.T) {
	externals := []struct {
		external External
		expected string
	}{
		{
			external: External{},
			expected: `false
`,
		},
		{
			external: External{
				External: false,
			},
			expected: `false
`,
		},
		{
			external: External{
				External: true,
			},
			expected: `true
`,
		},
		{
			external: External{
				External: true,
				Name:     "network-name",
			},
			expected: `name: network-name
`,
		},
	}
	for _, e := range externals {
		bytes, err := yaml.Marshal(e.external)
		assert.Nil(t, err)
		assert.Equal(t, e.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalExternal(t *testing.T) {
	externals := []struct {
		yaml     string
		expected *External
	}{
		{
			yaml: `true`,
			expected: &External{
				External: true,
			},
		},
		{
			yaml: `false`,
			expected: &External{
				External: false,
			},
		},
		{
			yaml: `
name: name-of-network`,
			expected: &External{
				External: true,
				Name:     "name-of-network",
			},
		},
	}
	for _, e := range externals {
		actual := &External{}
		err := yaml.Unmarshal([]byte(e.yaml), actual)
		assert.Nil(t, err)
		assert.Equal(t, e.expected, actual, "should be equal")
	}
}
