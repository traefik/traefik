package docker

import (
	"context"
	"strconv"
	"testing"

	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
	swarmtypes "github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeContainersClient struct {
	client.APIClient

	containers []containertypes.Summary
	inspect    containertypes.InspectResponse
	inspectErr error
}

func (c *fakeContainersClient) ContainerList(_ context.Context, _ client.ContainerListOptions) (client.ContainerListResult, error) {
	return client.ContainerListResult{Items: c.containers}, nil
}

func (c *fakeContainersClient) ContainerInspect(_ context.Context, _ string, _ client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	return client.ContainerInspectResult{Container: c.inspect}, c.inspectErr
}

func Test_getPort_docker(t *testing.T) {
	testCases := []struct {
		desc       string
		container  containertypes.InspectResponse
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
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("80/tcp"): {},
			})),
			expected: "80",
		},
		{
			desc: "binding, multiple ports, no server port label",
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("80/tcp"):  {},
				networktypes.MustParsePort("443/tcp"): {},
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
				ports(networktypes.PortMap{
					networktypes.MustParsePort("80/tcp"): {},
				})),
			serverPort: "8080",
			expected:   "8080",
		},
		{
			desc: "binding, multiple ports, server port label",
			container: containerJSON(ports(networktypes.PortMap{
				networktypes.MustParsePort("8080/tcp"): {},
				networktypes.MustParsePort("80/tcp"):   {},
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

func Test_listContainers_inspectFallback(t *testing.T) {
	labels := map[string]string{"traefik.http.services.foo.loadbalancer.server.port": "8080"}

	testCases := []struct {
		desc       string
		summary    containertypes.Summary
		expectKeep bool
	}{
		{
			desc:       "inspect error without labels is skipped",
			summary:    containertypes.Summary{ID: "abc", Names: []string{"/foo"}},
			expectKeep: false,
		},
		{
			desc:       "inspect error with labels falls back to summary",
			summary:    containertypes.Summary{ID: "abc", Names: []string{"/foo"}, Labels: labels},
			expectKeep: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fake := &fakeContainersClient{
				containers: []containertypes.Summary{test.summary},
				inspectErr: assert.AnError,
			}

			p := &Provider{Shared: Shared{ExposedByDefault: true}}
			got, err := p.listContainers(t.Context(), fake)
			require.NoError(t, err)

			if !test.expectKeep {
				assert.Empty(t, got)
				return
			}

			require.Len(t, got, 1)
			assert.Equal(t, "abc", got[0].ID)
			assert.Equal(t, "/foo", got[0].Name)
			assert.Equal(t, labels, got[0].Labels)
		})
	}
}

func Test_getPort_swarm(t *testing.T) {
	testCases := []struct {
		service    swarmtypes.Service
		serverPort string
		networks   map[string]*networktypes.Summary
		expected   string
	}{
		{
			service: swarmService(
				withEndpointSpec(modeDNSRR),
			),
			networks:   map[string]*networktypes.Summary{},
			serverPort: "8080",
			expected:   "8080",
		},
	}

	for serviceID, test := range testCases {
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			var p SwarmProvider
			require.NoError(t, p.Init())

			dData, err := p.parseService(t.Context(), test.service, test.networks)
			require.NoError(t, err)

			actual := getPort(dData, test.serverPort)
			assert.Equal(t, test.expected, actual)
		})
	}
}
