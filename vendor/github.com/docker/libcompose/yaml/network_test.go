package yaml

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestMarshalNetworks(t *testing.T) {
	networks := []struct {
		networks Networks
		expected string
	}{
		{
			networks: Networks{},
			expected: `{}
`,
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
					{
						Name: "network2",
					},
				},
			},
			expected: `network1: {}
network2: {}
`,
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
					{
						Name: "network2",
					},
				},
			},
			expected: `network1:
  aliases:
  - alias1
  - alias2
network2: {}
`,
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
					{
						Name:        "network2",
						IPv4Address: "172.16.238.10",
						IPv6Address: "2001:3984:3989::10",
					},
				},
			},
			expected: `network1:
  aliases:
  - alias1
  - alias2
network2:
  ipv4_address: 172.16.238.10
  ipv6_address: 2001:3984:3989::10
`,
		},
	}
	for _, network := range networks {
		bytes, err := yaml.Marshal(network.networks)
		assert.Nil(t, err)
		assert.Equal(t, network.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalNetworks(t *testing.T) {
	networks := []struct {
		yaml     string
		expected *Networks
	}{
		{
			yaml: `- network1
- network2`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
					{
						Name: "network2",
					},
				},
			},
		},
		{
			yaml: `network1:`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
				},
			},
		},
		{
			yaml: `network1: {}`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
				},
			},
		},
		{
			yaml: `network1:
  aliases:
    - alias1
    - alias2`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
				},
			},
		},
		{
			yaml: `network1:
  aliases:
    - alias1
    - alias2
  ipv4_address: 172.16.238.10
  ipv6_address: 2001:3984:3989::10`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name:        "network1",
						Aliases:     []string{"alias1", "alias2"},
						IPv4Address: "172.16.238.10",
						IPv6Address: "2001:3984:3989::10",
					},
				},
			},
		},
	}
	for _, network := range networks {
		actual := &Networks{}
		err := yaml.Unmarshal([]byte(network.yaml), actual)
		assert.Nil(t, err)
		assert.Equal(t, network.expected, actual, "should be equal")
	}
}
