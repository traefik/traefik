package mesos

import (
	"strings"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/stretchr/testify/assert"
)

// test helpers

func TestBuilder(t *testing.T) {
	result := aTask("ID1",
		withIP("10.10.10.10"),
		withLabel("foo", "bar"),
		withLabel("fii", "bar"),
		withLabel("fuu", "bar"),
		withInfo("name1",
			withPorts(withPort("TCP", 80, "p"),
				withPortTCP(81, "n"))),
		withStatus(withHealthy(true), withState("a")))

	expected := state.Task{
		FrameworkID: "",
		ID:          "ID1",
		SlaveIP:     "10.10.10.10",
		Name:        "",
		SlaveID:     "",
		State:       "",
		Statuses: []state.Status{{
			State:           "a",
			Healthy:         Bool(true),
			ContainerStatus: state.ContainerStatus{},
		}},
		DiscoveryInfo: state.DiscoveryInfo{
			Name: "name1",
			Labels: struct {
				Labels []state.Label `json:"labels"`
			}{},
			Ports: state.Ports{DiscoveryPorts: []state.DiscoveryPort{
				{Protocol: "TCP", Number: 80, Name: "p"},
				{Protocol: "TCP", Number: 81, Name: "n"}}}},
		Labels: []state.Label{
			{Key: "foo", Value: "bar"},
			{Key: "fii", Value: "bar"},
			{Key: "fuu", Value: "bar"}}}

	assert.Equal(t, expected, result)
}

func aTaskData(id, segment string, ops ...func(*state.Task)) taskData {
	ts := &state.Task{ID: id}
	for _, op := range ops {
		op(ts)
	}
	lbls := label.ExtractTraefikLabels(extractLabels(*ts))
	if len(lbls[segment]) > 0 {
		return taskData{Task: *ts, TraefikLabels: lbls[segment], SegmentName: segment}
	}
	return taskData{Task: *ts, TraefikLabels: lbls[""], SegmentName: segment}
}

func segmentedTaskData(segments []string, ts state.Task) []taskData {
	var td []taskData
	lbls := label.ExtractTraefikLabels(extractLabels(ts))
	for _, s := range segments {
		if l, ok := lbls[s]; !ok {
			td = append(td, taskData{Task: ts, TraefikLabels: lbls[""], SegmentName: s})
		} else {
			td = append(td, taskData{Task: ts, TraefikLabels: l, SegmentName: s})
		}
	}
	return td
}

func aTask(id string, ops ...func(*state.Task)) state.Task {
	ts := &state.Task{ID: id}
	for _, op := range ops {
		op(ts)
	}
	return *ts
}

func withIP(ip string) func(*state.Task) {
	return func(task *state.Task) {
		task.SlaveIP = ip
	}
}

func withInfo(name string, ops ...func(*state.DiscoveryInfo)) func(*state.Task) {
	return func(task *state.Task) {
		info := &state.DiscoveryInfo{Name: name}
		for _, op := range ops {
			op(info)
		}
		task.DiscoveryInfo = *info
	}
}

func withPorts(ops ...func(port *state.DiscoveryPort)) func(*state.DiscoveryInfo) {
	return func(info *state.DiscoveryInfo) {
		var ports []state.DiscoveryPort
		for _, op := range ops {
			pt := &state.DiscoveryPort{}
			op(pt)
			ports = append(ports, *pt)
		}

		info.Ports = state.Ports{
			DiscoveryPorts: ports,
		}
	}
}

func withPort(proto string, port int, name string) func(port *state.DiscoveryPort) {
	return func(p *state.DiscoveryPort) {
		p.Protocol = proto
		p.Number = port
		p.Name = name
	}
}

func withPortTCP(port int, name string) func(port *state.DiscoveryPort) {
	return withPort("TCP", port, name)
}

func withStatus(ops ...func(*state.Status)) func(*state.Task) {
	return func(task *state.Task) {
		st := &state.Status{}
		for _, op := range ops {
			op(st)
		}
		task.Statuses = append(task.Statuses, *st)
	}
}
func withDefaultStatus(ops ...func(*state.Status)) func(*state.Task) {
	return func(task *state.Task) {
		for _, op := range ops {
			st := &state.Status{
				State:   "TASK_RUNNING",
				Healthy: Bool(true),
			}
			op(st)
			task.Statuses = append(task.Statuses, *st)
		}
	}
}

func withHealthy(st bool) func(*state.Status) {
	return func(status *state.Status) {
		status.Healthy = Bool(st)
	}
}

func withState(st string) func(*state.Status) {
	return func(status *state.Status) {
		status.State = st
	}
}

func withLabel(key, value string) func(*state.Task) {
	return func(task *state.Task) {
		lbl := state.Label{Key: key, Value: value}
		task.Labels = append(task.Labels, lbl)
	}
}

func withSegmentLabel(key, value, segmentName string) func(*state.Task) {
	if len(segmentName) == 0 {
		panic("segmentName can not be empty")
	}

	property := strings.TrimPrefix(key, label.Prefix)
	return func(task *state.Task) {
		lbl := state.Label{Key: label.Prefix + segmentName + "." + property, Value: value}
		task.Labels = append(task.Labels, lbl)
	}
}

func Bool(v bool) *bool {
	return &v
}
