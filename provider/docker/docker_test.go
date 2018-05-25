package docker_test

import (
	"testing"

	"github.com/containous/traefik/provider/docker"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
)

func TestEventCallbackNoActorID(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			taskFuncCallCount++

			return []swarm.Task{}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			getServiceFuncCallCount++

			return swarm.Service{}, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
	}

	c.Execute(events.Message{})

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackNoTasksFoundOneRetry(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0
	executionFinishedChan := make(chan bool)

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			if taskFuncCallCount == 2 {
				task := swarm.Task{}
				task.Status.State = swarm.TaskStateRunning

				return []swarm.Task{task}, nil
			}

			return []swarm.Task{}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
		ExecutionFinishedChan: executionFinishedChan,
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	<-executionFinishedChan

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 2 {
		t.Fatal("expected", 2, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 2 {
		t.Fatal("expected", 2, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 3 {
		t.Fatal("expected", 3, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTaskGettingReadyTwoRetries(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0
	executionFinishedChan := make(chan bool)

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			switch taskFuncCallCount {
			case 1:
				task.Status.State = swarm.TaskStateNew

				break
			case 2:
				task.Status.State = swarm.TaskStatePreparing

				break
			case 3:
				task.Status.State = swarm.TaskStateRunning

				break
			}

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
		ExecutionFinishedChan: executionFinishedChan,
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	<-executionFinishedChan

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 3 {
		t.Fatal("expected", 3, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 3 {
		t.Fatal("expected", 3, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 5 {
		t.Fatal("expected", 5, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTaskGettingReadyFailsAfterTwoRetriesStillExecutesServiceListing(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0
	executionFinishedChan := make(chan bool)

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			switch taskFuncCallCount {
			case 1:
				task.Status.State = swarm.TaskStateNew

				break
			case 2:
				task.Status.State = swarm.TaskStatePreparing

				break
			case 3:
				task.Status.State = swarm.TaskStateFailed

				break
			}

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
		ExecutionFinishedChan: executionFinishedChan,
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	<-executionFinishedChan

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 3 {
		t.Fatal("expected", 3, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 3 {
		t.Fatal("expected", 3, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 5 {
		t.Fatal("expected", 5, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTaskGettingReadyGoesThroughAllPossibleRetrySteps(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0
	executionFinishedChan := make(chan bool)

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			switch taskFuncCallCount {
			case 1:
				task.Status.State = swarm.TaskStateNew

				break
			case 2:
				task.Status.State = swarm.TaskStatePending

				break
			case 3:
				task.Status.State = swarm.TaskStateAssigned

				break
			case 4:
				task.Status.State = swarm.TaskStateAccepted

				break
			case 5:
				task.Status.State = swarm.TaskStatePreparing

				break
			case 6:
				task.Status.State = swarm.TaskStateStarting

				break
			case 7:
				task.Status.State = swarm.TaskStateRunning

				break
			}

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
		ExecutionFinishedChan: executionFinishedChan,
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	<-executionFinishedChan

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 7 {
		t.Fatal("expected", 7, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 7 {
		t.Fatal("expected", 7, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 13 {
		t.Fatal("expected", 13, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTasksFoundGlobalMode(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTasksFoundFewerThanExpectedOneRetryReplicatedMode(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0
	executionFinishedChan := make(chan bool)

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning
			if taskFuncCallCount == 2 {
				return []swarm.Task{
					task,
					task,
				}, nil
			}

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			replicas := uint64(2)

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: &replicas,
			}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
		ExecutionFinishedChan: executionFinishedChan,
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	<-executionFinishedChan

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 2 {
		t.Fatal("expected", 2, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 2 {
		t.Fatal("expected", 2, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackNoModeSet(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			return []swarm.Task{}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackReplicatedModeNoNumberOfReplicas(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{task}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepFuncCallCount)
	}
}

func TestEventCallbackTasksFoundMoreThanExpectedNoRetryNeededReplicatedMode(t *testing.T) {
	serviceFuncCallCount := 0
	taskFuncCallCount := 0
	getServiceFuncCallCount := 0
	sleepFuncCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesFunc: func() error {
			serviceFuncCallCount++

			return nil
		},
		ListTasksFunc: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskFuncCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{
				task,
				task,
			}, nil
		},
		GetServiceFunc: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceFuncCallCount++

			replicas := uint64(1)

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: &replicas,
			}

			return service, nil
		},
		SleepFunc: func() {
			sleepFuncCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceFuncCallCount)
	}
	if taskFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", taskFuncCallCount)
	}
	if getServiceFuncCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceFuncCallCount)
	}
	if sleepFuncCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepFuncCallCount)
	}
}
