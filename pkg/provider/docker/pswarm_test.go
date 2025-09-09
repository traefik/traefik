package docker

import (
	"strconv"
	"testing"
	"time"

	networktypes "github.com/docker/docker/api/types/network"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTasks(t *testing.T) {
	testCases := []struct {
		service       swarmtypes.Service
		tasks         []swarmtypes.Task
		isGlobalSVC   bool
		expectedTasks []string
		networks      map[string]*networktypes.Summary
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarmtypes.Task{
				swarmTask("id1",
					taskSlot(1),
					taskNetworkAttachment("1", "network1", "overlay", []string{"127.0.0.1"}),
					taskStatus(taskState(swarmtypes.TaskStateRunning)),
				),
				swarmTask("id2",
					taskSlot(2),
					taskNetworkAttachment("1", "network1", "overlay", []string{"127.0.0.2"}),
					taskStatus(taskState(swarmtypes.TaskStatePending)),
				),
				swarmTask("id3",
					taskSlot(3),
					taskNetworkAttachment("1", "network1", "overlay", []string{"127.0.0.3"}),
				),
				swarmTask("id4",
					taskSlot(4),
					taskNetworkAttachment("1", "network1", "overlay", []string{"127.0.0.4"}),
					taskStatus(taskState(swarmtypes.TaskStateRunning)),
				),
				swarmTask("id5",
					taskSlot(5),
					taskNetworkAttachment("1", "network1", "overlay", []string{"127.0.0.5"}),
					taskStatus(taskState(swarmtypes.TaskStateFailed)),
				),
			},
			isGlobalSVC: false,
			expectedTasks: []string{
				"container.1",
				"container.4",
			},
			networks: map[string]*networktypes.Summary{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, test := range testCases {
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()

			var p SwarmProvider
			require.NoError(t, p.Init())

			dockerData, err := p.parseService(t.Context(), test.service, test.networks)
			require.NoError(t, err)

			dockerClient := &fakeTasksClient{tasks: test.tasks}
			taskDockerData, _ := listTasks(t.Context(), dockerClient, test.service.ID, dockerData, test.networks, test.isGlobalSVC)

			if len(test.expectedTasks) != len(taskDockerData) {
				t.Errorf("expected tasks %v, got %v", test.expectedTasks, taskDockerData)
			}

			for i, taskID := range test.expectedTasks {
				if taskDockerData[i].Name != taskID {
					t.Errorf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
				}
			}
		})
	}
}

func TestSwarmProvider_listServices(t *testing.T) {
	testCases := []struct {
		desc             string
		services         []swarmtypes.Service
		tasks            []swarmtypes.Task
		dockerVersion    string
		networks         []networktypes.Summary
		expectedServices []string
	}{
		{
			desc: "Should return no service due to no networks defined",
			services: []swarmtypes.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
						"traefik.swarm.LBSwarm": "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("1", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
						"traefik.swarm.LBSwarm": "true",
					}),
					withEndpointSpec(modeDNSRR)),
			},
			dockerVersion:    "1.30",
			networks:         []networktypes.Summary{},
			expectedServices: []string{},
		},
		{
			desc: "Should return only service1",
			services: []swarmtypes.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
						"traefik.swarm.LBSwarm": "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("yk6l57rfwizjzxxzftn4amaot", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
						"traefik.swarm.LBSwarm": "true",
					}),
					withEndpointSpec(modeDNSRR)),
			},
			dockerVersion: "1.30",
			networks: []networktypes.Summary{
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
						"com.docker.networktypes.driver.overlay.vxlanid_list": "4098",
						"com.docker.networktypes.enable_ipv6":                 "false",
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
		{
			desc: "Should return service1 and service2",
			services: []swarmtypes.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("yk6l57rfwizjzxxzftn4amaot", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						"traefik.swarm.network": "barnet",
					}),
					withEndpointSpec(modeDNSRR)),
			},
			tasks: []swarmtypes.Task{
				swarmTask("id1",
					taskNetworkAttachment("yk6l57rfwizjzxxzftn4amaot", "network_name", "overlay", []string{"127.0.0.1"}),
					taskStatus(taskState(swarmtypes.TaskStateRunning)),
				),
				swarmTask("id2",
					taskNetworkAttachment("yk6l57rfwizjzxxzftn4amaot", "network_name", "overlay", []string{"127.0.0.1"}),
					taskStatus(taskState(swarmtypes.TaskStateRunning)),
				),
			},
			dockerVersion: "1.30",
			networks: []networktypes.Summary{
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
						"com.docker.networktypes.driver.overlay.vxlanid_list": "4098",
						"com.docker.networktypes.enable_ipv6":                 "false",
					},
					Labels: map[string]string{
						"com.docker.stack.namespace": "test",
					},
				},
			},
			expectedServices: []string{
				"service1.0",
				"service1.0",
				"service2.0",
				"service2.0",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dockerClient := &fakeServicesClient{services: test.services, tasks: test.tasks, dockerVersion: test.dockerVersion, networks: test.networks}

			var p SwarmProvider
			require.NoError(t, p.Init())

			serviceDockerData, err := p.listServices(t.Context(), dockerClient)
			assert.NoError(t, err)

			assert.Len(t, serviceDockerData, len(test.expectedServices))
			for i, serviceName := range test.expectedServices {
				if len(serviceDockerData) <= i {
					require.Fail(t, "index", "invalid index %d", i)
				}
				assert.Equal(t, serviceName, serviceDockerData[i].Name)
			}
		})
	}
}

