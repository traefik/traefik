package ecs

import (
	"github.com/aws/aws-sdk-go/service/ecs"
)

func instance(ops ...func(*ecsInstance)) ecsInstance {
	e := &ecsInstance{
		containerDefinition: &ecs.ContainerDefinition{},
	}

	for _, op := range ops {
		op(e)
	}

	return *e
}

func name(name string) func(*ecsInstance) {
	return func(e *ecsInstance) {
		e.Name = name
	}
}

func ID(ID string) func(*ecsInstance) {
	return func(e *ecsInstance) {
		e.ID = ID
	}
}

func iMachine(opts ...func(*machine)) func(*ecsInstance) {
	return func(e *ecsInstance) {
		e.machine = &machine{}

		for _, opt := range opts {
			opt(e.machine)
		}
	}
}

func mState(state string) func(*machine) {
	return func(m *machine) {
		m.state = state
	}
}

func mName(name string) func(*machine) {
	return func(m *machine) {
		m.name = name
	}
}
func mPrivateIP(ip string) func(*machine) {
	return func(m *machine) {
		m.privateIP = ip
	}
}

func mPorts(opts ...func(*portMapping)) func(*machine) {
	return func(m *machine) {
		for _, opt := range opts {
			p := &portMapping{}
			opt(p)
			m.ports = append(m.ports, *p)
		}
	}
}

func mPort(containerPort int32, hostPort int32) func(*portMapping) {
	return func(pm *portMapping) {
		pm.containerPort = int64(containerPort)
		pm.hostPort = int64(hostPort)
	}
}

func labels(labels map[string]string) func(*ecsInstance) {
	return func(c *ecsInstance) {
		c.TraefikLabels = labels
	}
}

func dockerLabels(labels map[string]*string) func(*ecsInstance) {
	return func(c *ecsInstance) {
		c.containerDefinition.DockerLabels = labels
	}
}
