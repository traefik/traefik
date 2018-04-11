package ecs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildConfiguration(t *testing.T) {
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
					Name: "instance",
					ID:   "1",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{},
					},
					machine: &ec2.Instance{
						State: &ec2.InstanceState{
							Name: aws.String(ec2.InstanceStateNameRunning),
						},
						PrivateIpAddress: aws.String("10.0.0.1"),
					},
					container: &ecs.Container{
						NetworkBindings: []*ecs.NetworkBinding{{
							HostPort: aws.Int64(1337),
						}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL: "http://10.0.0.1:1337",
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance.",
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
					Name: "instance",
					ID:   "1",
					containerDefinition: &ecs.ContainerDefinition{
						DockerLabels: map[string]*string{
							label.TraefikBackendHealthCheckPath:     aws.String("/health"),
							label.TraefikBackendHealthCheckInterval: aws.String("1s"),
						}},
					machine: &ec2.Instance{
						State: &ec2.InstanceState{
							Name: aws.String(ec2.InstanceStateNameRunning),
						},
						PrivateIpAddress: aws.String("10.0.0.1"),
					},
					container: &ecs.Container{
						NetworkBindings: []*ecs.NetworkBinding{{
							HostPort: aws.Int64(1337),
						}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "1s",
						},
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL: "http://10.0.0.1:1337",
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance.",
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

							label.TraefikBackendCircuitBreakerExpression:         aws.String("NetworkErrorRatio() > 0.5"),
							label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
							label.TraefikBackendHealthCheckPort:                  aws.String("880"),
							label.TraefikBackendHealthCheckInterval:              aws.String("6"),
							label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
							label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
							label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
							label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),
							label.TraefikBackendMaxConnAmount:                    aws.String("666"),
							label.TraefikBackendMaxConnExtractorFunc:             aws.String("client.ip"),
							label.TraefikBackendBufferingMaxResponseBodyBytes:    aws.String("10485760"),
							label.TraefikBackendBufferingMemResponseBodyBytes:    aws.String("2097152"),
							label.TraefikBackendBufferingMaxRequestBodyBytes:     aws.String("10485760"),
							label.TraefikBackendBufferingMemRequestBodyBytes:     aws.String("2097152"),
							label.TraefikBackendBufferingRetryExpression:         aws.String("IsNetworkError() && Attempts() <= 2"),

							label.TraefikFrontendAuthBasic:                 aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
							label.TraefikFrontendEntryPoints:               aws.String("http,https"),
							label.TraefikFrontendPassHostHeader:            aws.String("true"),
							label.TraefikFrontendPassTLSCert:               aws.String("true"),
							label.TraefikFrontendPriority:                  aws.String("666"),
							label.TraefikFrontendRedirectEntryPoint:        aws.String("https"),
							label.TraefikFrontendRedirectRegex:             aws.String("nope"),
							label.TraefikFrontendRedirectReplacement:       aws.String("nope"),
							label.TraefikFrontendRedirectPermanent:         aws.String("true"),
							label.TraefikFrontendRule:                      aws.String("Host:traefik.io"),
							label.TraefikFrontendWhiteListSourceRange:      aws.String("10.10.10.10"),
							label.TraefikFrontendWhiteListUseXForwardedFor: aws.String("true"),

							label.TraefikFrontendRequestHeaders:          aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
							label.TraefikFrontendResponseHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
							label.TraefikFrontendSSLProxyHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
							label.TraefikFrontendAllowedHosts:            aws.String("foo,bar,bor"),
							label.TraefikFrontendHostsProxyHeaders:       aws.String("foo,bar,bor"),
							label.TraefikFrontendSSLHost:                 aws.String("foo"),
							label.TraefikFrontendCustomFrameOptionsValue: aws.String("foo"),
							label.TraefikFrontendContentSecurityPolicy:   aws.String("foo"),
							label.TraefikFrontendPublicKey:               aws.String("foo"),
							label.TraefikFrontendReferrerPolicy:          aws.String("foo"),
							label.TraefikFrontendCustomBrowserXSSValue:   aws.String("foo"),
							label.TraefikFrontendSTSSeconds:              aws.String("666"),
							label.TraefikFrontendSSLRedirect:             aws.String("true"),
							label.TraefikFrontendSSLTemporaryRedirect:    aws.String("true"),
							label.TraefikFrontendSTSIncludeSubdomains:    aws.String("true"),
							label.TraefikFrontendSTSPreload:              aws.String("true"),
							label.TraefikFrontendForceSTSHeader:          aws.String("true"),
							label.TraefikFrontendFrameDeny:               aws.String("true"),
							label.TraefikFrontendContentTypeNosniff:      aws.String("true"),
							label.TraefikFrontendBrowserXSSFilter:        aws.String("true"),
							label.TraefikFrontendIsDevelopment:           aws.String("true"),

							label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  aws.String("404"),
							label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: aws.String("foobar"),
							label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   aws.String("foo_query"),
							label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  aws.String("500,600"),
							label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: aws.String("foobar"),
							label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   aws.String("bar_query"),

							label.TraefikFrontendRateLimitExtractorFunc:                                        aws.String("client.ip"),
							label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  aws.String("6"),
							label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: aws.String("12"),
							label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   aws.String("18"),
							label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  aws.String("3"),
							label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: aws.String("6"),
							label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   aws.String("9"),
						}},
					machine: &ec2.Instance{
						State: &ec2.InstanceState{
							Name: aws.String(ec2.InstanceStateNameRunning),
						},
						PrivateIpAddress: aws.String("10.0.0.1"),
					},
					container: &ecs.Container{
						NetworkBindings: []*ecs.NetworkBinding{{
							HostPort: aws.Int64(1337),
						}},
					},
				},
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-testing-instance": {
						Servers: map[string]types.Server{
							"server-testing-instance-6": {
								URL:    "https://10.0.0.1:666",
								Weight: 12,
							},
						},
						CircuitBreaker: &types.CircuitBreaker{
							Expression: "NetworkErrorRatio() > 0.5",
						},
						LoadBalancer: &types.LoadBalancer{
							Method: "drr",
							Sticky: true,
							Stickiness: &types.Stickiness{
								CookieName: "chocolate",
							},
						},
						MaxConn: &types.MaxConn{
							Amount:        666,
							ExtractorFunc: "client.ip",
						},
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Port:     880,
							Interval: "6",
						},
						Buffering: &types.Buffering{
							MaxResponseBodyBytes: 10485760,
							MemResponseBodyBytes: 2097152,
							MaxRequestBodyBytes:  10485760,
							MemRequestBodyBytes:  2097152,
							RetryExpression:      "IsNetworkError() && Attempts() <= 2",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-testing-instance": {
						EntryPoints: []string{
							"http",
							"https",
						},
						Backend: "backend-testing-instance",
						Routes: map[string]types.Route{
							"route-frontend-testing-instance": {
								Rule: "Host:traefik.io",
							},
						},
						PassHostHeader: true,
						PassTLSCert:    true,
						Priority:       666,
						BasicAuth: []string{
							"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
							"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						},
						WhiteList: &types.WhiteList{
							SourceRange:      []string{"10.10.10.10"},
							UseXForwardedFor: true,
						},
						Headers: &types.Headers{
							CustomRequestHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
							},
							CustomResponseHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
							},
							AllowedHosts: []string{
								"foo",
								"bar",
								"bor",
							},
							HostsProxyHeaders: []string{
								"foo",
								"bar",
								"bor",
							},
							SSLRedirect:          true,
							SSLTemporaryRedirect: true,
							SSLHost:              "foo",
							SSLProxyHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
							},
							STSSeconds:              666,
							STSIncludeSubdomains:    true,
							STSPreload:              true,
							ForceSTSHeader:          true,
							FrameDeny:               true,
							CustomFrameOptionsValue: "foo",
							ContentTypeNosniff:      true,
							BrowserXSSFilter:        true,
							CustomBrowserXSSValue:   "foo",
							ContentSecurityPolicy:   "foo",
							PublicKey:               "foo",
							ReferrerPolicy:          "foo",
							IsDevelopment:           true,
						},
						Errors: map[string]*types.ErrorPage{
							"bar": {
								Status: []string{
									"500",
									"600",
								},
								Backend: "backend-foobar",
								Query:   "bar_query",
							},
							"foo": {
								Status: []string{
									"404",
								},
								Backend: "backend-foobar",
								Query:   "foo_query",
							},
						},
						RateLimit: &types.RateLimit{
							RateSet: map[string]*types.Rate{
								"bar": {
									Period:  flaeg.Duration(3 * time.Second),
									Average: 6,
									Burst:   9,
								},
								"foo": {
									Period:  flaeg.Duration(6 * time.Second),
									Average: 12,
									Burst:   18,
								},
							},
							ExtractorFunc: "client.ip",
						},
						Redirect: &types.Redirect{
							EntryPoint:  "https",
							Regex:       "",
							Replacement: "",
							Permanent:   true,
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

			p := &Provider{ExposedByDefault: true}

			instances := fakeLoadTraefikLabels(test.instances)

			got, err := p.buildConfiguration(instances)
			assert.Equal(t, test.err, err) // , err.Error()
			assert.Equal(t, test.expected, got, test.desc)
		})
	}
}

