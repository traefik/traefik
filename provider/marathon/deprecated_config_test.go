package marathon

import (
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigurationAPIErrorsV1(t *testing.T) {
	fakeClient := newFakeClient(true, marathon.Applications{})

	p := &Provider{
		marathonClient: fakeClient,
	}
	p.TemplateVersion = 1

	actualConfig := p.getConfiguration()
	fakeClient.AssertExpectations(t)

	if actualConfig != nil {
		t.Errorf("configuration should have been nil, got %v", actualConfig)
	}
}

func TestBuildConfigurationV1(t *testing.T) {
	testCases := []struct {
		desc              string
		application       marathon.Application
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "simple application",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "filtered task",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80), taskState(taskStateStaging))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {},
			},
		},
		{
			desc: "max connection extractor function label only",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),

				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: label.DefaultWeight,
						},
					},
					MaxConn: nil,
				},
			},
		},
		{
			desc: "multiple ports",
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: label.DefaultWeight,
						},
					},
				},
			},
		},
		{
			desc: "with all labels",
			application: application(
				appPorts(80),
				withTasks(task(host("127.0.0.1"), taskPorts(80))),

				withLabel(label.TraefikPort, "666"),
				withLabel(label.TraefikProtocol, "https"),
				withLabel(label.TraefikWeight, "12"),

				withLabel(label.TraefikBackend, "foobar"),

				withLabel(label.TraefikBackendCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
				withLabel(label.TraefikBackendHealthCheckPath, "/health"),
				withLabel(label.TraefikBackendHealthCheckInterval, "6"),
				withLabel(label.TraefikBackendLoadBalancerMethod, "drr"),
				withLabel(label.TraefikBackendLoadBalancerSticky, "true"),
				withLabel(label.TraefikBackendLoadBalancerStickiness, "true"),
				withLabel(label.TraefikBackendLoadBalancerStickinessCookieName, "chocolate"),
				withLabel(label.TraefikBackendMaxConnAmount, "666"),
				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),

				withLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				withLabel(label.TraefikFrontendEntryPoints, "http,https"),
				withLabel(label.TraefikFrontendPassHostHeader, "true"),
				withLabel(label.TraefikFrontendPriority, "666"),
				withLabel(label.TraefikFrontendRule, "Host:traefik.io"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backendfoobar",
					Routes: map[string]types.Route{
						"route-host-app": {
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
			expectedBackends: map[string]*types.Backend{
				"backendfoobar": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "https://127.0.0.1:666",
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
						Interval: "6",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.application.ID = "/app"

			for _, task := range test.application.Tasks {
				task.ID = "task"
				if task.State == "" {
					task.State = "TASK_RUNNING"
				}
			}

			p := &Provider{
				Domain:           "marathon.localhost",
				ExposedByDefault: true,
			}

			actualConfig := p.buildConfigurationV1(withApplications(test.application))

			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestBuildConfigurationServicesV1(t *testing.T) {
	testCases := []struct {
		desc              string
		application       marathon.Application
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "multiple ports with services",
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),

				withLabel(label.TraefikBackendMaxConnAmount, "1000"),
				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
				withSegmentLabel(label.TraefikPort, "80", "web"),
				withSegmentLabel(label.TraefikPort, "81", "admin"),
				withLabel("traefik..port", "82"), // This should be ignored, as it fails to match the servicesPropertiesRegexp regex.
				withSegmentLabel(label.TraefikFrontendRule, "Host:web.app.marathon.localhost", "web"),
				withSegmentLabel(label.TraefikFrontendRule, "Host:admin.app.marathon.localhost", "admin"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-service-web": {
					Backend: "backend-app-service-web",
					Routes: map[string]types.Route{
						`route-host-app-service-web`: {
							Rule: "Host:web.app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
				"frontend-app-service-admin": {
					Backend: "backend-app-service-admin",
					Routes: map[string]types.Route{
						`route-host-app-service-admin`: {
							Rule: "Host:admin.app.marathon.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-web": {
					Servers: map[string]types.Server{
						"server-task-service-web": {
							URL:    "http://localhost:80",
							Weight: label.DefaultWeight,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
				"backend-app-service-admin": {
					Servers: map[string]types.Server{
						"server-task-service-admin": {
							URL:    "http://localhost:81",
							Weight: label.DefaultWeight,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
		{
			desc: "when all labels are set",
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),

				withLabel(label.TraefikBackendCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
				withLabel(label.TraefikBackendHealthCheckPath, "/health"),
				withLabel(label.TraefikBackendHealthCheckInterval, "6"),
				withLabel(label.TraefikBackendLoadBalancerMethod, "drr"),
				withLabel(label.TraefikBackendLoadBalancerSticky, "true"),
				withLabel(label.TraefikBackendLoadBalancerStickiness, "true"),
				withLabel(label.TraefikBackendLoadBalancerStickinessCookieName, "chocolate"),
				withLabel(label.TraefikBackendMaxConnAmount, "666"),
				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),

				withSegmentLabel(label.TraefikPort, "80", "containous"),
				withSegmentLabel(label.TraefikProtocol, "https", "containous"),
				withSegmentLabel(label.TraefikWeight, "12", "containous"),

				withSegmentLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "containous"),
				withSegmentLabel(label.TraefikFrontendEntryPoints, "http,https", "containous"),
				withSegmentLabel(label.TraefikFrontendPassHostHeader, "true", "containous"),
				withSegmentLabel(label.TraefikFrontendPriority, "666", "containous"),
				withSegmentLabel(label.TraefikFrontendRule, "Host:traefik.io", "containous"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-service-containous": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-app-service-containous",
					Routes: map[string]types.Route{
						"route-host-app-service-containous": {
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
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-containous": {
					Servers: map[string]types.Server{
						"server-task-service-containous": {
							URL:    "https://localhost:80",
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
						Interval: "6",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.application.ID = "/app"

			for _, task := range test.application.Tasks {
				task.ID = "task"
				if task.State == "" {
					task.State = "TASK_RUNNING"
				}
			}

			p := &Provider{
				Domain:           "marathon.localhost",
				ExposedByDefault: true,
			}

			actualConfig := p.buildConfigurationV1(withApplications(test.application))

			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestGetPortV1(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		task        marathon.Task
		serviceName string
		expected    string
	}{
		{
			desc:        "port missing",
			application: application(),
			task:        task(),
			expected:    "",
		},
		{
			desc:        "numeric port",
			application: application(withLabel(label.TraefikPort, "80")),
			task:        task(),
			expected:    "80",
		},
		{
			desc:        "string port",
			application: application(withLabel(label.TraefikPort, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc:        "negative port",
			application: application(withLabel(label.TraefikPort, "-1")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc:        "task port available",
			application: application(),
			task:        task(taskPorts(80)),
			expected:    "80",
		},
		{
			desc: "port definition available",
			application: application(
				portDefinition(443),
			),
			task:     task(),
			expected: "443",
		},
		{
			desc:        "IP-per-task port available",
			application: application(ipAddrPerTask(8000)),
			task:        task(),
			expected:    "8000",
		},
		{
			desc:        "multiple task ports available",
			application: application(),
			task:        task(taskPorts(80, 443)),
			expected:    "80",
		},
		{
			desc:        "numeric port index specified",
			application: application(withLabel(label.TraefikPortIndex, "1")),
			task:        task(taskPorts(80, 443)),
			expected:    "443",
		},
		{
			desc:        "string port index specified",
			application: application(withLabel(label.TraefikPortIndex, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "80",
		},
		{
			desc: "port and port index specified",
			application: application(
				withLabel(label.TraefikPort, "80"),
				withLabel(label.TraefikPortIndex, "1"),
			),
			task:     task(taskPorts(80, 443)),
			expected: "80",
		},
		{
			desc:        "task and application ports specified",
			application: application(appPorts(9999)),
			task:        task(taskPorts(7777)),
			expected:    "7777",
		},
		{
			desc:        "multiple task ports with service index available",
			application: application(withLabel(label.Prefix+"http.portIndex", "0")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
		{
			desc:        "multiple task ports with service port available",
			application: application(withLabel(label.Prefix+"https.port", "443")),
			task:        task(taskPorts(80, 443)),
			serviceName: "https",
			expected:    "443",
		},
		{
			desc:        "multiple task ports with services but default port available",
			application: application(withLabel(label.Prefix+"http.weight", "100")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getPortV1(test.task, test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFrontendRuleV1(t *testing.T) {
	testCases := []struct {
		desc                    string
		application             marathon.Application
		serviceName             string
		expected                string
		marathonLBCompatibility bool
	}{
		{
			desc:                    "label missing",
			application:             application(appID("test")),
			marathonLBCompatibility: true,
			expected:                "Host:test.marathon.localhost",
		},
		{
			desc: "HAProxy vhost available and LB compat disabled",
			application: application(
				appID("test"),
				withLabel("HAPROXY_0_VHOST", "foo.bar"),
			),
			marathonLBCompatibility: false,
			expected:                "Host:test.marathon.localhost",
		},
		{
			desc:                    "HAProxy vhost available and LB compat enabled",
			application:             application(withLabel("HAPROXY_0_VHOST", "foo.bar")),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc: "frontend rule available",

			application: application(
				withLabel(label.TraefikFrontendRule, "Host:foo.bar"),
				withLabel("HAPROXY_0_VHOST", "unused"),
			),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc:                    "service label existing",
			application:             application(withSegmentLabel(label.TraefikFrontendRule, "Host:foo.bar", "app")),
			serviceName:             "app",
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				Domain:                  "marathon.localhost",
				MarathonLBCompatibility: test.marathonLBCompatibility,
			}

			actual := p.getFrontendRuleV1(test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendV1(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(appID("/group/app")),
			expected:    "backend-group-app",
		},
		{
			desc:        "label existing",
			application: application(withLabel(label.TraefikBackend, "bar")),
			expected:    "backendbar",
		},
		{
			desc:        "service label existing",
			application: application(withSegmentLabel(label.TraefikBackend, "bar", "app")),
			serviceName: "app",
			expected:    "backendbar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{}

			actual := p.getBackendNameV1(test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendServerV1(t *testing.T) {
	host := "host"
	testCases := []struct {
		desc              string
		application       marathon.Application
		task              marathon.Task
		forceTaskHostname bool
		expectedServer    string
	}{
		{
			desc:           "application without IP-per-task",
			application:    application(),
			expectedServer: host,
		},
		{
			desc:              "task hostname override",
			application:       application(ipAddrPerTask(8000)),
			forceTaskHostname: true,
			expectedServer:    host,
		},
		{
			desc:           "task IP address missing",
			application:    application(ipAddrPerTask(8000)),
			task:           task(),
			expectedServer: "",
		},
		{
			desc:           "single task IP address",
			application:    application(ipAddrPerTask(8000)),
			task:           task(ipAddresses("1.1.1.1")),
			expectedServer: "1.1.1.1",
		},
		{
			desc:           "multiple task IP addresses without index label",
			application:    application(ipAddrPerTask(8000)),
			task:           task(ipAddresses("1.1.1.1", "2.2.2.2")),
			expectedServer: "",
		},
		{
			desc: "multiple task IP addresses with invalid index label",
			application: application(
				withLabel("traefik.ipAddressIdx", "invalid"),
				ipAddrPerTask(8000),
			),
			task:           task(ipAddresses("1.1.1.1", "2.2.2.2")),
			expectedServer: "",
		},
		{
			desc: "multiple task IP addresses with valid index label",
			application: application(
				withLabel("traefik.ipAddressIdx", "1"),
				ipAddrPerTask(8000),
			),
			task:           task(ipAddresses("1.1.1.1", "2.2.2.2")),
			expectedServer: "2.2.2.2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{ForceTaskHostname: test.forceTaskHostname}
			test.task.Host = host

			actualServer := p.getBackendServerV1(test.task, test.application)

			assert.Equal(t, test.expectedServer, actualServer)
		})
	}
}

func TestGetStickyV1(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		expected    bool
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    false,
		},
		{
			desc:        "label existing",
			application: application(withLabel(label.TraefikBackendLoadBalancerSticky, "true")),
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getStickyV1(test.application)
			if actual != test.expected {
				t.Errorf("actual %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestGetServersV1(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		segmentName string
		expected    map[string]types.Server
	}{
		{
			desc:        "should return nil when no task",
			application: application(ipAddrPerTask(80)),
			expected:    nil,
		},
		{
			desc: "should return nil when all hosts are empty",
			application: application(
				withTasks(
					task(ipAddresses("1.1.1.1"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), withTaskID("B"), taskPorts(80)),
					task(ipAddresses("1.1.1.3"), withTaskID("C"), taskPorts(80))),
			),
			expected: nil,
		},
		{
			desc: "with 3 tasks and hosts set",
			application: application(
				withTasks(
					task(ipAddresses("1.1.1.1"), host("2.2.2.2"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), host("2.2.2.2"), withTaskID("B"), taskPorts(81)),
					task(ipAddresses("1.1.1.3"), host("2.2.2.2"), withTaskID("C"), taskPorts(82))),
			),
			expected: map[string]types.Server{
				"server-A": {
					URL:    "http://2.2.2.2:80",
					Weight: label.DefaultWeight,
				},
				"server-B": {
					URL:    "http://2.2.2.2:81",
					Weight: label.DefaultWeight,
				},
				"server-C": {
					URL:    "http://2.2.2.2:82",
					Weight: label.DefaultWeight,
				},
			},
		},
		{
			desc: "with 3 tasks and ipAddrPerTask set",
			application: application(
				ipAddrPerTask(80),
				withTasks(
					task(ipAddresses("1.1.1.1"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), withTaskID("B"), taskPorts(80)),
					task(ipAddresses("1.1.1.3"), withTaskID("C"), taskPorts(80))),
			),
			expected: map[string]types.Server{
				"server-A": {
					URL:    "http://1.1.1.1:80",
					Weight: label.DefaultWeight,
				},
				"server-B": {
					URL:    "http://1.1.1.2:80",
					Weight: label.DefaultWeight,
				},
				"server-C": {
					URL:    "http://1.1.1.3:80",
					Weight: label.DefaultWeight,
				},
			},
		},
		{
			desc: "with 3 tasks and bridge network",
			application: application(
				bridgeNetwork(),
				withTasks(
					task(ipAddresses("1.1.1.1"), host("2.2.2.2"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), host("2.2.2.2"), withTaskID("B"), taskPorts(81)),
					task(ipAddresses("1.1.1.3"), host("2.2.2.2"), withTaskID("C"), taskPorts(82))),
			),
			expected: map[string]types.Server{
				"server-A": {
					URL:    "http://2.2.2.2:80",
					Weight: label.DefaultWeight,
				},
				"server-B": {
					URL:    "http://2.2.2.2:81",
					Weight: label.DefaultWeight,
				},
				"server-C": {
					URL:    "http://2.2.2.2:82",
					Weight: label.DefaultWeight,
				},
			},
		},
		{
			desc: "with 3 tasks and cni set",
			application: application(
				containerNetwork(),
				withTasks(
					task(ipAddresses("1.1.1.1"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), withTaskID("B"), taskPorts(80)),
					task(ipAddresses("1.1.1.3"), withTaskID("C"), taskPorts(80))),
			),
			expected: map[string]types.Server{
				"server-A": {
					URL:    "http://1.1.1.1:80",
					Weight: label.DefaultWeight,
				},
				"server-B": {
					URL:    "http://1.1.1.2:80",
					Weight: label.DefaultWeight,
				},
				"server-C": {
					URL:    "http://1.1.1.3:80",
					Weight: label.DefaultWeight,
				},
			},
		},
	}

	p := &Provider{}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := p.getServersV1(test.application, test.segmentName)

			assert.Equal(t, test.expected, actual)
		})
	}
}
