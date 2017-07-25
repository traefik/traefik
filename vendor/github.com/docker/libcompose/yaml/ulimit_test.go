package yaml

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestMarshalUlimit(t *testing.T) {
	ulimits := []struct {
		ulimits  *Ulimits
		expected string
	}{
		{
			ulimits: &Ulimits{
				Elements: []Ulimit{
					{
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
						Name: "nproc",
					},
				},
			},
			expected: `nproc: 65535
`,
		},
		{
			ulimits: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
				},
			},
			expected: `nofile:
  soft: 20000
  hard: 40000
`,
		},
	}

	for _, ulimit := range ulimits {

		bytes, err := yaml.Marshal(ulimit.ulimits)

		assert.Nil(t, err)
		assert.Equal(t, ulimit.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalUlimits(t *testing.T) {
	ulimits := []struct {
		yaml     string
		expected *Ulimits
	}{
		{
			yaml: "nproc: 65535",
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nproc",
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
					},
				},
			},
		},
		{
			yaml: `nofile:
  soft: 20000
  hard: 40000`,
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
				},
			},
		},
		{
			yaml: `nproc: 65535
nofile:
  soft: 20000
  hard: 40000`,
			expected: &Ulimits{
				Elements: []Ulimit{
					{
						Name: "nofile",
						ulimitValues: ulimitValues{
							Soft: 20000,
							Hard: 40000,
						},
					},
					{
						Name: "nproc",
						ulimitValues: ulimitValues{
							Soft: 65535,
							Hard: 65535,
						},
					},
				},
			},
		},
	}

	for _, ulimit := range ulimits {
		actual := &Ulimits{}
		err := yaml.Unmarshal([]byte(ulimit.yaml), actual)

		assert.Nil(t, err)
		assert.Equal(t, ulimit.expected, actual, "should be equal")
	}
}
