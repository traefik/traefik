package marathon

import (
	"errors"
	"fmt"
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
			desc: "simple application",
			application: marathon.Application{
				Ports:  []int{80},
				Labels: &map[string]string{},
			},
			task: marathon.Task{
				Host:  "localhost",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			desc: "filtered task",
			application: marathon.Application{
				Ports:  []int{80},
				Labels: &map[string]string{},
			},
			task: marathon.Task{
				Ports: []int{80},
				State: "TASK_STAGING",
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			application: marathon.Application{

				Ports: []int{80},
				Labels: &map[string]string{
					types.LabelBackendLoadbalancerMethod:       "drr",
					types.LabelBackendCircuitbreakerExpression: "NetworkErrorRatio() > 0.5",
				},
			},
			task: marathon.Task{
				Host:  "localhost",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			application: marathon.Application{
				Ports: []int{80},
				Labels: &map[string]string{
					types.LabelBackendMaxconnAmount:        "1000",
					types.LabelBackendMaxconnExtractorfunc: "client.ip",
				},
			},
			task: marathon.Task{
				Host:  "localhost",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			desc: "max connection amount label",
			application: marathon.Application{
				Ports: []int{80},
				Labels: &map[string]string{
					types.LabelBackendMaxconnAmount: "1000",
				},
			},
			task: marathon.Task{
				Host:  "localhost",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			desc: "max connection extractor function label",
			application: marathon.Application{
				Ports: []int{80},
				Labels: &map[string]string{
					types.LabelBackendMaxconnExtractorfunc: "client.ip",
				},
			},
			task: marathon.Task{
				Host:  "localhost",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
			application: marathon.Application{
				Ports: []int{80},
				Labels: &map[string]string{
					types.LabelBackendHealthcheckPath:     "/path",
					types.LabelBackendHealthcheckInterval: "5m",
				},
			},
			task: marathon.Task{
				Host:  "127.0.0.1",
				Ports: []int{80},
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "127.0.0.1",
						Protocol:  "tcp",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend:        "backend-app",
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
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
		task             marathon.Task
		application      marathon.Application
		expected         bool
		exposedByDefault bool
	}{
		{
			task: marathon.Task{
				AppID: "missing-port",
				Ports: []int{},
			},
			application: marathon.Application{
				ID:     "missing-port",
				Labels: &map[string]string{},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "task-not-running",
				Ports: []int{80},
				State: "TASK_STAGING",
			},
			application: marathon.Application{
				ID:     "task-not-running",
				Ports:  []int{80},
				Labels: &map[string]string{},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "existing-port",
				Ports: []int{80},
			},
			application: marathon.Application{
				ID:     "existing-port",
				Ports:  []int{80},
				Labels: &map[string]string{},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "specify-both-port-index-and-number",
				Ports: []int{80, 443},
			},
			application: marathon.Application{
				ID:    "specify-both-port-index-and-number",
				Ports: []int{80, 443},
				Labels: &map[string]string{
					types.LabelPort:      "443",
					types.LabelPortIndex: "1",
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "healthcheck-available",
				Ports: []int{80},
			},
			application: marathon.Application{
				ID:     "healthcheck-available",
				Ports:  []int{80},
				Labels: &map[string]string{},
				HealthChecks: &[]marathon.HealthCheck{
					*marathon.NewDefaultHealthCheck(),
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "healthcheck-false",
				Ports: []int{80},
				HealthCheckResults: []*marathon.HealthCheckResult{
					{
						Alive: false,
					},
				},
			},
			application: marathon.Application{
				ID:     "healthcheck-false",
				Ports:  []int{80},
				Labels: &map[string]string{},
				HealthChecks: &[]marathon.HealthCheck{
					*marathon.NewDefaultHealthCheck(),
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "healthcheck-mixed-results",
				Ports: []int{80},
				HealthCheckResults: []*marathon.HealthCheckResult{
					{
						Alive: true,
					},
					{
						Alive: false,
					},
				},
			},
			application: marathon.Application{
				ID:     "healthcheck-mixed-results",
				Ports:  []int{80},
				Labels: &map[string]string{},
				HealthChecks: &[]marathon.HealthCheck{
					*marathon.NewDefaultHealthCheck(),
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "healthcheck-alive",
				Ports: []int{80},
				HealthCheckResults: []*marathon.HealthCheckResult{
					{
						Alive: true,
					},
				},
			},
			application: marathon.Application{
				ID:     "healthcheck-alive",
				Ports:  []int{80},
				Labels: &map[string]string{},
				HealthChecks: &[]marathon.HealthCheck{
					*marathon.NewDefaultHealthCheck(),
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
	}

	provider := &Provider{}
	for i, c := range cases {
		if c.task.State == "" {
			c.task.State = taskStateRunning
		}
		actual := provider.taskFilter(c.task, c.application)
		if actual != c.expected {
			t.Fatalf("App %s (#%d): got %v, expected %v", c.task.AppID, i, actual, c.expected)
		}
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
			desc: "tags missing",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
			},
			marathonLBCompatibility: false,
			expected:                false,
		},
		{
			desc: "tag matching",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelTags: "valid",
				},
			},
			marathonLBCompatibility: false,
			expected:                true,
		},
		{
			desc: "LB compatibility tag matching",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					"HAPROXY_GROUP": "valid",
					types.LabelTags: "notvalid",
				},
			},
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
				t.Fatalf("got %v, expected %v", actual, c.expected)
			}
		})
	}
}

func TestMarathonApplicationFilterEnabled(t *testing.T) {
	cases := []struct {
		desc         string
		exposed      bool
		enabledLabel string
		expected     bool
	}{
		{
			desc:         "exposed",
			exposed:      true,
			enabledLabel: "",
			expected:     true,
		},
		{
			desc:         "exposed and tolerated by valid label value",
			exposed:      true,
			enabledLabel: "true",
			expected:     true,
		},
		{
			desc:         "exposed and tolerated by invalid label value",
			exposed:      true,
			enabledLabel: "invalid",
			expected:     true,
		},
		{
			desc:         "exposed but overridden by label",
			exposed:      true,
			enabledLabel: "false",
			expected:     false,
		},
		{
			desc:         "non-exposed",
			exposed:      false,
			enabledLabel: "",
			expected:     false,
		},
		{
			desc:         "non-exposed but overridden by label",
			exposed:      false,
			enabledLabel: "true",
			expected:     true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{ExposedByDefault: c.exposed}
			app := marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelEnable: c.enabledLabel,
				},
			}
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
		expected    string
	}{
		{
			desc: "port missing",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "",
		},
		{
			desc: "explicit port taken",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelPort: "80",
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "80",
		},
		{
			desc: "illegal explicit port specified",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelPort: "foobar",
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "",
		},
		{
			desc: "illegal explicit port integer specified",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelPort: "-1",
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "",
		},
		{
			desc: "task port available",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
				PortDefinitions: &[]marathon.PortDefinition{
					{
						Port: testhelpers.Intp(443),
					},
				},
				IPAddressPerTask: &marathon.IPAddressPerTask{
					Discovery: &marathon.Discovery{
						Ports: &[]marathon.Port{
							{
								Number: 8000,
							},
						},
					},
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "80",
		},
		{
			desc: "port mapping port available",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
				PortDefinitions: &[]marathon.PortDefinition{
					{
						Port: testhelpers.Intp(443),
					},
				},
				IPAddressPerTask: &marathon.IPAddressPerTask{
					Discovery: &marathon.Discovery{
						Ports: &[]marathon.Port{
							{
								Number: 8000,
							},
						},
					},
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "443",
		},
		{
			desc: "IP-per-task port available",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
				IPAddressPerTask: &marathon.IPAddressPerTask{
					Discovery: &marathon.Discovery{
						Ports: &[]marathon.Port{
							{
								Number: 8000,
							},
						},
					},
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "8000",
		},
		{
			desc: "first port taken from multiple ports",
			application: marathon.Application{
				ID:     "app",
				Labels: &map[string]string{},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80, 443},
			},
			expected: "80",
		},
		{
			desc: "indexed port taken",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelPortIndex: "1",
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80, 443},
			},
			expected: "443",
		},
		{
			desc: "illegal port index specified",
			application: marathon.Application{
				ID: "app",
				Labels: &map[string]string{
					types.LabelPortIndex: "foobar",
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "",
		},
		{
			desc: "task port preferred over application port",
			application: marathon.Application{
				ID:     "app",
				Ports:  []int{9999},
				Labels: &map[string]string{},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{7777},
			},
			expected: "7777",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			actual := provider.getPort(c.task, c.application)
			if actual != c.expected {
				t.Errorf("got %q, want %q", c.expected, actual)
			}
		})
	}
}

func TestMarathonGetWeight(t *testing.T) {
	provider := &Provider{}

	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc: "weight label missing",
			application: marathon.Application{
				Labels: &map[string]string{},
			},
			expected: "0",
		},
		{
			desc: "weight label existing",
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelWeight: "10",
				},
			},
			expected: "10",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			actual := provider.getWeight(c.application)
			if actual != c.expected {
				t.Fatalf("actual %s, expected %s", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetDomain(t *testing.T) {
	provider := &Provider{
		Domain: "docker.localhost",
	}

	applications := []struct {
		application marathon.Application
		expected    string
	}{
		{
			application: marathon.Application{
				Labels: &map[string]string{}},
			expected: "docker.localhost",
		},
		{
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelDomain: "foo.bar",
				},
			},
			expected: "foo.bar",
		},
	}

	for _, a := range applications {
		actual := provider.getDomain(a.application)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
	}
}

func TestMarathonGetProtocol(t *testing.T) {
	provider := &Provider{}

	cases := []struct {
		desc        string
		application marathon.Application
		expected    string
	}{
		{
			desc: "protocol label missing",
			application: marathon.Application{
				Labels: &map[string]string{},
			},
			expected: "http",
		},
		{
			desc: "protocol label existing",
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelProtocol: "https",
				},
			},
			expected: "https",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			actual := provider.getProtocol(c.application)
			if actual != c.expected {
				t.Errorf("got protocol '%s', want '%s'", actual, c.expected)
			}
		})
	}
}

func TestMarathonGetPassHostHeader(t *testing.T) {
	provider := &Provider{}

	applications := []struct {
		application marathon.Application
		expected    string
	}{
		{
			application: marathon.Application{
				Labels: &map[string]string{}},
			expected: "true",
		},
		{
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelFrontendPassHostHeader: "false",
				},
			},
			expected: "false",
		},
	}

	for _, a := range applications {
		actual := provider.getPassHostHeader(a.application)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
	}
}

