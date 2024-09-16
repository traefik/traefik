package docker

import (
	"context"
	"strconv"
	"testing"

	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getPort_docker(t *testing.T) {
	testCases := []struct {
		desc       string
		container  docker.ContainerJSON
		serverPort string
		expected   string
	}{
		{
			desc:      "no binding, no server port label",
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			desc: "binding, no server port label",
			container: containerJSON(ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "80",
		},
		{
			desc: "binding, multiple ports, no server port label",
			container: containerJSON(ports(nat.PortMap{
				"80/tcp":  {},
				"443/tcp": {},
			})),
			expected: "80",
		},
		{
			desc:       "no binding, server port label",
			container:  containerJSON(),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, server port label",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": {},
				})),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, multiple ports, server port label",
			container: containerJSON(ports(nat.PortMap{
				"8080/tcp": {},
				"80/tcp":   {},
			})),
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func Test_getPort_swarm(t *testing.T) {
	testCases := []struct {
		service    swarm.Service
		serverPort string
		networks   map[string]*docker.NetworkResource
		expected   string
	}{
		{
			service: swarmService(
				withEndpointSpec(modeDNSRR),
			),
			networks:   map[string]*docker.NetworkResource{},
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for serviceID, test := range testCases {
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			p := SwarmProvider{}

			dData, err := p.parseService(context.Background(), test.service, test.networks)
			require.NoError(t, err)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}
