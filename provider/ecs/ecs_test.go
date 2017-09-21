package ecs

import (
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
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Protocol label is not set should return a string equals to http",
			expected:     "http",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
		{
			desc:     "Protocol label is set to http should return a string equals to http",
			expected: "http",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelProtocol: aws.String("http"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Protocol label is set to https should return a string equals to https",
			expected: "https",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelProtocol: aws.String("https"),
			}),
			provider: &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getProtocol(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsHost(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Default host should be 10.0.0.0",
			expected:     "10.0.0.0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getHost(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsPort(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Default port should be 80",
			expected:     "80",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getPort(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsWeight(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Weight label not set should return a string equals to 0",
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
		{
			desc:     "Weight label set 0 should return a string equals to 0",
			expected: "0",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelWeight: aws.String("0"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Weight label set -1 should return a string equals to -1",
			expected: "-1",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelWeight: aws.String("-1"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Weight label set 10 should return a string equals to 10",
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelWeight: aws.String("10"),
			}),
			provider: &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getWeight(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsPassHostHeader(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Frontend pass host header label not set should return a string equals to true",
			expected:     "true",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
		{
			desc:     "Frontend pass host header label set to false should return a string equals to false",
			expected: "false",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendPassHostHeader: aws.String("false"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Frontend pass host header label set to true should return a string equals to true",
			expected: "true",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendPassHostHeader: aws.String("true"),
			}),
			provider: &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getPassHostHeader(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsPriority(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Frontend priority label not set should return a string equals to 0",
			expected:     "0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
		{
			desc:     "Frontend priority label set to 10 should return a string equals to 10",
			expected: "10",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendPriority: aws.String("10"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Frontend priority label set to -1 should return a string equals to -1",
			expected: "-1",
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendPriority: aws.String("-1"),
			}),
			provider: &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getPriority(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestEcsEntryPoints(t *testing.T) {
	tests := []struct {
		desc         string
		expected     []string
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Frontend entrypoints label not set should return empty array",
			expected:     []string{},
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider:     &Provider{},
		},
		{
			desc:     "Frontend entrypoints label set to http should return a string array of 1 element",
			expected: []string{"http"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendEntryPoints: aws.String("http"),
			}),
			provider: &Provider{},
		},
		{
			desc:     "Frontend entrypoints label set to http,https should return a string array of 2 elements",
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelFrontendEntryPoints: aws.String("http,https"),
			}),
			provider: &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.getEntryPoints(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
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

	tests := []struct {
		desc         string
		expected     bool
		instanceInfo ecsInstance
		provider     *Provider
	}{
		{
			desc:         "Instance without enable label and exposed by default enabled should be not filtered",
			expected:     true,
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:         "Instance without enable label and exposed by default disabled should be filtered",
			expected:     false,
			instanceInfo: simpleEcsInstance(map[string]*string{}),
			provider: &Provider{
				ExposedByDefault: false,
			},
		},
		{
			desc:     "Instance with enable label set to false and exposed by default enabled should be filtered",
			expected: false,
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelEnable: aws.String("false"),
			}),
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:     "Instance with enable label set to true and exposed by default disabled should be not filtered",
			expected: true,
			instanceInfo: simpleEcsInstance(map[string]*string{
				types.LabelEnable: aws.String("true"),
			}),
			provider: &Provider{
				ExposedByDefault: false,
			},
		},
		{
			desc:         "Instance with nil private ip and exposed by default enabled should be filtered",
			expected:     false,
			instanceInfo: nilPrivateIP,
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:         "Instance with nil machine and exposed by default enabled should be filtered",
			expected:     false,
			instanceInfo: nilMachine,
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:         "Instance with nil machine state and exposed by default enabled should be filtered",
			expected:     false,
			instanceInfo: nilMachineState,
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:         "Instance with nil machine state name and exposed by default enabled should be filtered",
			expected:     false,
			instanceInfo: nilMachineStateName,
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
		{
			desc:         "Instance with invalid machine state and exposed by default enabled should be filtered",
			expected:     false,
			instanceInfo: invalidMachineState,
			provider: &Provider{
				ExposedByDefault: true,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.provider.filterInstance(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestTaskChunking(t *testing.T) {
	testval := "a"
	tests := []struct {
		desc            string
		count           int
		expectedLengths []int
		provider        *Provider
	}{
		{
			desc:            "0 parameter should return nil",
			count:           0,
			expectedLengths: []int(nil),
			provider:        &Provider{},
		},
		{
			desc:            "1 parameter should return 1 array of 1 element",
			count:           1,
			expectedLengths: []int{1},
			provider:        &Provider{},
		},
		{
			desc:            "99 parameters should return 1 array of 99 elements",
			count:           99,
			expectedLengths: []int{99},
			provider:        &Provider{},
		},
		{
			desc:            "100 parameters should return 1 array of 100 elements",
			count:           100,
			expectedLengths: []int{100},
			provider:        &Provider{},
		},
		{
			desc:            "101 parameters should return 1 array of 100 elements and 1 array of 1 element",
			count:           101,
			expectedLengths: []int{100, 1},
			provider:        &Provider{},
		},
		{
			desc:            "199 parameters should return 1 array of 100 elements and 1 array of 99 elements",
			count:           199,
			expectedLengths: []int{100, 99},
			provider:        &Provider{},
		},
		{
			desc:            "200 parameters should return 2 arrays of 100 elements each",
			count:           200,
			expectedLengths: []int{100, 100},
			provider:        &Provider{},
		},
		{
			desc:            "201 parameters should return 2 arrays of 100 elements each and 1 array of 1 element",
			count:           201,
			expectedLengths: []int{100, 100, 1},
			provider:        &Provider{},
		},
		{
			desc:            "555 parameters should return 5 arrays of 100 elements each and 1 array of 55 elements",
			count:           555,
			expectedLengths: []int{100, 100, 100, 100, 100, 55},
			provider:        &Provider{},
		},
		{
			desc:            "1001 parameters should return 10 arrays of 100 elements each and 1 array of 1 element",
			count:           1001,
			expectedLengths: []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 1},
			provider:        &Provider{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			var tasks []*string
			for v := 0; v < test.count; v++ {
				tasks = append(tasks, &testval)
			}

			out := test.provider.chunkedTaskArns(tasks)
			var outCount []int

			for _, el := range out {
				outCount = append(outCount, len(el))
			}

			assert.Equal(t, test.expectedLengths, outCount, "Chunking %d elements", test.count)
		})

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