func TestMarathonGetEntryPoints(t *testing.T) {
	provider := &Provider{}

	applications := []struct {
		application marathon.Application
		expected    []string
	}{
		{
			application: marathon.Application{
				Labels: &map[string]string{}},
			expected: []string{},
		},
		{
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelFrontendEntryPoints: "http,https",
				},
			},
			expected: []string{"http", "https"},
		},
	}

	for _, a := range applications {
		actual := provider.getEntryPoints(a.application)

		if !reflect.DeepEqual(a.expected, actual) {
			t.Fatalf("expected %#v, got %#v", a.expected, actual)
		}
	}
}

func TestMarathonGetFrontendRule(t *testing.T) {
	applications := []struct {
		application             marathon.Application
		expected                string
		marathonLBCompatibility bool
	}{
		{
			application: marathon.Application{
				Labels: &map[string]string{}},
			marathonLBCompatibility: true,
			expected:                "Host:.docker.localhost",
		},
		{
			application: marathon.Application{
				ID: "test",
				Labels: &map[string]string{
					"HAPROXY_0_VHOST": "foo.bar",
				},
			},
			marathonLBCompatibility: false,
			expected:                "Host:test.docker.localhost",
		},
		{
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelFrontendRule: "Host:foo.bar",
					"HAPROXY_0_VHOST":       "notvalid",
				},
			},
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			application: marathon.Application{
				Labels: &map[string]string{
					"HAPROXY_0_VHOST": "foo.bar",
				},
			},
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
	}

	for _, a := range applications {
		provider := &Provider{
			Domain:                  "docker.localhost",
			MarathonLBCompatibility: a.marathonLBCompatibility,
		}
		actual := provider.getFrontendRule(a.application)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
	}
}