func TestSwarmProvider_parseService_task(t *testing.T) {
	testCases := []struct {
		service     swarmtypes.Service
		tasks       []swarmtypes.Task
		nodes       []swarmtypes.Node
		isGlobalSVC bool
		expected    map[string]dockerData
		networks    map[string]*networktypes.Summary
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarmtypes.Task{
				swarmTask("id1", taskSlot(1)),
				swarmTask("id2", taskSlot(2)),
				swarmTask("id3", taskSlot(3)),
			},
			isGlobalSVC: false,
			expected: map[string]dockerData{
				"id1": {
					Name: "container.1",
				},
				"id2": {
					Name: "container.2",
				},
				"id3": {
					Name: "container.3",
				},
			},
			networks: map[string]*networktypes.Summary{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(serviceName("container")),
			tasks: []swarmtypes.Task{
				swarmTask("id1"),
				swarmTask("id2"),
				swarmTask("id3"),
			},
			isGlobalSVC: true,
			expected: map[string]dockerData{
				"id1": {
					Name: "container.id1",
				},
				"id2": {
					Name: "container.id2",
				},
				"id3": {
					Name: "container.id3",
				},
			},
			networks: map[string]*networktypes.Summary{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceName("container"),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", ""),
				),
			),
			tasks: []swarmtypes.Task{
				swarmTask(
					"id1",
					taskNetworkAttachment("1", "vlan", "macvlan", []string{"127.0.0.1"}),
					taskStatus(
						taskState(swarmtypes.TaskStateRunning),
						taskContainerStatus("c1"),
					),
				),
			},
			isGlobalSVC: true,
			expected: map[string]dockerData{
				"id1": {
					Name: "container.id1",
					NetworkSettings: networkSettings{
						Networks: map[string]*networkData{
							"vlan": {
								Name: "vlan",
								Addr: "10.11.12.13",
							},
						},
					},
				},
			},
			networks: map[string]*networktypes.Summary{
				"1": {
					Name: "vlan",
				},
			},
		},
		{
			service: swarmService(serviceName("container")),
			tasks: []swarmtypes.Task{
				swarmTask("id1",
					taskSlot(1),
					taskNodeID("id1"),
				),
			},
			nodes: []swarmtypes.Node{
				{
					ID: "id1",
					Status: swarmtypes.NodeStatus{
						Addr: "10.11.12.13",
					},
				},
			},
			expected: map[string]dockerData{
				"id1": {
					Name:   "container.1",
					NodeIP: "10.11.12.13",
				},
			},
			networks: map[string]*networktypes.Summary{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, test := range testCases {
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()

			var p SwarmProvider
			require.NoError(t, p.Init())

			dData, err := p.parseService(t.Context(), test.service, test.networks)
			require.NoError(t, err)

			dockerClient := &fakeServicesClient{
				tasks: test.tasks,
				nodes: test.nodes,
			}

			for _, task := range test.tasks {
				taskDockerData, err := parseTasks(t.Context(), dockerClient, task, dData, test.networks, test.isGlobalSVC)
				require.NoError(t, err)

				expected := test.expected[task.ID]
				assert.Equal(t, expected.Name, taskDockerData.Name)
				assert.Equal(t, expected.NodeIP, taskDockerData.NodeIP)
			}
		})
	}
}
