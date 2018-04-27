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

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() {
			serviceHelperCallCount++
		},
		ListTasksHelper: func(msg events.Message) []swarm.Task {
			taskHelperCallCount++

			return []swarm.Task{}
		},
	}

	c.Execute(events.Message{})

	if serviceHelperCallCount != 1 {
		t.Fatal("expected", 1, "got", serviceHelperCallCount)
	}
	if taskHelperCallCount != 0 {
		t.Fatal("expected", 0, "got", taskHelperCallCount)
	}
}

func TestEventCallbackNoTasksFoundOneRetry(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() {
			serviceHelperCallCount++
		},
		ListTasksHelper: func(msg events.Message) []swarm.Task {
			if msg.Actor.ID != "deadbeef" {
				t.Fatal("expected", "deadbeef", "got", msg.Actor.ID)
			}

			taskHelperCallCount++

			if taskHelperCallCount == 2 {
				task := swarm.Task{}
				task.Status.State = swarm.TaskStateRunning

				return []swarm.Task{task}
			}

			return []swarm.Task{}
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
}

func TestEventCallbackTaskGettingReadyTwoRetries(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() {
			serviceHelperCallCount++
		},
		ListTasksHelper: func(msg events.Message) []swarm.Task {
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

			return []swarm.Task{task}
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
}

func TestEventCallbackTaskGettingReadyFailsAfterTwoRetriesStillExecutesServiceListing(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() {
			serviceHelperCallCount++
		},
		ListTasksHelper: func(msg events.Message) []swarm.Task {
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

			return []swarm.Task{task}
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
}

func TestEventCallbackTaskGettingReadyGoesThroughAllPossibleRetrySteps(t *testing.T) {
	serviceHelperCallCount := 0
	taskHelperCallCount := 0

	c := docker.EventCallback{
		ListAndUpdateServicesHelper: func() {
			serviceHelperCallCount++
		},
		ListTasksHelper: func(msg events.Message) []swarm.Task {
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

			return []swarm.Task{task}
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
}