func TestMarathonGetBackend(t *testing.T) {
	provider := &Provider{}

	applications := []struct {
		application marathon.Application
		expected    string
	}{
		{
			application: marathon.Application{
				ID: "foo",
				Labels: &map[string]string{
					types.LabelBackend: "bar",
				},
			},
			expected: "bar",
		},
	}

	for _, a := range applications {
		actual := provider.getBackend(a.application)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
	}
}

func TestMarathonGetSubDomain(t *testing.T) {
	providerGroups := &Provider{GroupsAsSubDomains: true}
	providerNoGroups := &Provider{GroupsAsSubDomains: false}

	apps := []struct {
		path     string
		expected string
		provider *Provider
	}{
		{"/test", "test", providerNoGroups},
		{"/test", "test", providerGroups},
		{"/a/b/c/d", "d.c.b.a", providerGroups},
		{"/b/a/d/c", "c.d.a.b", providerGroups},
		{"/d/c/b/a", "a.b.c.d", providerGroups},
		{"/c/d/a/b", "b.a.d.c", providerGroups},
		{"/a/b/c/d", "a-b-c-d", providerNoGroups},
		{"/b/a/d/c", "b-a-d-c", providerNoGroups},
		{"/d/c/b/a", "d-c-b-a", providerNoGroups},
		{"/c/d/a/b", "c-d-a-b", providerNoGroups},
	}

	for _, a := range apps {
		actual := a.provider.getSubDomain(a.path)

		if actual != a.expected {
			t.Errorf("expected %q, got %q", a.expected, actual)
		}
	}
}

