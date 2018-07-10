package marathon

import (
	"strings"
	"time"

	"github.com/containous/traefik/provider/label"
	"github.com/gambol99/go-marathon"
)

const testTaskName = "taskID"

func withAppData(app marathon.Application, segmentName string) appData {
	segmentProperties := label.ExtractTraefikLabels(stringValueMap(app.Labels))
	return appData{
		Application:   app,
		SegmentLabels: segmentProperties[segmentName],
		SegmentName:   segmentName,
		LinkedApps:    nil,
	}
}

// Functions related to building applications.

func withApplications(apps ...marathon.Application) *marathon.Applications {
	return &marathon.Applications{Apps: apps}
}

func application(ops ...func(*marathon.Application)) marathon.Application {
	app := marathon.Application{}
	app.EmptyLabels()
	app.Deployments = []map[string]string{}
	app.ReadinessChecks = &[]marathon.ReadinessCheck{}
	app.ReadinessCheckResults = &[]marathon.ReadinessCheckResult{}

	for _, op := range ops {
		op(&app)
	}

	return app
}

func appID(name string) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.Name(name)
	}
}

func appPorts(ports ...int) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.Ports = append(app.Ports, ports...)
	}
}

func withLabel(key, value string) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.AddLabel(key, value)
	}
}

func constraint(value string) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.AddConstraint(strings.Split(value, ":")...)
	}
}

func withSegmentLabel(key, value string, segmentName string) func(*marathon.Application) {
	if len(segmentName) == 0 {
		panic("segmentName can not be empty")
	}

	property := strings.TrimPrefix(key, label.Prefix)
	return func(app *marathon.Application) {
		app.AddLabel(label.Prefix+segmentName+"."+property, value)
	}
}

func portDefinition(port int) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.AddPortDefinition(marathon.PortDefinition{
			Port: &port,
		})
	}
}

func bridgeNetwork() func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.SetNetwork("bridge", marathon.BridgeNetworkMode)
	}
}

func containerNetwork() func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.SetNetwork("cni", marathon.ContainerNetworkMode)
	}
}

func hostNetwork() func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.SetNetwork("host", marathon.HostNetworkMode)
	}
}

func ipAddrPerTask(port int) func(*marathon.Application) {
	return func(app *marathon.Application) {
		p := marathon.Port{
			Number: port,
			Name:   "port",
		}
		disc := marathon.Discovery{}
		disc.AddPort(p)
		ipAddr := marathon.IPAddressPerTask{}
		ipAddr.SetDiscovery(disc)
		app.SetIPAddressPerTask(ipAddr)
	}
}

func deployments(ids ...string) func(*marathon.Application) {
	return func(app *marathon.Application) {
		for _, id := range ids {
			app.Deployments = append(app.Deployments, map[string]string{
				"ID": id,
			})
		}
	}
}

func readinessCheck(timeout time.Duration) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.ReadinessChecks = &[]marathon.ReadinessCheck{
			{
				Path:           "/ready",
				TimeoutSeconds: int(timeout.Seconds()),
			},
		}
	}
}

func readinessCheckResult(taskID string, ready bool) func(*marathon.Application) {
	return func(app *marathon.Application) {
		*app.ReadinessCheckResults = append(*app.ReadinessCheckResults, marathon.ReadinessCheckResult{
			TaskID: taskID,
			Ready:  ready,
		})
	}
}

func withTasks(tasks ...marathon.Task) func(*marathon.Application) {
	return func(application *marathon.Application) {
		for _, task := range tasks {
			tu := task
			application.Tasks = append(application.Tasks, &tu)
		}
	}
}

// Functions related to building tasks.

func task(ops ...func(*marathon.Task)) marathon.Task {
	t := &marathon.Task{
		ID: testTaskName,
		// The vast majority of tests expect the task state to be TASK_RUNNING.
		State: string(taskStateRunning),
	}

	for _, op := range ops {
		op(t)
	}

	return *t
}

func withTaskID(id string) func(*marathon.Task) {
	return func(task *marathon.Task) {
		task.ID = id
	}
}

func localhostTask(ops ...func(*marathon.Task)) marathon.Task {
	t := task(
		host("localhost"),
		ipAddresses("127.0.0.1"),
		taskState(taskStateRunning),
	)

	for _, op := range ops {
		op(&t)
	}

	return t
}

func taskPorts(ports ...int) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.Ports = append(t.Ports, ports...)
	}
}

func taskState(state TaskState) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.State = string(state)
	}
}

func host(h string) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.Host = h
	}
}

func ipAddresses(addresses ...string) func(*marathon.Task) {
	return func(t *marathon.Task) {
		for _, addr := range addresses {
			t.IPAddresses = append(t.IPAddresses, &marathon.IPAddress{
				IPAddress: addr,
				Protocol:  "tcp",
			})
		}
	}
}

func startedAt(timestamp string) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.StartedAt = timestamp
	}
}

func startedAtFromNow(offset time.Duration) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.StartedAt = time.Now().Add(-offset).Format(time.RFC3339)
	}
}
