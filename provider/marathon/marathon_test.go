package marathon

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/containous/traefik/provider/marathon/mocks"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeClient struct {
	mocks.Marathon
}

func newFakeClient(applicationsError bool, applications marathon.Applications) *fakeClient {
	// create an instance of our test object
	fakeClient := new(fakeClient)
	if applicationsError {
		fakeClient.On("Applications", mock.Anything).Return(nil, errors.New("fake Marathon server error"))
	} else {
		fakeClient.On("Applications", mock.Anything).Return(&applications, nil)
	}
	return fakeClient
}

func TestMarathonLoadConfigAPIErrors(t *testing.T) {
	fakeClient := newFakeClient(true, marathon.Applications{})
	provider := &Provider{
		marathonClient: fakeClient,
	}
	actualConfig := provider.loadMarathonConfig()
	fakeClient.AssertExpectations(t)
	if actualConfig != nil {
		t.Errorf("configuration should have been nil, got %v", actualConfig)
	}
}

func TestMarathonLoadConfigNonAPIErrors(t *testing.T) {
	cases := []struct {
		desc              string
		application       marathon.Application
		task              marathon.Task
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:        "simple application",
			application: application(appPorts(80)),
			task:        localhostTask(taskPorts(80)),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc:        "filtered task",
			application: application(appPorts(80)),
			task: localhostTask(
				taskPorts(80),
				state(taskStateStaging),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: nil,
		},
		{
			desc: "load balancer / circuit breaker labels",
			application: application(
				appPorts(80),
				label(types.LabelBackendLoadbalancerMethod, "drr"),
				label(types.LabelBackendCircuitbreakerExpression, "NetworkErrorRatio() > 0.5"),
			),
			task: localhostTask(taskPorts(80)),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
				},
			},
		},
		{
			desc: "general max connection labels",
			application: application(
				appPorts(80),
				label(types.LabelBackendMaxconnAmount, "1000"),
				label(types.LabelBackendMaxconnExtractorfunc, "client.ip"),
			),
			task: localhostTask(taskPorts(80)),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
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
			desc: "max connection amount label only",
			application: application(
				appPorts(80),
				label(types.LabelBackendMaxconnAmount, "1000"),
			),
			task: localhostTask(taskPorts(80)),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
		{
			desc: "max connection extractor function label only",
			application: application(
				appPorts(80),
				label(types.LabelBackendMaxconnExtractorfunc, "client.ip"),
			),
			task: localhostTask(taskPorts(80)),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
		{
			desc: "health check labels",
			application: application(
				appPorts(80),
				label(types.LabelBackendHealthcheckPath, "/path"),
				label(types.LabelBackendHealthcheckInterval, "5m"),
			),
			task: task(
				host("127.0.0.1"),
				taskPorts(80),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					HealthCheck: &types.HealthCheck{
						Path:     "/path",
						Interval: "5m",
					},
				},
			},
		},
		{
			desc: "multiple ports",
			application: application(
				appPorts(80, 81),
			),
			task: localhostTask(
				taskPorts(80, 81),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
				},
			},
		},
		{
			desc: "multiple ports with services",
			application: application(
				appPorts(80, 81),
				label(types.LabelBackendMaxconnAmount, "1000"),
				label(types.LabelBackendMaxconnExtractorfunc, "client.ip"),
				label("traefik.web.port", "80"),
				label("traefik.admin.port", "81"),
				label("traefik..port", "82"), // This should be ignored, as it fails to match the servicesPropertiesRegexp regex.
				label("traefik.web.frontend.rule", "Host:web.app.docker.localhost"),
				label("traefik.admin.frontend.rule", "Host:admin.app.docker.localhost"),
			),
			task: localhostTask(
				taskPorts(80, 81),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-service-web": {
					Backend: "backend-app-service-web",
					Routes: map[string]types.Route{
						`route-host-app-service-web`: {
							Rule: "Host:web.app.docker.localhost",
						},
					},
				},
				"frontend-app-service-admin": {
					Backend: "backend-app-service-admin",
					Routes: map[string]types.Route{
						`route-host-app-service-admin`: {
							Rule: "Host:admin.app.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-web": {
					Servers: map[string]types.Server{
						"server-task-service-web": {
							URL:    "http://localhost:80",
							Weight: 0,
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
							Weight: 0,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()

			c.application.ID = "/app"
			c.task.ID = "task"
			if c.task.State == "" {
				c.task.State = "TASK_RUNNING"
			}
			c.application.Tasks = []*marathon.Task{&c.task}

			for _, frontend := range c.expectedFrontends {
				frontend.PassHostHeader = true
				frontend.BasicAuth = []string{}
				frontend.EntryPoints = []string{}
			}

			fakeClient := newFakeClient(false,
				marathon.Applications{Apps: []marathon.Application{c.application}})
			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				marathonClient:   fakeClient,
			}
			actualConfig := provider.loadMarathonConfig()
			fakeClient.AssertExpectations(t)

			expectedConfig := &types.Configuration{
				Backends:  c.expectedBackends,
				Frontends: c.expectedFrontends,
			}
			assert.Equal(t, expectedConfig, actualConfig)
		})
	}
}

func TestMarathonTaskFilter(t *testing.T) {
	cases := []struct {
		desc         string
		task         marathon.Task
		application  marathon.Application
		readyChecker *readinessChecker
		expected     bool
	}{
		{
			desc:        "missing port",
			task:        task(),
			application: application(),
			expected:    true,
		},
		{
			desc: "task not running",
			task: task(
				taskPorts(80),
				state(taskStateStaging),
			),
			application: application(appPorts(80)),
			expected:    false,
		},
		{
			desc:        "existing port",
			task:        task(taskPorts(80)),
			application: application(appPorts(80)),
			expected:    true,
		},
		{
			desc: "ambiguous port specification",
			task: task(taskPorts(80, 443)),
			application: application(
				appPorts(80, 443),
				label(types.LabelPort, "443"),
				label(types.LabelPortIndex, "1"),
			),
			expected: true,
		},
		{
			desc: "single service without port",
			task: task(taskPorts(80, 81)),
			application: application(
				appPorts(80, 81),
				labelWithService(types.LabelPort, "80", "web"),
				labelWithService(types.LabelPort, "illegal", "admin"),
			),
			expected: true,
		},
		{
			desc: "single service missing port",
			task: task(taskPorts(80, 81)),
			application: application(
				appPorts(80, 81),
				labelWithService(types.LabelPort, "81", "admin"),
			),
			expected: true,
		},
		{
			desc: "healthcheck available",
			task: task(taskPorts(80)),
			application: application(
				appPorts(80),
				healthChecks(marathon.NewDefaultHealthCheck()),
			),
			expected: true,
		},
		{
			desc: "healthcheck result false",
			task: task(
				taskPorts(80),
				healthCheckResultLiveness(false),
			),
			application: application(
				appPorts(80),
				healthChecks(marathon.NewDefaultHealthCheck()),
			),
			expected: false,
		},
		{
			desc: "healthcheck results mixed",
			task: task(
				taskPorts(80),
				healthCheckResultLiveness(true, false),
			),
			application: application(
				appPorts(80),
				healthChecks(marathon.NewDefaultHealthCheck()),
			),
			expected: false,
		},
		{
			desc: "healthcheck result true",
			task: task(
				taskPorts(80),
				healthCheckResultLiveness(true),
			),
			application: application(
				appPorts(80),
				healthChecks(marathon.NewDefaultHealthCheck()),
			),
			expected: true,
		},
		{
			desc: "readiness check false",
			task: task(taskPorts(80)),
			application: application(
				appPorts(80),
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult(testTaskName, false),
			),
			readyChecker: testReadinessChecker(),
			expected:     false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{readyChecker: c.readyChecker}
			actual := provider.taskFilter(c.task, c.application)
			if actual != c.expected {
				t.Errorf("actual %v, expected %v", actual, c.expected)
			}
		})
	}
}

func TestMarathonApplicationFilterConstraints(t *testing.T) {
	cases := []struct {
		desc                    string
		application             marathon.Application
		marathonLBCompatibility bool
		expected                bool
	}{
		{
			desc:                    "tags missing",
			application:             application(),
			marathonLBCompatibility: false,
			expected:                false,
		},
		{
			desc:                    "tag matching",
			application:             application(label(types.LabelTags, "valid")),
			marathonLBCompatibility: false,
			expected:                true,
		},
		{
			desc: "LB compatibility tag matching",
			application: application(
				label("HAPROXY_GROUP", "valid"),
				label(types.LabelTags, "notvalid"),
			),
			marathonLBCompatibility: true,
			expected:                true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{
				ExposedByDefault:        true,
				MarathonLBCompatibility: c.marathonLBCompatibility,
			}
			constraint, err := types.NewConstraint("tag==valid")
			if err != nil {
				panic(fmt.Sprintf("failed to create constraint 'tag==valid': %s", err))
			}
			provider.Constraints = types.Constraints{constraint}
			actual := provider.applicationFilter(c.application)
			if actual != c.expected {
				t.Errorf("got %v, expected %v", actual, c.expected)
			}
		})
	}
}

