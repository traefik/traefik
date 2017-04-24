package ecs

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func makeEcsInstance(containerDef *ecs.ContainerDefinition) ecsInstance {
	container := &ecs.Container{
		Name:            containerDef.Name,
		NetworkBindings: make([]*ecs.NetworkBinding, len(containerDef.PortMappings)),
	}

	for i, pm := range containerDef.PortMappings {
		container.NetworkBindings[i] = &ecs.NetworkBinding{
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
			Protocol:      pm.Protocol,
			BindIP:        aws.String("0.0.0.0"),
		}
	}

	return ecsInstance{
		Name: "foo-http",
		ID:   "123456789abc",
		task: &ecs.Task{
			Containers: []*ecs.Container{container},
		},
		taskDefinition: &ecs.TaskDefinition{
			ContainerDefinitions: []*ecs.ContainerDefinition{containerDef},
		},
		container:           container,
		containerDefinition: containerDef,
		machine: &ec2.Instance{
			PrivateIpAddress: aws.String("10.0.0.0"),
			State: &ec2.InstanceState{
				Name: aws.String(ec2.InstanceStateNameRunning),
			},
		},
	}
}

func simpleEcsInstance(labels map[string]*string) ecsInstance {
	return makeEcsInstance(&ecs.ContainerDefinition{
		Name: aws.String("http"),
		PortMappings: []*ecs.PortMapping{{
			HostPort:      aws.Int64(80),
			ContainerPort: aws.Int64(80),
			Protocol:      aws.String("tcp"),
		}},
		DockerLabels: labels,
	})
}

func TestEcsProtocol(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "http",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "https",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.protocol": aws.String("https"),
			}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.Protocol()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsHost(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "10.0.0.0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.Host()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsPort(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "80",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.Port()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsWeight(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.weight": aws.String("10"),
			}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.Weight()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsPassHostHeader(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "true",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "false",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.passHostHeader": aws.String("false"),
			}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.PassHostHeader()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsPriority(t *testing.T) {
	cases := []struct {
		expected     string
		instanceInfo ecsInstance
	}{
		{
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.priority": aws.String("10"),
			}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.Priority()
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestEcsEntryPoints(t *testing.T) {
	cases := []struct {
		expected     []string
		instanceInfo ecsInstance
	}{
		{
			expected:     []string{},
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			expected: []string{"http"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.entryPoints": aws.String("http"),
			}),
		},
		{
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.frontend.entryPoints": aws.String("http,https"),
			}),
		},
	}

	for i, c := range cases {
		value := c.instanceInfo.EntryPoints()
		if !reflect.DeepEqual(value, c.expected) {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestFilterInstance(t *testing.T) {

	nilPrivateIP := simpleEcsInstance(map[string]*string{})
	nilPrivateIP.machine.PrivateIpAddress = nil

	nilMachine := simpleEcsInstance(map[string]*string{})
	nilMachine.machine = nil

	nilMachineState := simpleEcsInstance(map[string]*string{})
	nilMachineState.machine.State = nil

	nilMachineStateName := simpleEcsInstance(map[string]*string{})
	nilMachineStateName.machine.State.Name = nil

	invalidMachineState := simpleEcsInstance(map[string]*string{})
	invalidMachineState.machine.State.Name = aws.String(ec2.InstanceStateNameStopped)

	cases := []struct {
		expected         bool
		exposedByDefault bool
		instanceInfo     ecsInstance
	}{
		{
			expected:         true,
			exposedByDefault: true,
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
		},
		{
			expected:         false,
			exposedByDefault: false,
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.enable": aws.String("false"),
			}),
		},
		{
			expected:         true,
			exposedByDefault: false,
			instanceInfo: simpleEcsInstance(map[string]*string{
				"traefik.enable": aws.String("true"),
			}),
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo:     nilPrivateIP,
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo:     nilMachine,
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo:     nilMachineState,
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo:     nilMachineStateName,
		},
		{
			expected:         false,
			exposedByDefault: true,
			instanceInfo:     invalidMachineState,
		},
	}

	for i, c := range cases {
		provider := &Provider{
			ExposedByDefault: c.exposedByDefault,
		}
		value := provider.filterInstance(c.instanceInfo)
		if value != c.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", c.expected, value, i)
		}
	}
}

func TestTaskChunking(t *testing.T) {
	provider := &Provider{}

	testval := "a"
	cases := []struct {
		count           int
		expectedLengths []int
	}{
		{0, []int(nil)},
		{1, []int{1}},
		{99, []int{99}},
		{100, []int{100}},
		{101, []int{100, 1}},
		{199, []int{100, 99}},
		{200, []int{100, 100}},
		{201, []int{100, 100, 1}},
		{555, []int{100, 100, 100, 100, 100, 55}},
		{1001, []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 1}},
	}

	for _, c := range cases {
		var tasks []*string
		for v := 0; v < c.count; v++ {
			tasks = append(tasks, &testval)
		}

		out := provider.chunkedTaskArns(tasks)
		var outCount []int

		for _, el := range out {
			outCount = append(outCount, len(el))
		}

		if !reflect.DeepEqual(outCount, c.expectedLengths) {
			t.Errorf("Chunking %d elements, expected %#v, got %#v", c.count, c.expectedLengths, outCount)
		}
	}
}
