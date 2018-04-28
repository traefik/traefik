package docker_test

import (
	"testing"

	"github.com/containous/traefik/provider/docker"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/swarm"
)

func TestEventCallbackNoActorID(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			taskHelperCallCount++

			return []swarm.Task{}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			getServiceHelperCallCount++

			return swarm.Service{}, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	c.Execute(events.Message{})

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackNoTasksFoundOneRetry(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			if taskHelperCallCount == 2 {
				task := swarm.Task{}
				task.Status.State = swarm.TaskStateRunning

				return []swarm.Task{task}, nil
			}

			return []swarm.Task{}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 2 {
		t.Fatal("expected", 2, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 2 {
		t.Fatal("expected", 2, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 3 {
		t.Fatal("expected", 3, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTaskGettingReadyTwoRetries(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			switch taskHelperCallCount {
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
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 3 {
		t.Fatal("expected", 3, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 3 {
		t.Fatal("expected", 3, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 5 {
		t.Fatal("expected", 5, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTaskGettingReadyFailsAfterTwoRetriesStillExecutesServiceListing(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			switch taskHelperCallCount {
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
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 3 {
		t.Fatal("expected", 3, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 3 {
		t.Fatal("expected", 3, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 5 {
		t.Fatal("expected", 5, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTaskGettingReadyGoesThroughAllPossibleRetrySteps(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			switch taskHelperCallCount {
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
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 7 {
		t.Fatal("expected", 7, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 7 {
		t.Fatal("expected", 7, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 13 {
		t.Fatal("expected", 13, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTasksFoundGlobalMode(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{task}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Global = &swarm.GlobalService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTasksFoundFewerThanExpectedOneRetryReplicatedMode(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning
			if taskHelperCallCount == 2 {
				return []swarm.Task{
					task,
					task,
				}, nil
			}

			return []swarm.Task{task}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			replicas := uint64(2)

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: &replicas,
			}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 2 {
		t.Fatal("expected", 2, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 2 {
		t.Fatal("expected", 2, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackNoModeSet(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			return []swarm.Task{}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackReplicatedModeNoNumberOfReplicas(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{task}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepHelperCallCount)
	}
}

func TestEventCallbackTasksFoundMoreThanExpectedNoRetryNeededReplicatedMode(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0
	getServiceHelperCallCount := 0
	sleepHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() error {
			serviceHelperCallCount++

			return nil
		},
		ListTasksHelper: func(msg events.Message) ([]swarm.Task, error) {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			task := swarm.Task{}
			task.Status.State = swarm.TaskStateRunning

			return []swarm.Task{
				task,
				task,
			}, nil
		},
		GetServiceHelper: func(id string) (swarm.Service, error) {
			if id != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", id)
			}

			getServiceHelperCallCount++

			replicas := uint64(1)

			service := swarm.Service{}
			service.ID = id
			service.Spec.Mode.Replicated = &swarm.ReplicatedService{
				Replicas: &replicas,
			}

			return service, nil
		},
		SleepHelper: func() {
			sleepHelperCallCount++
		},
	}

	msg := events.Message{}
	msg.Actor.ID = "deadbeef"
	c.Execute(msg)

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", taskHelperCallCount)
	}
	if getServiceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", getServiceHelperCallCount)
	}
	if sleepHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", sleepHelperCallCount)
	}
}