func TestMarathonHasHealthCheckLabels(t *testing.T) {
	tests := []struct {
		desc  string
		value *string
		want  bool
	}{
		{
			desc:  "label missing",
			value: nil,
			want:  false,
		},
		{
			desc:  "empty path",
			value: stringp(""),
			want:  false,
		},
		{
			desc:  "non-empty path",
			value: stringp("/path"),
			want:  true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			app := marathon.Application{
				Labels: &map[string]string{},
			}
			if test.value != nil {
				app.AddLabel(types.LabelBackendHealthcheckPath, *test.value)
			}
			prov := &Provider{}
			got := prov.hasHealthCheckLabels(app)
			if got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}

func TestMarathonGetHealthCheckPath(t *testing.T) {
	tests := []struct {
		desc  string
		value *string
		want  string
	}{
		{
			desc:  "label missing",
			value: nil,
			want:  "",
		},
		{
			desc:  "path existing",
			value: stringp("/path"),
			want:  "/path",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			app := marathon.Application{}
			app.EmptyLabels()
			if test.value != nil {
				app.AddLabel(types.LabelBackendHealthcheckPath, *test.value)
			}
			prov := &Provider{}
			got := prov.getHealthCheckPath(app)
			if got != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}

func TestMarathonGetHealthCheckInterval(t *testing.T) {
	tests := []struct {
		desc  string
		value *string
		want  string
	}{
		{
			desc:  "label missing",
			value: nil,
			want:  "",
		},
		{
			desc:  "interval existing",
			value: stringp("5m"),
			want:  "5m",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			app := marathon.Application{
				Labels: &map[string]string{},
			}
			if test.value != nil {
				app.AddLabel(types.LabelBackendHealthcheckInterval, *test.value)
			}
			prov := &Provider{}
			got := prov.getHealthCheckInterval(app)
			if got != test.want {
				t.Errorf("got %s, want %s", got, test.want)
			}
		})
	}
}

func stringp(s string) *string {
	return &s
}

func TestGetBackendServer(t *testing.T) {
	appID := "appId"
	host := "host"
	tests := []struct {
		desc              string
		application       marathon.Application
		addIPAddrPerTask  bool
		task              marathon.Task
		forceTaskHostname bool
		wantServer        string
	}{
		{
			desc:       "application without IP-per-task",
			wantServer: host,
		},
		{
			desc:              "task hostname override",
			addIPAddrPerTask:  true,
			forceTaskHostname: true,
			wantServer:        host,
		},
		{
			desc: "task IP address missing",
			task: marathon.Task{
				IPAddresses: []*marathon.IPAddress{},
			},
			addIPAddrPerTask: true,
			wantServer:       "",
		},
		{
			desc: "single task IP address",
			task: marathon.Task{
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "1.1.1.1",
					},
				},
			},
			addIPAddrPerTask: true,
			wantServer:       "1.1.1.1",
		},
		{
			desc: "multiple task IP addresses without index label",
			task: marathon.Task{
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "1.1.1.1",
					},
					{
						IPAddress: "2.2.2.2",
					},
				},
			},
			addIPAddrPerTask: true,
			wantServer:       "",
		},
		{
			desc: "multiple task IP addresses with invalid index label",
			application: marathon.Application{
				Labels: &map[string]string{"traefik.ipAddressIdx": "invalid"},
			},
			task: marathon.Task{
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "1.1.1.1",
					},
					{
						IPAddress: "2.2.2.2",
					},
				},
			},
			addIPAddrPerTask: true,
			wantServer:       "",
		},
		{
			desc: "multiple task IP addresses with valid index label",
			application: marathon.Application{
				Labels: &map[string]string{"traefik.ipAddressIdx": "1"},
			},
			task: marathon.Task{
				IPAddresses: []*marathon.IPAddress{
					{
						IPAddress: "1.1.1.1",
					},
					{
						IPAddress: "2.2.2.2",
					},
				},
			},
			addIPAddrPerTask: true,
			wantServer:       "2.2.2.2",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			provider := &Provider{ForceTaskHostname: test.forceTaskHostname}

			// Populate application.
			if test.application.ID == "" {
				test.application.ID = appID
			}
			if test.application.Labels == nil {
				test.application.Labels = &map[string]string{}
			}
			if test.addIPAddrPerTask {
				test.application.IPAddressPerTask = &marathon.IPAddressPerTask{
					Discovery: &marathon.Discovery{
						Ports: &[]marathon.Port{
							{
								Number: 8000,
								Name:   "port",
							},
						},
					},
				}
			}

			// Populate task.
			test.task.AppID = appID
			test.task.Host = "host"

			gotServer := provider.getBackendServer(test.task, test.application)

			if gotServer != test.wantServer {
				t.Errorf("got server '%s', want '%s'", gotServer, test.wantServer)
			}
		})
	}
}

func TestParseIndex(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("parseIndex(%s, %d)", test.idxStr, test.length), func(t *testing.T) {
			t.Parallel()
			parsed, err := parseIndex(test.idxStr, test.length)

			if test.shouldSucceed != (err == nil) {
				t.Fatalf("got error '%s', want error: %t", err, !test.shouldSucceed)
			}

			if test.shouldSucceed && parsed != test.parsed {
				t.Errorf("got parsed index %d, want %d", parsed, test.parsed)
			}
		})
	}
}

func TestMarathonGetBasicAuth(t *testing.T) {
	provider := &Provider{}

	cases := []struct {
		desc        string
		application marathon.Application
		expected    []string
	}{
		{
			desc: "basic auth label is empty",
			application: marathon.Application{
				Labels: &map[string]string{}},
			expected: []string{},
		},
		{
			desc: "basic auth label is set with user:password",
			application: marathon.Application{
				Labels: &map[string]string{
					types.LabelFrontendAuthBasic: "user:password",
				},
			},
			expected: []string{"user:password"},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getBasicAuth(c.application)
			if !reflect.DeepEqual(c.expected, actual) {
				t.Errorf("expected %q, got %q", c.expected, actual)
			}
		})
	}
}
