package docker

import (
	"strconv"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	docker "github.com/docker/docker/api/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type fakeTasksClient struct {
	dockerclient.APIClient
	tasks []swarm.Task
	err   error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}

func TestListTasks(t *testing.T) {
	testCases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedTasks []string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1", taskSlot(1), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id2", taskSlot(2), taskStatus(taskState(swarm.TaskStatePending))),
				swarmTask("id3", taskSlot(3)),
				swarmTask("id4", taskSlot(4), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id5", taskSlot(5), taskStatus(taskState(swarm.TaskStateFailed))),
			},
			isGlobalSVC: false,
			expectedTasks: []string{
				"container.1",
				"container.4",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(test.service, test.networks)
			dockerClient := &fakeTasksClient{tasks: test.tasks}
			taskDockerData, _ := listTasks(context.Background(), dockerClient, test.service.ID, dockerData, map[string]*docker.NetworkResource{}, test.isGlobalSVC)

			if len(test.expectedTasks) != len(taskDockerData) {
				t.Errorf("expected tasks %v, got %v", spew.Sdump(test.expectedTasks), spew.Sdump(taskDockerData))
			}

			for i, taskID := range test.expectedTasks {
				if taskDockerData[i].Name != taskID {
					t.Errorf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
				}
			}
		})
	}
}

type fakeServicesClient struct {
	dockerclient.APIClient
	dockerVersion string
	networks      []dockertypes.NetworkResource
	services      []swarm.Service
	err           error
}

func (c *fakeServicesClient) ServiceList(ctx context.Context, options dockertypes.ServiceListOptions) ([]swarm.Service, error) {
	return c.services, c.err
}

func (c *fakeServicesClient) ServerVersion(ctx context.Context) (dockertypes.Version, error) {
	return dockertypes.Version{APIVersion: c.dockerVersion}, c.err
}

func (c *fakeServicesClient) NetworkList(ctx context.Context, options dockertypes.NetworkListOptions) ([]dockertypes.NetworkResource, error) {
	return c.networks, c.err
}

func TestListServices(t *testing.T) {
	testCases := []struct {
		desc             string
		services         []swarm.Service
		dockerVersion    string
		networks         []dockertypes.NetworkResource
		expectedServices []string
	}{
		{
			desc: "Should return no service due to no networks defined",
			services: []swarm.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						labelDockerNetwork:            "barnet",
						labelBackendLoadBalancerSwarm: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("1", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						labelDockerNetwork: "barnet",
					}),
					withEndpointSpec(modeDNSSR)),
			},
			dockerVersion:    "1.30",
			networks:         []dockertypes.NetworkResource{},
			expectedServices: []string{},
		},
		{
			desc: "Should return only service1",
			services: []swarm.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						labelDockerNetwork:            "barnet",
						labelBackendLoadBalancerSwarm: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("yk6l57rfwizjzxxzftn4amaot", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						labelDockerNetwork: "barnet",
					}),
					withEndpointSpec(modeDNSSR)),
			},
			dockerVersion: "1.30",
			networks: []dockertypes.NetworkResource{
				{
					Name:       "network_name",
					ID:         "yk6l57rfwizjzxxzftn4amaot",
					Created:    time.Now(),
					Scope:      "swarm",
					Driver:     "overlay",
					EnableIPv6: false,
					Internal:   true,
					Ingress:    false,
					ConfigOnly: false,
					Options: map[string]string{
						"com.docker.network.driver.overlay.vxlanid_list": "4098",
						"com.docker.network.enable_ipv6":                 "false",
					},
					Labels: map[string]string{
						"com.docker.stack.namespace": "test",
					},
				},
			},
			expectedServices: []string{
				"service1",
			},
		},
	}

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerClient := &fakeServicesClient{services: test.services, dockerVersion: test.dockerVersion, networks: test.networks}

			serviceDockerData, err := listServices(context.Background(), dockerClient)
			assert.NoError(t, err)

			assert.Equal(t, len(test.expectedServices), len(serviceDockerData))
			for i, serviceName := range test.expectedServices {
				assert.Equal(t, serviceName, serviceDockerData[i].Name)
			}
		})
	}
}
