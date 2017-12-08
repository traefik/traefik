package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildConfiguration(t *testing.T) {
	provider := &Provider{}
	tests := []struct {
		desc     string
		services map[string][]ecsInstance
		expected *types.Configuration
		err      error
	}{
		{
			desc: "config parsed successfully",
			services: map[string][]ecsInstance{
				"testing": {
					{
						Name: "instance-1",
						containerDefinition: &ecs.ContainerDefinition{
							DockerLabels: map[string]*string{},
						},
						machine: &ec2.Instance{
							PrivateIpAddress: func(s string) *string { return &s }("10.0.0.1"),
						},
						container: &ecs.Container{
							NetworkBindings: []*ecs.NetworkBinding{
								{
									HostPort: func(i int64) *int64 { return &i }(1337),
								},
							},
						},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance-1": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL: "http://10.0.0.1:1337",
							},
						},
					},
					"backend-testing": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-testing": {
						EntryPoints: []string{},
						Backend:     "backend-testing",
						Routes: map[string]types.Route{
							"route-frontend-testing": {
								Rule: "Host:instance-1.",
							},
						},
						PassHostHeader: true,
						BasicAuth:      []string{},
					},
				},
			},
		},
		{
			desc: "config parsed successfully with health check labels",
			services: map[string][]ecsInstance{
				"testing": {
					{
						Name: "instance-1",
						containerDefinition: &ecs.ContainerDefinition{
							DockerLabels: map[string]*string{
								label.TraefikBackendHealthCheckPath:     func(s string) *string { return &s }("/health"),
								label.TraefikBackendHealthCheckInterval: func(s string) *string { return &s }("1s"),
							},
						},
						machine: &ec2.Instance{
							PrivateIpAddress: func(s string) *string { return &s }("10.0.0.1"),
						},
						container: &ecs.Container{
							NetworkBindings: []*ecs.NetworkBinding{
								{
									HostPort: func(i int64) *int64 { return &i }(1337),
								},
							},
						},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance-1": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL: "http://10.0.0.1:1337",
							},
						},
					},
					"backend-testing": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "1s",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-testing": {
						EntryPoints: []string{},
						Backend:     "backend-testing",
						Routes: map[string]types.Route{
							"route-frontend-testing": {
								Rule: "Host:instance-1.",
							},
						},
						PassHostHeader: true,
						BasicAuth:      []string{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got, err := provider.buildConfiguration(test.services)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, got)
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

	noNetwork := simpleEcsInstanceNoNetwork(map[string]*string{})

	noNetworkWithLabel := simpleEcsInstanceNoNetwork(map[string]*string{
		label.TraefikPort: aws.String("80"),
	})

	tests := []struct {
		desc             string
		instanceInfo     ecsInstance
		exposedByDefault bool
		expected         bool
	}{
		{
			desc:             "Instance without enable label and exposed by default enabled should be not filtered",
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
			exposedByDefault: true,
			expected:         true,
		},
		{
			desc:             "Instance without enable label and exposed by default disabled should be filtered",
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "Instance with enable label set to false and exposed by default enabled should be filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikEnable: aws.String("false"),
			}),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with enable label set to true and exposed by default disabled should be not filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikEnable: aws.String("true"),
			}),
			exposedByDefault: false,
			expected:         true,
		},
		{
			desc:             "Instance with nil private ip and exposed by default enabled should be filtered",
			instanceInfo:     nilPrivateIP,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with nil machine and exposed by default enabled should be filtered",
			instanceInfo:     nilMachine,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with nil machine state and exposed by default enabled should be filtered",
			instanceInfo:     nilMachineState,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with nil machine state name and exposed by default enabled should be filtered",
			instanceInfo:     nilMachineStateName,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with invalid machine state and exposed by default enabled should be filtered",
			instanceInfo:     invalidMachineState,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with no port mappings should be filtered",
			instanceInfo:     noNetwork,
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with no port mapping and with label should not be filtered",
			instanceInfo:     noNetworkWithLabel,
			exposedByDefault: true,
			expected:         true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			prov := &Provider{
				ExposedByDefault: test.exposedByDefault,
			}
			actual := prov.filterInstance(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestChunkedTaskArns(t *testing.T) {
	testVal := "a"
	tests := []struct {
		desc            string
		count           int
		expectedLengths []int
	}{
		{
			desc:            "0 parameter should return nil",
			count:           0,
			expectedLengths: []int(nil),
		},
		{
			desc:            "1 parameter should return 1 array of 1 element",
			count:           1,
			expectedLengths: []int{1},
		},
		{
			desc:            "99 parameters should return 1 array of 99 elements",
			count:           99,
			expectedLengths: []int{99},
		},
		{
			desc:            "100 parameters should return 1 array of 100 elements",
			count:           100,
			expectedLengths: []int{100},
		},
		{
			desc:            "101 parameters should return 1 array of 100 elements and 1 array of 1 element",
			count:           101,
			expectedLengths: []int{100, 1},
		},
		{
			desc:            "199 parameters should return 1 array of 100 elements and 1 array of 99 elements",
			count:           199,
			expectedLengths: []int{100, 99},
		},
		{
			desc:            "200 parameters should return 2 arrays of 100 elements each",
			count:           200,
			expectedLengths: []int{100, 100},
		},
		{
			desc:            "201 parameters should return 2 arrays of 100 elements each and 1 array of 1 element",
			count:           201,
			expectedLengths: []int{100, 100, 1},
		},
		{
			desc:            "555 parameters should return 5 arrays of 100 elements each and 1 array of 55 elements",
			count:           555,
			expectedLengths: []int{100, 100, 100, 100, 100, 55},
		},
		{
			desc:            "1001 parameters should return 10 arrays of 100 elements each and 1 array of 1 element",
			count:           1001,
			expectedLengths: []int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 1},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			var tasks []*string
			for v := 0; v < test.count; v++ {
				tasks = append(tasks, &testVal)
			}

			out := chunkedTaskArns(tasks)
			var outCount []int

			for _, el := range out {
				outCount = append(outCount, len(el))
			}

			assert.Equal(t, test.expectedLengths, outCount, "Chunking %d elements", test.count)
		})

	}
}