func TestMarathonApplicationFilterEnabled(t *testing.T) {
	cases := []struct {
		desc             string
		exposedByDefault bool
		enabledLabel     string
		expected         bool
	}{
		{
			desc:             "exposed",
			exposedByDefault: true,
			enabledLabel:     "",
			expected:         true,
		},
		{
			desc:             "exposed and tolerated by valid label value",
			exposedByDefault: true,
			enabledLabel:     "true",
			expected:         true,
		},
		{
			desc:             "exposed and tolerated by invalid label value",
			exposedByDefault: true,
			enabledLabel:     "invalid",
			expected:         true,
		},
		{
			desc:             "exposed but overridden by label",
			exposedByDefault: true,
			enabledLabel:     "false",
			expected:         false,
		},
		{
			desc:             "non-exposed",
			exposedByDefault: false,
			enabledLabel:     "",
			expected:         false,
		},
		{
			desc:             "non-exposed but overridden by label",
			exposedByDefault: false,
			enabledLabel:     "true",
			expected:         true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{ExposedByDefault: c.exposedByDefault}
			app := application(label(types.LabelEnable, c.enabledLabel))
			if provider.applicationFilter(app) != c.expected {
				t.Errorf("got unexpected filtering = %t", !c.expected)
			}
		})
	}
}

