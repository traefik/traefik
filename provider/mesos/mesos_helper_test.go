package mesos

import (
	"github.com/containous/traefik/log"
	"github.com/mesosphere/mesos-dns/records/state"
)

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
			log.Debugf("Label1.1 : %v", label)
			t.Labels = append(t.Labels, label)
			log.Debugf("Label1.2 : %v", t.Labels)
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
