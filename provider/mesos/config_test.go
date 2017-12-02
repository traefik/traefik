package mesos

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/stretchr/testify/assert"
)

// FIXME fill this test!!
func TestBuildConfiguration(t *testing.T) {
	cases := []struct {
		applicationsError bool
		tasksError        bool
		mesosTask         state.Task
		expected          bool
		exposedByDefault  bool
		expectedNil       bool
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{}

	for _, c := range cases {
		provider := &Provider{
			Domain:           "docker.localhost",
			ExposedByDefault: true,
		}
		actualConfig := provider.buildConfiguration()
		if c.expectedNil {
			if actualConfig != nil {
				t.Fatalf("Should have been nil, got %v", actualConfig)
			}
		} else {
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Fatalf("Expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Fatalf("Expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		}
	}
}

func TestTaskFilter(t *testing.T) {
	testCases := []struct {
		mesosTask        state.Task
		expected         bool
		exposedByDefault bool
	}{
		{
			mesosTask:        state.Task{},
			expected:         false,
			exposedByDefault: true,
		},
		{
			mesosTask:        task(statuses(status(setState("TASK_RUNNING")))),
			expected:         false,
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "false"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // because label traefik.enable = false
			exposedByDefault: false,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true,
			exposedByDefault: false,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true,
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "false"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // because label traefik.enable = false (even wherek exposedByDefault = true)
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPortIndex, "1",
					label.TraefikPort, "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.portIndex & traefik.port cannot be set both
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPortIndex, "1"),
				discovery(setDiscoveryPorts("TCP", 80, "WEB HTTP", "TCP", 443, "WEB HTTPS")),
			),
			expected:         true,
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true"),
				discovery(setDiscoveryPorts("TCP", 80, "WEB HTTP", "TCP", 443, "WEB HTTPS")),
			),
			expected:         true, // Default to first index
			exposedByDefault: true,
		},
		{
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPortIndex, "1"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.portIndex and discoveryPorts don't correspond
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPortIndex, "0"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // traefik.portIndex and discoveryPorts correspond
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPort, "TRAEFIK"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.port is not an integer
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPort, "443"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.port is not the same as discovery.port
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(true))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPort, "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // traefik.port is the same as discovery.port
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPort, "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // No healthCheck
			exposedByDefault: true,
		}, {
			mesosTask: task(
				statuses(
					status(
						setState("TASK_RUNNING"),
						setHealthy(false))),
				setLabels(label.TraefikEnable, "true",
					label.TraefikPort, "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // HealthCheck at false
			exposedByDefault: true,
		},
	}

	for index, test := range testCases {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()

			actual := taskFilter(test.mesosTask, test.exposedByDefault)
			if actual != test.expected {
				t.Logf("Statuses : %v", test.mesosTask.Statuses)
				t.Logf("Label : %v", test.mesosTask.Labels)
				t.Logf("DiscoveryInfo : %v", test.mesosTask.DiscoveryInfo)
				t.Fatalf("Expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestTaskRecords(t *testing.T) {
	var task = state.Task{
		SlaveID: "s_id",
		State:   "TASK_RUNNING",
	}
	var framework = state.Framework{
		Tasks: []state.Task{task},
	}
	var slave = state.Slave{
		ID:       "s_id",
		Hostname: "127.0.0.1",
	}
	var taskState = state.State{
		Slaves:     []state.Slave{slave},
		Frameworks: []state.Framework{framework},
	}

	var p = taskRecords(taskState)
	if len(p) == 0 {
		t.Fatal("No task")
	}
	if p[0].SlaveIP != slave.Hostname {
		t.Fatalf("The SlaveIP (%s) should be set with the slave hostname (%s)", p[0].SlaveID, slave.Hostname)
	}
}

func TestGetSubDomain(t *testing.T) {
	providerGroups := &Provider{GroupsAsSubDomains: true}
	providerNoGroups := &Provider{GroupsAsSubDomains: false}

	testCases := []struct {
		path     string
		expected string
		provider *Provider
	}{
		{"/test", "test", providerNoGroups},
		{"/test", "test", providerGroups},
		{"/a/b/c/d", "d.c.b.a", providerGroups},
		{"/b/a/d/c", "c.d.a.b", providerGroups},
		{"/d/c/b/a", "a.b.c.d", providerGroups},
		{"/c/d/a/b", "b.a.d.c", providerGroups},
		{"/a/b/c/d", "a-b-c-d", providerNoGroups},
		{"/b/a/d/c", "b-a-d-c", providerNoGroups},
		{"/d/c/b/a", "d-c-b-a", providerNoGroups},
		{"/c/d/a/b", "c-d-a-b", providerNoGroups},
	}

	for _, test := range testCases {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			actual := test.provider.getSubDomain(test.path)

			assert.Equal(t, test.expected, actual)
		})
	}
}
