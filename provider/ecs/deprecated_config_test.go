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

func TestBuildConfigurationV1(t *testing.T) {
	testCases := []struct {
		desc      string
		instances []ecsInstance
		expected  *types.Configuration
		err       error
	}{
		{
			desc: "config parsed successfully",
			instances: []ecsInstance{
				{
					Name: "testing",
					ID:   "1",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{},
					},
					machine: &machine{
						state:     ec2.InstanceStateNameRunning,
						privateIP: "10.0.0.1",
						ports:     []portMapping{{hostPort: 1337}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-testing": {
						Servers: map[string]types.Server{
							"server-testing1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
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
								Rule: "Host:testing.",
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
			instances: []ecsInstance{
				{
					Name: "testing",
					ID:   "1",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{
							label.TraefikBackendHealthCheckPath:     aws.String("/health"),
							label.TraefikBackendHealthCheckInterval: aws.String("1s"),
						}},
					machine: &machine{
						state:     ec2.InstanceStateNameRunning,
						privateIP: "10.0.0.1",
						ports:     []portMapping{{hostPort: 1337}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-testing": {
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "1s",
						},
						Servers: map[string]types.Server{
							"server-testing1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
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
								Rule: "Host:testing.",
							},
						},
						PassHostHeader: true,
						BasicAuth:      []string{},
					},
				},
			},
		},
		{
			desc: "when all labels are set",
			instances: []ecsInstance{
				{
					Name: "testing-instance",
					ID:   "6",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{
							label.TraefikPort:     aws.String("666"),
							label.TraefikProtocol: aws.String("https"),
							label.TraefikWeight:   aws.String("12"),

							label.TraefikBackend: aws.String("foobar"),

							label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
							label.TraefikBackendHealthCheckInterval:              aws.String("6"),
							label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
							label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
							label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
							label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),

							label.TraefikFrontendAuthBasic:      aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
							label.TraefikFrontendEntryPoints:    aws.String("http,https"),
							label.TraefikFrontendPassHostHeader: aws.String("true"),
							label.TraefikFrontendPriority:       aws.String("666"),
							label.TraefikFrontendRule:           aws.String("Host:traefik.io"),
						}},
					machine: &machine{
						state:     ec2.InstanceStateNameRunning,
						privateIP: "10.0.0.1",
						ports:     []portMapping{{hostPort: 1337}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-foobar": {
						Servers: map[string]types.Server{
							"server-testing-instance6": {
								URL:    "https://10.0.0.1:666",
								Weight: 12,
							},
						},
						LoadBalancer: &types.LoadBalancer{
							Method: "drr",
							Sticky: true,
							Stickiness: &types.Stickiness{
								CookieName: "chocolate",
							},
						},
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "6",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-foobar": {
						EntryPoints: []string{
							"http",
							"https",
						},
						Backend: "backend-foobar",
						Routes: map[string]types.Route{
							"route-frontend-foobar": {
								Rule: "Host:traefik.io",
							},
						},
						PassHostHeader: true,
						Priority:       666,
						BasicAuth: []string{
							"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
							"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						},
					},
				},
			},
		}, {
			desc: "Containers with same backend name",
			instances: []ecsInstance{
				{
					Name: "testing-instance",
					ID:   "6",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{
							label.TraefikPort:     aws.String("666"),
							label.TraefikProtocol: aws.String("https"),
							label.TraefikWeight:   aws.String("12"),

							label.TraefikBackend: aws.String("foobar"),

							label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
							label.TraefikBackendHealthCheckInterval:              aws.String("6"),
							label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
							label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
							label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
							label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),

							label.TraefikFrontendAuthBasic:      aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
							label.TraefikFrontendEntryPoints:    aws.String("http,https"),
							label.TraefikFrontendPassHostHeader: aws.String("true"),
							label.TraefikFrontendPriority:       aws.String("666"),
							label.TraefikFrontendRule:           aws.String("Host:traefik.io"),
						}},
					machine: &machine{
						state:     ec2.InstanceStateNameRunning,
						privateIP: "10.0.0.1",
						ports:     []portMapping{{hostPort: 1337}},
					},
				}, {
					Name: "testing-instance-v2",
					ID:   "6",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{
							label.TraefikPort:     aws.String("555"),
							label.TraefikProtocol: aws.String("https"),
							label.TraefikWeight:   aws.String("12"),

							label.TraefikBackend: aws.String("foobar"),

							label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
							label.TraefikBackendHealthCheckInterval:              aws.String("6"),
							label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
							label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
							label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
							label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),

							label.TraefikFrontendAuthBasic:      aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
							label.TraefikFrontendEntryPoints:    aws.String("http,https"),
							label.TraefikFrontendPassHostHeader: aws.String("true"),
							label.TraefikFrontendPriority:       aws.String("666"),
							label.TraefikFrontendRule:           aws.String("Host:traefik.io"),
						}},
					machine: &machine{
						state:     ec2.InstanceStateNameRunning,
						privateIP: "10.0.0.2",
						ports:     []portMapping{{hostPort: 1337}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-foobar": {
						Servers: map[string]types.Server{
							"server-testing-instance6": {
								URL:    "https://10.0.0.1:666",
								Weight: 12,
							},
							"server-testing-instance-v26": {
								URL:    "https://10.0.0.2:555",
								Weight: 12,
							},
						},
						LoadBalancer: &types.LoadBalancer{
							Method: "drr",
							Sticky: true,
							Stickiness: &types.Stickiness{
								CookieName: "chocolate",
							},
						},
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "6",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-foobar": {
						EntryPoints: []string{
							"http",
							"https",
						},
						Backend: "backend-foobar",
						Routes: map[string]types.Route{
							"route-frontend-foobar": {
								Rule: "Host:traefik.io",
							},
						},
						PassHostHeader: true,
						Priority:       666,
						BasicAuth: []string{
							"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
							"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{ExposedByDefault: true}

			instances := fakeLoadTraefikLabels(test.instances)

			got, err := provider.buildConfigurationV1(instances)
			assert.Equal(t, test.err, err) // , err.Error()
			assert.Equal(t, test.expected, got, test.desc)
		})
	}
}

func TestGetFuncStringValueV1(t *testing.T) {
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFuncStringValueV1(label.TraefikProtocol, label.DefaultProtocol)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncSliceStringV1(t *testing.T) {
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFuncSliceStringV1(label.TraefikFrontendEntryPoints)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}
