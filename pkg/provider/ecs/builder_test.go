package ecs

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func instance(ops ...func(*ecsInstance)) ecsInstance {
	e := &ecsInstance{
		containerDefinition: &ecstypes.ContainerDefinition{},
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

func id(id string) func(*ecsInstance) {
	return func(e *ecsInstance) {
		e.ID = id
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

func mState(state ec2types.InstanceStateName) func(*machine) {
	return func(m *machine) {
		m.state = state
	}
}

func mPrivateIP(ip string) func(*machine) {
	return func(m *machine) {
		m.privateIP = ip
	}
}

func mHealthStatus(status ecstypes.HealthStatus) func(*machine) {
	return func(m *machine) {
		m.healthStatus = status
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

func mPort(containerPort, hostPort int32, protocol ecstypes.TransportProtocol) func(*portMapping) {
	return func(pm *portMapping) {
		pm.containerPort = containerPort
		pm.hostPort = hostPort
		pm.protocol = protocol
	}
}

func labels(labels map[string]string) func(*ecsInstance) {
	return func(c *ecsInstance) {
		c.Labels = labels
	}
}
