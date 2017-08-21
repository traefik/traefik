package marathon

import (
	"strings"
	"time"

	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
)

const testTaskName string = "taskID"

// Functions related to building applications.

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

func label(key, value string) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.AddLabel(key, value)
	}
}

func labelWithService(key, value string, serviceName string) func(*marathon.Application) {
	if len(serviceName) == 0 {
		panic("serviceName can not be empty")
	}

	property := strings.TrimPrefix(key, types.LabelPrefix)
	return func(app *marathon.Application) {
		app.AddLabel(types.LabelPrefix+serviceName+"."+property, value)
	}
}

func healthChecks(checks ...*marathon.HealthCheck) func(*marathon.Application) {
	return func(app *marathon.Application) {
		for _, check := range checks {
			app.AddHealthCheck(*check)
		}
	}
}

func portDefinition(port int) func(*marathon.Application) {
	return func(app *marathon.Application) {
		app.AddPortDefinition(marathon.PortDefinition{
			Port: &port,
		})
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

// Functions related to building tasks.

func task(ops ...func(*marathon.Task)) marathon.Task {
	t := marathon.Task{
		ID: testTaskName,
		// The vast majority of tests expect the task state to be TASK_RUNNING.
		State: string(taskStateRunning),
	}

	for _, op := range ops {
		op(&t)
	}

	return t
}

func localhostTask(ops ...func(*marathon.Task)) marathon.Task {
	t := task(
		host("localhost"),
		ipAddresses("127.0.0.1"),
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

func state(s TaskState) func(*marathon.Task) {
	return func(t *marathon.Task) {
		t.State = string(s)
	}
}

func healthCheckResultLiveness(alive ...bool) func(*marathon.Task) {
	return func(t *marathon.Task) {
		for _, a := range alive {
			t.HealthCheckResults = append(t.HealthCheckResults, &marathon.HealthCheckResult{
				Alive: a,
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
