package marathon

import (
	"errors"
	"reflect"
	"testing"

	"fmt"

	"github.com/containous/traefik/mocks"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/mock"
)

type fakeClient struct {
	mocks.Marathon
}

func newFakeClient(applicationsError bool, applications *marathon.Applications, tasksError bool, tasks *marathon.Tasks) *fakeClient {
	// create an instance of our test object
	fakeClient := new(fakeClient)
	if applicationsError {
		fakeClient.On("Applications", mock.Anything).Return(nil, errors.New("error"))
	} else {
		fakeClient.On("Applications", mock.Anything).Return(applications, nil)
	}
	if !applicationsError {
		if tasksError {
			fakeClient.On("AllTasks", mock.Anything).Return(nil, errors.New("error"))
		} else {
			fakeClient.On("AllTasks", mock.Anything).Return(tasks, nil)
		}
	}
	return fakeClient
}

func TestMarathonLoadConfig(t *testing.T) {
	cases := []struct {
		applicationsError bool
		applications      *marathon.Applications
		tasksError        bool
		tasks             *marathon.Tasks
		expectedNil       bool
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			applications:      &marathon.Applications{},
			tasks:             &marathon.Tasks{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			applicationsError: true,
			applications:      &marathon.Applications{},
			tasks:             &marathon.Tasks{},
			expectedNil:       true,
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			applications:      &marathon.Applications{},
			tasksError:        true,
			tasks:             &marathon.Tasks{},
			expectedNil:       true,
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "/test",
						Ports:  []int{80},
						Labels: &map[string]string{},
					},
				},
			},
			tasks: &marathon.Tasks{
				Tasks: []marathon.Task{
					{
						ID:    "test",
						AppID: "/test",
						Host:  "localhost",
						Ports: []int{80},
						IPAddresses: []*marathon.IPAddress{
							{
								IPAddress: "127.0.0.1",
								Protocol:  "tcp",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`frontend-test`: {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						`route-host-test`: {
							Rule: "Host:test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "/testLoadBalancerAndCircuitBreaker.dot",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.backend.loadbalancer.method":       "drr",
							"traefik.backend.circuitbreaker.expression": "NetworkErrorRatio() > 0.5",
						},
					},
				},
			},
			tasks: &marathon.Tasks{
				Tasks: []marathon.Task{
					{
						ID:    "testLoadBalancerAndCircuitBreaker.dot",
						AppID: "/testLoadBalancerAndCircuitBreaker.dot",
						Host:  "localhost",
						Ports: []int{80},
						IPAddresses: []*marathon.IPAddress{
							{
								IPAddress: "127.0.0.1",
								Protocol:  "tcp",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`frontend-testLoadBalancerAndCircuitBreaker.dot`: {
					Backend:        "backend-testLoadBalancerAndCircuitBreaker.dot",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						`route-host-testLoadBalancerAndCircuitBreaker.dot`: {
							Rule: "Host:testLoadBalancerAndCircuitBreaker.dot.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-testLoadBalancerAndCircuitBreaker.dot": {
					Servers: map[string]types.Server{
						"server-testLoadBalancerAndCircuitBreaker-dot": {
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "/testMaxConn",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.backend.maxconn.amount":        "1000",
							"traefik.backend.maxconn.extractorfunc": "client.ip",
						},
					},
				},
			},
			tasks: &marathon.Tasks{
				Tasks: []marathon.Task{
					{
						ID:    "testMaxConn",
						AppID: "/testMaxConn",
						Host:  "localhost",
						Ports: []int{80},
						IPAddresses: []*marathon.IPAddress{
							{
								IPAddress: "127.0.0.1",
								Protocol:  "tcp",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`frontend-testMaxConn`: {
					Backend:        "backend-testMaxConn",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						`route-host-testMaxConn`: {
							Rule: "Host:testMaxConn.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-testMaxConn": {
					Servers: map[string]types.Server{
						"server-testMaxConn": {
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "/testMaxConnOnlySpecifyAmount",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.backend.maxconn.amount": "1000",
						},
					},
				},
			},
			tasks: &marathon.Tasks{
				Tasks: []marathon.Task{
					{
						ID:    "testMaxConnOnlySpecifyAmount",
						AppID: "/testMaxConnOnlySpecifyAmount",
						Host:  "localhost",
						Ports: []int{80},
						IPAddresses: []*marathon.IPAddress{
							{
								IPAddress: "127.0.0.1",
								Protocol:  "tcp",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`frontend-testMaxConnOnlySpecifyAmount`: {
					Backend:        "backend-testMaxConnOnlySpecifyAmount",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						`route-host-testMaxConnOnlySpecifyAmount`: {
							Rule: "Host:testMaxConnOnlySpecifyAmount.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-testMaxConnOnlySpecifyAmount": {
					Servers: map[string]types.Server{
						"server-testMaxConnOnlySpecifyAmount": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
		{
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "/testMaxConnOnlyExtractorFunc",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.backend.maxconn.extractorfunc": "client.ip",
						},
					},
				},
			},
			tasks: &marathon.Tasks{
				Tasks: []marathon.Task{
					{
						ID:    "testMaxConnOnlyExtractorFunc",
						AppID: "/testMaxConnOnlyExtractorFunc",
						Host:  "localhost",
						Ports: []int{80},
						IPAddresses: []*marathon.IPAddress{
							{
								IPAddress: "127.0.0.1",
								Protocol:  "tcp",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`frontend-testMaxConnOnlyExtractorFunc`: {
					Backend:        "backend-testMaxConnOnlyExtractorFunc",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						`route-host-testMaxConnOnlyExtractorFunc`: {
							Rule: "Host:testMaxConnOnlyExtractorFunc.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-testMaxConnOnlyExtractorFunc": {
					Servers: map[string]types.Server{
						"server-testMaxConnOnlyExtractorFunc": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
	}

	for _, c := range cases {
		appID := ""
		if len(c.applications.Apps) > 0 {
			appID = c.applications.Apps[0].ID
		}
		t.Run(fmt.Sprintf("app ID: %s", appID), func(t *testing.T) {
			t.Parallel()
			fakeClient := newFakeClient(c.applicationsError, c.applications, c.tasksError, c.tasks)
			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				marathonClient:   fakeClient,
			}
			actualConfig := provider.loadMarathonConfig()
			fakeClient.AssertExpectations(t)
			if c.expectedNil {
				if actualConfig != nil {
					t.Fatalf("configuration should have been nil, got %v", actualConfig)
				}
			} else {
				// Compare backends
				if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
					t.Errorf("got backend %v, want %v", spew.Sdump(actualConfig.Backends), spew.Sdump(c.expectedBackends))
				}
				if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
					t.Errorf("got frontend %v, want %v", spew.Sdump(actualConfig.Frontends), spew.Sdump(c.expectedFrontends))
				}
			}
		})
	}
}

func TestMarathonTaskFilter(t *testing.T) {
	cases := []struct {
		task             marathon.Task
		applications     *marathon.Applications
		expected         bool
		exposedByDefault bool
	}{
		{
			task:             marathon.Task{},
			applications:     &marathon.Applications{},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "test",
				Ports: []int{80},
			},
			applications:     &marathon.Applications{},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "test",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "foo",
						Labels: &map[string]string{},
					},
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "missing-port",
				Ports: []int{},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "missing-port",
						Labels: &map[string]string{},
					},
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "existing-port",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "existing-port",
						Ports:  []int{80},
						Labels: &map[string]string{},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "disable",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.enable": "false",
						},
					},
				},
			},
			expected:         false,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "specify-both-port-index-and-number",
				Ports: []int{80, 443},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "specify-both-port-index-and-number",
						Ports: []int{80, 443},
						Labels: &map[string]string{
							"traefik.port":      "443",
							"traefik.portIndex": "1",
						},
					},
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "healthcheck-available",
						Ports:  []int{80},
						Labels: &map[string]string{},
						HealthChecks: &[]marathon.HealthCheck{
							*marathon.NewDefaultHealthCheck(),
						},
					},
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "healthcheck-false",
						Ports:  []int{80},
						Labels: &map[string]string{},
						HealthChecks: &[]marathon.HealthCheck{
							*marathon.NewDefaultHealthCheck(),
						},
					},
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "healthcheck-mixed-results",
						Ports:  []int{80},
						Labels: &map[string]string{},
						HealthChecks: &[]marathon.HealthCheck{
							*marathon.NewDefaultHealthCheck(),
						},
					},
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
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "healthcheck-alive",
						Ports:  []int{80},
						Labels: &map[string]string{},
						HealthChecks: &[]marathon.HealthCheck{
							*marathon.NewDefaultHealthCheck(),
						},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "disable-default-expose",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "disable-default-expose",
						Ports:  []int{80},
						Labels: &map[string]string{},
					},
				},
			},
			expected:         false,
			exposedByDefault: false,
		},
		{
			task: marathon.Task{
				AppID: "disable-default-expose-disable-in-label",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "disable-default-expose-disable-in-label",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.enable": "false",
						},
					},
				},
			},
			expected:         false,
			exposedByDefault: false,
		},
		{
			task: marathon.Task{
				AppID: "disable-default-expose-enable-in-label",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "disable-default-expose-enable-in-label",
						Ports: []int{80},
						Labels: &map[string]string{
							"traefik.enable": "true",
						},
					},
				},
			},
			expected:         true,
			exposedByDefault: false,
		},
	}

	provider := &Provider{}
	for _, c := range cases {
		actual := provider.taskFilter(c.task, c.applications, c.exposedByDefault)
		if actual != c.expected {
			t.Fatalf("App %s: expected %v, got %v", c.task.AppID, c.expected, actual)
		}
	}
}

func TestMarathonAppConstraints(t *testing.T) {
	cases := []struct {
		application             marathon.Application
		filteredTasks           []marathon.Task
		expected                bool
		marathonLBCompatibility bool
	}{
		{
			application: marathon.Application{
				ID:     "foo1",
				Labels: &map[string]string{},
			},
			filteredTasks: []marathon.Task{
				{
					AppID: "foo1",
				},
			},
			marathonLBCompatibility: false,
			expected:                false,
		},
		{
			application: marathon.Application{
				ID: "foo2",
				Labels: &map[string]string{
					"traefik.tags": "valid",
				},
			},
			filteredTasks: []marathon.Task{
				{
					AppID: "foo2",
				},
			},
			marathonLBCompatibility: false,
			expected:                true,
		},
		{
			application: marathon.Application{
				ID: "foo3",
				Labels: &map[string]string{
					"HAPROXY_GROUP": "valid",
					"traefik.tags":  "notvalid",
				},
			},
			filteredTasks: []marathon.Task{
				{
					AppID: "foo3",
				},
			},
			marathonLBCompatibility: true,
			expected:                true,
		},
	}

	for _, c := range cases {
		provider := &Provider{
			MarathonLBCompatibility: c.marathonLBCompatibility,
		}
		constraint, _ := types.NewConstraint("tag==valid")
		provider.Constraints = types.Constraints{constraint}
		actual := provider.applicationFilter(c.application, c.filteredTasks)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v: %v", c.expected, actual, c.application)
		}
	}

}
func TestMarathonTaskConstraints(t *testing.T) {
	cases := []struct {
		applications            []marathon.Application
		filteredTask            marathon.Task
		expected                bool
		marathonLBCompatibility bool
	}{
		{
			applications: []marathon.Application{
				{
					ID:     "bar1",
					Labels: &map[string]string{},
				}, {
					ID: "foo1",
					Labels: &map[string]string{
						"traefik.tags": "other",
					},
				},
			},
			filteredTask: marathon.Task{
				AppID: "foo1",
				Ports: []int{80},
			},
			marathonLBCompatibility: false,
			expected:                false,
		},
		{
			applications: []marathon.Application{
				{
					ID: "foo2",
					Labels: &map[string]string{
						"traefik.tags": "valid",
					},
				},
			},
			filteredTask: marathon.Task{
				AppID: "foo2",
				Ports: []int{80},
			},
			marathonLBCompatibility: false,
			expected:                true,
		},
		{
			applications: []marathon.Application{
				{
					ID: "foo3",
					Labels: &map[string]string{
						"HAPROXY_GROUP": "valid",
						"traefik.tags":  "notvalid",
					},
				}, {
					ID: "foo4",
					Labels: &map[string]string{
						"HAPROXY_GROUP": "notvalid",
						"traefik.tags":  "valid",
					},
				},
			},
			filteredTask: marathon.Task{
				AppID: "foo3",
				Ports: []int{80},
			},
			marathonLBCompatibility: true,
			expected:                true,
		},
	}

	for _, c := range cases {
		provider := &Provider{
			MarathonLBCompatibility: c.marathonLBCompatibility,
		}
		constraint, _ := types.NewConstraint("tag==valid")
		provider.Constraints = types.Constraints{constraint}
		apps := new(marathon.Applications)
		apps.Apps = c.applications
		actual := provider.taskFilter(c.filteredTask, apps, true)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v: %v", c.expected, actual, c.filteredTask)
		}
	}
}

func TestMarathonApplicationFilter(t *testing.T) {
	cases := []struct {
		application   marathon.Application
		filteredTasks []marathon.Task
		expected      bool
	}{
		{
			application: marathon.Application{
				Labels: &map[string]string{},
			},
			filteredTasks: []marathon.Task{},
			expected:      false,
		},
		{
			application: marathon.Application{
				ID:     "test",
				Labels: &map[string]string{},
			},
			filteredTasks: []marathon.Task{},
			expected:      false,
		},
		{
			application: marathon.Application{
				ID:     "foo",
				Labels: &map[string]string{},
			},
			filteredTasks: []marathon.Task{
				{
					AppID: "bar",
				},
			},
			expected: false,
		},
		{
			application: marathon.Application{
				ID:     "foo",
				Labels: &map[string]string{},
			},
			filteredTasks: []marathon.Task{
				{
					AppID: "foo",
				},
			},
			expected: true,
		},
	}

	provider := &Provider{}
	for _, c := range cases {
		actual := provider.applicationFilter(c.application, c.filteredTasks)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v", c.expected, actual)
		}
	}
}

func TestMarathonGetPort(t *testing.T) {
	provider := &Provider{}

	cases := []struct {
		desc         string
		applications []marathon.Application
		task         marathon.Task
		expected     string
	}{
		{
			desc:         "no applications",
			applications: []marathon.Application{},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "",
		},
		{
			desc: "application mismatch",
			applications: []marathon.Application{
				{
					ID:     "test1",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "test2",
				Ports: []int{80},
			},
			expected: "",
		},
		{
			desc: "port missing",
			applications: []marathon.Application{
				{
					ID:     "app",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "",
		},
		{
			desc: "explicit port taken",
			applications: []marathon.Application{
				{
					ID: "app",
					Labels: &map[string]string{
						"traefik.port": "80",
					},
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
			applications: []marathon.Application{
				{
					ID: "app",
					Labels: &map[string]string{
						"traefik.port": "foobar",
					},
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
			applications: []marathon.Application{
				{
					ID: "app",
					Labels: &map[string]string{
						"traefik.port": "-1",
					},
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
			applications: []marathon.Application{
				{
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
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80},
			},
			expected: "80",
		},
		{
			desc: "port mapping port available",
			applications: []marathon.Application{
				{
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
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "443",
		},
		{
			desc: "IP-per-task port available",
			applications: []marathon.Application{
				{
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
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{},
			},
			expected: "8000",
		},
		{
			desc: "first port taken from multiple ports",
			applications: []marathon.Application{
				{
					ID:     "app",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "app",
				Ports: []int{80, 443},
			},
			expected: "80",
		},
		{
			desc: "indexed port taken",
			applications: []marathon.Application{
				{
					ID: "app",
					Labels: &map[string]string{
						"traefik.portIndex": "1",
					},
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
			applications: []marathon.Application{
				{
					ID: "app",
					Labels: &map[string]string{
						"traefik.portIndex": "foobar",
					},
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
			applications: []marathon.Application{
				{
					ID:     "app",
					Ports:  []int{9999},
					Labels: &map[string]string{},
				},
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
			actual := provider.getPort(c.task, c.applications)
			if actual != c.expected {
				t.Errorf("got %q, want %q", c.expected, actual)
			}
		})
	}
}

func TestMarathonGetWeight(t *testing.T) {
	provider := &Provider{}

	applications := []struct {
		applications []marathon.Application
		task         marathon.Task
		expected     string
	}{
		{
			applications: []marathon.Application{},
			task:         marathon.Task{},
			expected:     "0",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test1",
					Labels: &map[string]string{
						"traefik.weight": "10",
					},
				},
			},
			task: marathon.Task{
				AppID: "test2",
			},
			expected: "0",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test",
					Labels: &map[string]string{
						"traefik.test": "10",
					},
				},
			},
			task: marathon.Task{
				AppID: "test",
			},
			expected: "0",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test",
					Labels: &map[string]string{
						"traefik.weight": "10",
					},
				},
			},
			task: marathon.Task{
				AppID: "test",
			},
			expected: "10",
		},
	}

	for _, a := range applications {
		actual := provider.getWeight(a.task, a.applications)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
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
					"traefik.domain": "foo.bar",
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

	applications := []struct {
		applications []marathon.Application
		task         marathon.Task
		expected     string
	}{
		{
			applications: []marathon.Application{},
			task:         marathon.Task{},
			expected:     "http",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test1",
					Labels: &map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			task: marathon.Task{
				AppID: "test2",
			},
			expected: "http",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test",
					Labels: &map[string]string{
						"traefik.foo": "bar",
					},
				},
			},
			task: marathon.Task{
				AppID: "test",
			},
			expected: "http",
		},
		{
			applications: []marathon.Application{
				{
					ID: "test",
					Labels: &map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			task: marathon.Task{
				AppID: "test",
			},
			expected: "https",
		},
	}

	for _, a := range applications {
		actual := provider.getProtocol(a.task, a.applications)
		if actual != a.expected {
			t.Fatalf("expected %q, got %q", a.expected, actual)
		}
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
					"traefik.frontend.passHostHeader": "false",
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
					"traefik.frontend.entryPoints": "http,https",
				},
			},
			expected: []string{"http", "https"},
		},
	}

	for _, a := range applications {
		actual := provider.getEntryPoints(a.application)

		if !reflect.DeepEqual(actual, a.expected) {
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
					"traefik.frontend.rule": "Host:foo.bar",
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
					"traefik.backend": "bar",
				},
			},
			expected: "bar",
		},
	}

	for _, a := range applications {
		actual := provider.getFrontendBackend(a.application)
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
			desc:        "application missing",
			application: marathon.Application{ID: "other"},
			wantServer:  "",
		},
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
			applications := []marathon.Application{test.application}

			// Populate task.
			test.task.AppID = appID
			test.task.Host = "host"

			gotServer := provider.getBackendServer(test.task, applications)

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