func TestFilterInstance(t *testing.T) {
	testCases := []struct {
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
			desc: "Instance with nil private ip and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilPrivateIP := simpleEcsInstance(map[string]*string{})
				nilPrivateIP.machine.PrivateIpAddress = nil
				return nilPrivateIP
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with nil machine and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilMachine := simpleEcsInstance(map[string]*string{})
				nilMachine.machine = nil
				return nilMachine
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with nil machine state and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilMachineState := simpleEcsInstance(map[string]*string{})
				nilMachineState.machine.State = nil
				return nilMachineState
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with nil machine state name and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilMachineStateName := simpleEcsInstance(map[string]*string{})
				nilMachineStateName.machine.State.Name = nil
				return nilMachineStateName
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with invalid machine state and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				invalidMachineState := simpleEcsInstance(map[string]*string{})
				invalidMachineState.machine.State.Name = aws.String(ec2.InstanceStateNameStopped)
				return invalidMachineState
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with no port mappings should be filtered",
			instanceInfo:     simpleEcsInstanceNoNetwork(map[string]*string{}),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with no port mapping and with label should not be filtered",
			instanceInfo: simpleEcsInstanceNoNetwork(map[string]*string{
				label.TraefikPort: aws.String("80"),
			}),
			exposedByDefault: true,
			expected:         true,
		},
	}

	for _, test := range testCases {
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
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

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
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getHost(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPort(t *testing.T) {
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getPort(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncStringValue(t *testing.T) {
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

func TestGetFuncSliceString(t *testing.T) {
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

	instance := ecsInstance{
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

	if containerDef != nil {
		instance.TraefikLabels = aws.StringValueMap(containerDef.DockerLabels)
	}

	return instance
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

func fakeLoadTraefikLabels(instances []ecsInstance) []ecsInstance {
	var result []ecsInstance
	for _, instance := range instances {
		instance.TraefikLabels = aws.StringValueMap(instance.containerDefinition.DockerLabels)
		result = append(result, instance)
	}
	return result
}
