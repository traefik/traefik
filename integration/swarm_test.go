package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	networktypes "github.com/moby/moby/api/types/network"
	swarmtypes "github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/traefik/traefik/v3/integration/try"
)

// whoamiServices lists the Swarm services created by SwarmSuite, covering every
// combination of publish mode (ingress/host) and published port (dynamic/fixed).
// This mirrors https://github.com/traefik/traefik/issues/13239: useBindPortIP silently
// fell back to the (potentially unreachable) overlay IP for every Swarm service, because
// the published port binding was never resolved for Swarm tasks, unlike standalone
// containers.
var whoamiServices = []struct {
	name          string
	publishMode   swarmtypes.PortConfigPublishMode
	publishedPort uint32 // 0 requests a dynamically assigned port, as reported in the issue.
}{
	{name: "whoami-ingress-dynamic", publishMode: swarmtypes.PortConfigPublishModeIngress, publishedPort: 0},
	{name: "whoami-ingress-fixed", publishMode: swarmtypes.PortConfigPublishModeIngress, publishedPort: 30080},
	{name: "whoami-host-dynamic", publishMode: swarmtypes.PortConfigPublishModeHost, publishedPort: 0},
	{name: "whoami-host-fixed", publishMode: swarmtypes.PortConfigPublishModeHost, publishedPort: 30081},
}

// SwarmSuite exercises the Swarm provider against a real Docker Swarm cluster, in
// particular the providers.swarm.useBindPortIP option.
type SwarmSuite struct {
	BaseSuite

	dockerClient   *testcontainers.DockerClient
	overlayNetwork string
}

func TestSwarmSuite(t *testing.T) {
	suite.Run(t, new(SwarmSuite))
}

func (s *SwarmSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	dockerClient, err := testcontainers.NewDockerClientWithOpts(s.T().Context())
	require.NoError(s.T(), err)
	s.dockerClient = dockerClient

	_, err = dockerClient.SwarmInit(s.T().Context(), client.SwarmInitOptions{})
	if err != nil && !strings.Contains(err.Error(), "already part of a swarm") {
		require.NoError(s.T(), err)
	}

	// Host-mode published ports are only reported on the task, and Swarm only attaches
	// a task's overlay network when the service explicitly requests one: an otherwise
	// unattached host-mode service is invisible to the Swarm provider regardless of
	// useBindPortIP, since it never gets an IP to build a router for.
	networkCreated, err := dockerClient.NetworkCreate(s.T().Context(), "traefik-swarm-test-net", client.NetworkCreateOptions{
		Driver:     "overlay",
		Attachable: true,
	})
	require.NoError(s.T(), err)
	s.overlayNetwork = networkCreated.ID

	for _, svc := range whoamiServices {
		s.createWhoamiService(svc.name, svc.publishMode, svc.publishedPort)
	}

	for _, svc := range whoamiServices {
		s.waitServiceRunning(svc.name)
	}
}

func (s *SwarmSuite) TearDownSuite() {
	for _, svc := range whoamiServices {
		_, _ = s.dockerClient.ServiceRemove(s.T().Context(), svc.name, client.ServiceRemoveOptions{})
	}

	if s.overlayNetwork != "" {
		_ = try.Do(10*time.Second, func() error {
			_, err := s.dockerClient.NetworkRemove(s.T().Context(), s.overlayNetwork, client.NetworkRemoveOptions{})
			return err
		})
	}

	s.BaseSuite.TearDownSuite()
}

// TestUseBindPortIP verifies that, for every combination of Swarm publish mode
// (ingress/host) and published port (dynamic/fixed), Traefik with useBindPortIP=true
// routes requests to the Swarm node's published port instead of falling back to the
// (here, unreachable) overlay network IP.
func (s *SwarmSuite) TestUseBindPortIP() {
	file := s.adaptFile("fixtures/swarm/use_bind_port_ip.toml", struct{ DockerHost string }{
		DockerHost: s.getDockerHost(),
	})

	s.traefikCmd(withConfigFile(file))

	for _, svc := range whoamiServices {
		req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8000/", nil)
		require.NoError(s.T(), err)
		req.Host = svc.name + ".test"

		resp, err := try.ResponseUntilStatusCode(req, 30*time.Second, http.StatusOK)
		require.NoErrorf(s.T(), err, "service %s never became reachable through useBindPortIP", svc.name)

		err = try.Request(req, 2*time.Second, try.BodyContains("Hostname:"))
		require.NoError(s.T(), err)

		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

// createWhoamiService deploys a single-replica traefik/whoami service publishing
// container port 80 in the given mode, with a Traefik router keyed on the service name.
func (s *SwarmSuite) createWhoamiService(name string, publishMode swarmtypes.PortConfigPublishMode, publishedPort uint32) {
	_, err := s.dockerClient.ServiceCreate(s.T().Context(), client.ServiceCreateOptions{
		Spec: swarmtypes.ServiceSpec{
			Annotations: swarmtypes.Annotations{
				Name: name,
				Labels: map[string]string{
					"traefik.enable": "true",
					fmt.Sprintf("traefik.http.routers.%s.rule", name): fmt.Sprintf("Host(`%s.test`)", name),
				},
			},
			TaskTemplate: swarmtypes.TaskSpec{
				ContainerSpec: &swarmtypes.ContainerSpec{
					Image: "traefik/whoami:latest",
				},
				Networks: []swarmtypes.NetworkAttachmentConfig{
					{Target: s.overlayNetwork},
				},
			},
			EndpointSpec: &swarmtypes.EndpointSpec{
				Ports: []swarmtypes.PortConfig{
					{
						Protocol:      networktypes.TCP,
						TargetPort:    80,
						PublishedPort: publishedPort,
						PublishMode:   publishMode,
					},
				},
			},
		},
	})
	require.NoError(s.T(), err)
}

func (s *SwarmSuite) waitServiceRunning(name string) {
	err := try.Do(45*time.Second, func() error {
		tasks, err := s.dockerClient.TaskList(s.T().Context(), client.TaskListOptions{
			Filters: make(client.Filters).Add("service", name).Add("desired-state", "running"),
		})
		if err != nil {
			return err
		}

		for _, task := range tasks.Items {
			if task.Status.State == swarmtypes.TaskStateRunning {
				return nil
			}
		}
		return fmt.Errorf("service %s has no running task yet", name)
	})
	require.NoError(s.T(), err)
}
