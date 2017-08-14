package marathon

import "github.com/gambol99/go-marathon"

// Functions related to building applications.

func createApplication(ops ...func(*marathon.Application)) marathon.Application {
	app := marathon.Application{}
	app.EmptyLabels()

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

// Functions related to building tasks.

func createTask(ops ...func(*marathon.Task)) marathon.Task {
	t := marathon.Task{
		// The vast majority of tests expect the task state to be TASK_RUNNING.
		State: string(taskStateRunning),
	}

	for _, op := range ops {
		op(&t)
	}

	return t
}

func createLocalhostTask(ops ...func(*marathon.Task)) marathon.Task {
	t := createTask(
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
