package ecs

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
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
				types.LabelProtocol: aws.String("https"),
			}),
		},
	}

	for i, test := range cases {
		value := test.instanceInfo.Protocol()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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

	for i, test := range cases {
		value := test.instanceInfo.Host()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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

	for i, test := range cases {
		value := test.instanceInfo.Port()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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
				types.LabelWeight: aws.String("10"),
			}),
		},
	}

	for i, test := range cases {
		value := test.instanceInfo.Weight()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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
				types.LabelFrontendPassHostHeader: aws.String("false"),
			}),
		},
	}

	for i, test := range cases {
		value := test.instanceInfo.PassHostHeader()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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
				types.LabelFrontendPriority: aws.String("10"),
			}),
		},
	}

	for i, test := range cases {
		value := test.instanceInfo.Priority()
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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
				types.LabelFrontendEntryPoints: aws.String("http"),
			}),
		},
		{
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendEntryPoints: aws.String("http,https"),
			}),
		},
	}

	for i, test := range cases {
		value := test.instanceInfo.EntryPoints()
		if !reflect.DeepEqual(value, test.expected) {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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
				types.LabelEnable: aws.String("false"),
			}),
		},
		{
			expected:         true,
			exposedByDefault: false,
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelEnable: aws.String("true"),
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

	for i, test := range cases {
		provider := &Provider{
			ExposedByDefault: test.exposedByDefault,
		}
		value := provider.filterInstance(test.instanceInfo)
		if value != test.expected {
			t.Fatalf("Should have been %v, got %v (case %d)", test.expected, value, i)
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

	for _, test := range cases {
		var tasks []*string
		for v := 0; v < test.count; v++ {
			tasks = append(tasks, &testval)
		}

		out := provider.chunkedTaskArns(tasks)
		var outCount []int

		for _, el := range out {
			outCount = append(outCount, len(el))
		}

		if !reflect.DeepEqual(outCount, test.expectedLengths) {
			t.Errorf("Chunking %d elements, expected %#v, got %#v", test.count, test.expectedLengths, outCount)
		}
	}
}

func TestEcsGetBasicAuth(t *testing.T) {
	cases := []struct {
		desc     string
		instance ecsInstance
		expected []string
	}{
		{
			desc:     "label missing",
			instance: simpleEcsInstance(map[string]*string{}),
			expected: []string{},
		},
		{
			desc: "label existing",
			instance: simpleEcsInstance(map[string]*string{
				types.LabelFrontendAuthBasic: aws.String("user:password"),
			}),
			expected: []string{"user:password"},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getBasicAuth(test.instance)
			assert.Equal(t, test.expected, actual)
		})
	}
}
