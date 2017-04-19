package mesos

import (
	"reflect"
	"testing"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records/state"
)

func TestMesosTaskFilter(t *testing.T) {

	cases := []struct {
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
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "false"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // because label traefik.enable = false
			exposedByDefault: false,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true,
			exposedByDefault: false,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true,
			exposedByDefault: true,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "false"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // because label traefik.enable = false (even wherek exposedByDefault = true)
			exposedByDefault: true,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.portIndex", "1",
					"traefik.port", "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.portIndex & traefik.port cannot be set both
			exposedByDefault: true,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.portIndex", "1"),
				discovery(setDiscoveryPorts("TCP", 80, "WEB HTTP", "TCP", 443, "WEB HTTPS")),
			),
			expected:         true,
			exposedByDefault: true,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true"),
				discovery(setDiscoveryPorts("TCP", 80, "WEB HTTP", "TCP", 443, "WEB HTTPS")),
			),
			expected:         true, // Default to first index
			exposedByDefault: true,
		},
		{
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.portIndex", "1"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.portIndex and discoveryPorts don't correspond
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.portIndex", "0"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // traefik.portIndex and discoveryPorts correspond
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.port", "TRAEFIK"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.port is not an integer
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.port", "443"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // traefik.port is not the same as discovery.port
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(true))),
				setLabels("traefik.enable", "true",
					"traefik.port", "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // traefik.port is the same as discovery.port
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"))),
				setLabels("traefik.enable", "true",
					"traefik.port", "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         true, // No healthCheck
			exposedByDefault: true,
		}, {
			mesosTask: task(statuses(status(
				setState("TASK_RUNNING"),
				setHealthy(false))),
				setLabels("traefik.enable", "true",
					"traefik.port", "80"),
				discovery(setDiscoveryPort("TCP", 80, "WEB")),
			),
			expected:         false, // HealthCheck at false
			exposedByDefault: true,
		},
	}

	for _, c := range cases {
		actual := mesosTaskFilter(c.mesosTask, c.exposedByDefault)
		log.Errorf("Statuses : %v", c.mesosTask.Statuses)
		log.Errorf("Label : %v", c.mesosTask.Labels)
		log.Errorf("DiscoveryInfo : %v", c.mesosTask.DiscoveryInfo)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v", c.expected, actual)
		}
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
	var state = state.State{
		Slaves:     []state.Slave{slave},
		Frameworks: []state.Framework{framework},
	}

	provider := &Provider{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}
	var p = provider.taskRecords(state)
	if len(p) == 0 {
		t.Fatal("taskRecord should return at least one task")
	}
	if p[0].SlaveIP != slave.Hostname {
		t.Fatalf("The SlaveIP (%s) should be set with the slave hostname (%s)", p[0].SlaveID, slave.Hostname)
	}
}

func TestMesosLoadConfig(t *testing.T) {
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
		actualConfig := provider.loadMesosConfig()
		if c.expectedNil {
			if actualConfig != nil {
				t.Fatalf("Should have been nil, got %v", actualConfig)
			}
		} else {
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		}
	}
}

func TestMesosGetSubDomain(t *testing.T) {
	providerGroups := &Provider{GroupsAsSubDomains: true}
	providerNoGroups := &Provider{GroupsAsSubDomains: false}

	apps := []struct {
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

	for _, a := range apps {
		actual := a.provider.getSubDomain(a.path)

		if actual != a.expected {
			t.Errorf("expected %q, got %q", a.expected, actual)
		}
	}
}

// test helpers

type (
	taskOpt   func(*state.Task)
	statusOpt func(*state.Status)
)

func task(opts ...taskOpt) state.Task {
	var t state.Task
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

func statuses(st ...state.Status) taskOpt {
	return func(t *state.Task) {
		t.Statuses = append(t.Statuses, st...)
	}
}

func discovery(dp state.DiscoveryInfo) taskOpt {
	return func(t *state.Task) {
		t.DiscoveryInfo = dp
	}
}

func setLabels(kvs ...string) taskOpt {
	return func(t *state.Task) {
		if len(kvs)%2 != 0 {
			panic("odd number")
		}

		for i := 0; i < len(kvs); i += 2 {
			var label = state.Label{Key: kvs[i], Value: kvs[i+1]}
			log.Errorf("Label1.1 : %v", label)
			t.Labels = append(t.Labels, label)
			log.Errorf("Label1.2 : %v", t.Labels)
		}

	}
}

func status(opts ...statusOpt) state.Status {
	var s state.Status
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func setDiscoveryPort(proto string, port int, name string) state.DiscoveryInfo {

	dp := state.DiscoveryPort{
		Protocol: proto,
		Number:   port,
		Name:     name,
	}

	discoveryPorts := []state.DiscoveryPort{dp}

	ports := state.Ports{
		DiscoveryPorts: discoveryPorts,
	}

	return state.DiscoveryInfo{
		Ports: ports,
	}
}

func setDiscoveryPorts(proto1 string, port1 int, name1 string, proto2 string, port2 int, name2 string) state.DiscoveryInfo {

	dp1 := state.DiscoveryPort{
		Protocol: proto1,
		Number:   port1,
		Name:     name1,
	}

	dp2 := state.DiscoveryPort{
		Protocol: proto2,
		Number:   port2,
		Name:     name2,
	}

	discoveryPorts := []state.DiscoveryPort{dp1, dp2}

	ports := state.Ports{
		DiscoveryPorts: discoveryPorts,
	}

	return state.DiscoveryInfo{
		Ports: ports,
	}
}

func setState(st string) statusOpt {
	return func(s *state.Status) {
		s.State = st
	}
}
func setHealthy(b bool) statusOpt {
	return func(s *state.Status) {
		s.Healthy = &b
	}
}