func TestMarathonGetPort(t *testing.T) {
	provider := &Provider{}

	cases := []struct {
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
			application: application(label(types.LabelPort, "80")),
			task:        task(),
			expected:    "80",
		},
		{
			desc:        "string port",
			application: application(label(types.LabelPort, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc:        "negative port",
			application: application(label(types.LabelPort, "-1")),
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
			application: application(label(types.LabelPortIndex, "1")),
			task:        task(taskPorts(80, 443)),
			expected:    "443",
		},
		{
			desc:        "string port index specified",
			application: application(label(types.LabelPortIndex, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc: "port and port index specified",
			application: application(
				label(types.LabelPort, "80"),
				label(types.LabelPortIndex, "1"),
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
			application: application(label(types.LabelPrefix+"http.portIndex", "0")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
		{
			desc:        "multiple task ports with service port available",
			application: application(label(types.LabelPrefix+"https.port", "443")),
			task:        task(taskPorts(80, 443)),
			serviceName: "https",
			expected:    "443",
		},
		{
			desc:        "multiple task ports with services but default port available",
			application: application(label(types.LabelPrefix+"http.weight", "100")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			actual := provider.getPort(c.task, c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", c.expected, actual)
			}
		})
	}
}

func TestMarathonGetWeight(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "0",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelWeight, "10")),
			expected:    "10",
		},
		{
			desc:        "service label existing",
			application: application(labelWithService(types.LabelWeight, "10", "app")),
			serviceName: "app",
			expected:    "10",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getWeight(c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetDomain(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "docker.localhost",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelDomain, "foo.bar")),
			expected:    "foo.bar",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getDomain(c.application)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetProtocol(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "http",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelProtocol, "https")),
			expected:    "https",
		},
		{
			desc:        "service label existing",
			application: application(labelWithService(types.LabelProtocol, "https", "app")),
			serviceName: "app",
			expected:    "https",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getProtocol(c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetSticky(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "false",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackendLoadbalancerSticky, "true")),
			expected:    "true",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getSticky(c.application)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetPassHostHeader(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "true",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelFrontendPassHostHeader, "false")),
			expected:    "false",
		},
		{
			desc:        "label existing",
			application: application(labelWithService(types.LabelFrontendPassHostHeader, "false", "app")),
			serviceName: "app",
			expected:    "false",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getPassHostHeader(c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonMaxConnAmount(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    int64
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    math.MaxInt64,
		},
		{
			desc:        "non-integer value",
			application: application(label(types.LabelBackendMaxconnAmount, "foobar")),
			expected:    math.MaxInt64,
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackendMaxconnAmount, "32")),
			expected:    32,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getMaxConnAmount(c.application)
			if actual != c.expected {
				t.Errorf("actual %d, expected %d", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetMaxConnExtractorFunc(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "request.host",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackendMaxconnExtractorfunc, "client.ip")),
			expected:    "client.ip",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getMaxConnExtractorFunc(c.application)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetLoadBalancerMethod(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "wrr",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackendLoadbalancerMethod, "drr")),
			expected:    "drr",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getLoadBalancerMethod(c.application)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetCircuitBreakerExpression(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    "NetworkErrorRatio() > 1",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackendCircuitbreakerExpression, "NetworkErrorRatio() > 0.5")),
			expected:    "NetworkErrorRatio() > 0.5",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getCircuitBreakerExpression(c.application)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetEntryPoints(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    []string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    []string{},
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelFrontendEntryPoints, "http,https")),
			expected:    []string{"http", "https"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getEntryPoints(c.application, "")
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("actual %#v, expected %#v", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetFrontendRule(t *testing.T) {
	cases := []struct {
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
			expected:                "Host:test.docker.localhost",
		},
		{
			desc: "HAProxy vhost available and LB compat disabled",
			application: application(
				appID("test"),
				label("HAPROXY_0_VHOST", "foo.bar"),
			),
			marathonLBCompatibility: false,
			expected:                "Host:test.docker.localhost",
		},
		{
			desc:                    "HAProxy vhost available and LB compat enabled",
			application:             application(label("HAPROXY_0_VHOST", "foo.bar")),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc: "frontend rule available",

			application: application(
				label(types.LabelFrontendRule, "Host:foo.bar"),
				label("HAPROXY_0_VHOST", "unused"),
			),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc:                    "service label existing",
			application:             application(labelWithService(types.LabelFrontendRule, "Host:foo.bar", "app")),
			serviceName:             "app",
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{
				Domain:                  "docker.localhost",
				MarathonLBCompatibility: c.marathonLBCompatibility,
			}
			actual := provider.getFrontendRule(c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetBackend(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(appID("/group/app")),
			expected:    "-group-app",
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelBackend, "bar")),
			expected:    "bar",
		},
		{
			desc:        "service label existing",
			application: application(labelWithService(types.LabelBackend, "bar", "app")),
			serviceName: "app",
			expected:    "bar",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getBackend(c.application, c.serviceName)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetSubDomain(t *testing.T) {
	cases := []struct {
		path             string
		expected         string
		groupAsSubDomain bool
	}{
		{"/test", "test", false},
		{"/test", "test", true},
		{"/a/b/c/d", "d.c.b.a", true},
		{"/b/a/d/c", "c.d.a.b", true},
		{"/d/c/b/a", "a.b.c.d", true},
		{"/c/d/a/b", "b.a.d.c", true},
		{"/a/b/c/d", "a-b-c-d", false},
		{"/b/a/d/c", "b-a-d-c", false},
		{"/d/c/b/a", "d-c-b-a", false},
		{"/c/d/a/b", "c-d-a-b", false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("path=%s,group=%t", c.path, c.groupAsSubDomain), func(t *testing.T) {
			t.Parallel()
			provider := &Provider{GroupsAsSubDomains: c.groupAsSubDomain}
			actual := provider.getSubDomain(c.path)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonHasHealthCheckLabels(t *testing.T) {
	cases := []struct {
		desc     string
		value    *string
		expected bool
	}{
		{
			desc:     "label missing",
			value:    nil,
			expected: false,
		},
		{
			desc:     "empty path",
			value:    testhelpers.Stringp(""),
			expected: false,
		},
		{
			desc:     "non-empty path",
			value:    testhelpers.Stringp("/path"),
			expected: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			app := application()
			if c.value != nil {
				app.AddLabel(types.LabelBackendHealthcheckPath, *c.value)
			}
			provider := &Provider{}
			actual := provider.hasHealthCheckLabels(app)
			if actual != c.expected {
				t.Errorf("actual %t, expected %t", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetHealthCheckPath(t *testing.T) {
	cases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "label missing",
			expected: "",
		},
		{
			desc:     "path existing",
			value:    "/path",
			expected: "/path",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			app := application()
			if c.value != "" {
				app.AddLabel(types.LabelBackendHealthcheckPath, c.value)
			}
			provider := &Provider{}
			actual := provider.getHealthCheckPath(app)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetHealthCheckInterval(t *testing.T) {
	cases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "label missing",
			expected: "",
		},
		{
			desc:     "interval existing",
			value:    "5m",
			expected: "5m",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			app := application()
			if c.value != "" {
				app.AddLabel(types.LabelBackendHealthcheckInterval, c.value)
			}
			provider := &Provider{}
			actual := provider.getHealthCheckInterval(app)
			if actual != c.expected {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}

func TestGetBackendServer(t *testing.T) {
	host := "host"
	cases := []struct {
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
				label("traefik.ipAddressIdx", "invalid"),
				ipAddrPerTask(8000),
			),
			task:           task(ipAddresses("1.1.1.1", "2.2.2.2")),
			expectedServer: "",
		},
		{
			desc: "multiple task IP addresses with valid index label",
			application: application(
				label("traefik.ipAddressIdx", "1"),
				ipAddrPerTask(8000),
			),
			task:           task(ipAddresses("1.1.1.1", "2.2.2.2")),
			expectedServer: "2.2.2.2",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{ForceTaskHostname: c.forceTaskHostname}
			c.task.Host = host
			actualServer := provider.getBackendServer(c.task, c.application)
			if actualServer != c.expectedServer {
				t.Errorf("actual %q, expected %q", actualServer, c.expectedServer)
			}
		})
	}
}

func TestParseIndex(t *testing.T) {
	cases := []struct {
		idxStr        string
		length        int
		shouldSucceed bool
		parsed        int
	}{
		{
			idxStr:        "illegal",
			length:        42,
			shouldSucceed: false,
		},
		{
			idxStr:        "-1",
			length:        42,
			shouldSucceed: false,
		},
		{
			idxStr:        "10",
			length:        1,
			shouldSucceed: false,
		},
		{
			idxStr:        "10",
			length:        10,
			shouldSucceed: false,
		},
		{
			idxStr:        "0",
			length:        1,
			shouldSucceed: true,
			parsed:        0,
		},
		{
			idxStr:        "10",
			length:        11,
			shouldSucceed: true,
			parsed:        10,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("parseIndex(%s, %d)", c.idxStr, c.length), func(t *testing.T) {
			t.Parallel()
			parsed, err := parseIndex(c.idxStr, c.length)

			if c.shouldSucceed != (err == nil) {
				t.Fatalf("actual error %q, expected error: %t", err, !c.shouldSucceed)
			}

			if c.shouldSucceed && parsed != c.parsed {
				t.Errorf("actual parsed index %d, expected %d", parsed, c.parsed)
			}
		})
	}
}

func TestMarathonGetBasicAuth(t *testing.T) {
	cases := []struct {
		desc        string
		application marathon.Application
		expected    []string
	}{
		{
			desc:        "label missing",
			application: application(),
			expected:    []string{},
		},
		{
			desc:        "label existing",
			application: application(label(types.LabelFrontendAuthBasic, "user:password")),
			expected:    []string{"user:password"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{}
			actual := provider.getBasicAuth(c.application, "")
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("actual %q, expected %q", actual, c.expected)
			}
		})
	}
}