func TestGetHost(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Default host should be 10.0.0.0",
			expected:     "10.0.0.0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := getHost(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPort(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Default port should be 80",
			expected:     "80",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Label should override network port",
			expected: "4242",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikPort: aws.String("4242"),
			}),
		},
		{
			desc:     "Label should provide exposed port",
			expected: "80",
			instanceInfo: simpleEcsInstanceNoNetwork(map[string]*string{
				label.TraefikPort: aws.String("80"),
			}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := getPort(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncStringValue(t *testing.T) {
	tests := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Protocol label is not set should return a string equals to http",
			expected:     "http",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Protocol label is set to http should return a string equals to http",
			expected: "http",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikProtocol: aws.String("http"),
			}),
		},
		{
			desc:     "Protocol label is set to https should return a string equals to https",
			expected: "https",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikProtocol: aws.String("https"),
			}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFuncStringValue(label.TraefikProtocol, label.DefaultProtocol)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncSliceString(t *testing.T) {
	tests := []struct {
		desc         string
		expected     []string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Frontend entrypoints label not set should return empty array",
			expected:     nil,
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Frontend entrypoints label set to http should return a string array of 1 element",
			expected: []string{"http"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikFrontendEntryPoints: aws.String("http"),
			}),
		},
		{
			desc:     "Frontend entrypoints label set to http,https should return a string array of 2 elements",
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikFrontendEntryPoints: aws.String("http,https"),
			}),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := getFuncSliceString(label.TraefikFrontendEntryPoints)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

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

func simpleEcsInstanceNoNetwork(labels map[string]*string) ecsInstance {
	return makeEcsInstance(&ecs.ContainerDefinition{
		Name:         aws.String("http"),
		PortMappings: []*ecs.PortMapping{},
		DockerLabels: labels,
	})
}
