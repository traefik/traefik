package mesos

import (
	"testing"

	"github.com/mesos/mesos-go/upid"
	"github.com/mesosphere/mesos-dns/records/state"
)

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
	slave.PID.UPID = &upid.UPID{}
	slave.PID.Host = slave.Hostname

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
