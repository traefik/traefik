package yaml

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestMarshalVolumes(t *testing.T) {
	volumes := []struct {
		volumes  Volumes
		expected string
	}{
		{
			volumes: Volumes{},
			expected: `[]
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- /in/the/container
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
						AccessMode:  "ro",
					},
				},
			},
			expected: `- ./a/path:/in/the/container:ro
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- ./a/path:/in/the/container
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
					{
						Source:      "named",
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- ./a/path:/in/the/container
- named:/in/the/container
`,
		},
	}
	for _, volume := range volumes {
		bytes, err := yaml.Marshal(volume.volumes)
		assert.Nil(t, err)
		assert.Equal(t, volume.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalVolumes(t *testing.T) {
	volumes := []struct {
		yaml     string
		expected *Volumes
	}{
		{
			yaml: `- ./a/path:/in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
				},
			},
		},
		{
			yaml: `- /in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Destination: "/in/the/container",
					},
				},
			},
		},
		{
			yaml: `- /a/path:/in/the/container:ro`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "/a/path",
						Destination: "/in/the/container",
						AccessMode:  "ro",
					},
				},
			},
		},
		{
			yaml: `- /a/path:/in/the/container
- named:/somewhere/in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "/a/path",
						Destination: "/in/the/container",
					},
					{
						Source:      "named",
						Destination: "/somewhere/in/the/container",
					},
				},
			},
		},
	}
	for _, volume := range volumes {
		actual := &Volumes{}
		err := yaml.Unmarshal([]byte(volume.yaml), actual)
		assert.Nil(t, err)
		assert.Equal(t, volume.expected, actual, "should be equal")
	}
}
