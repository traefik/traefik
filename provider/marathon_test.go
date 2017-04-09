package provider

import (
	"errors"
	"reflect"
	"testing"

	"github.com/containous/traefik/mocks"
	"github.com/containous/traefik/types"
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
							URL:    "http://127.0.0.1:80",
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
							URL:    "http://127.0.0.1:80",
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
							URL:    "http://127.0.0.1:80",
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
							URL:    "http://127.0.0.1:80",
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
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
	}

	for _, c := range cases {
		fakeClient := newFakeClient(c.applicationsError, c.applications, c.tasksError, c.tasks)
		provider := &Marathon{
			Domain:           "docker.localhost",
			ExposedByDefault: true,
			marathonClient:   fakeClient,
		}
		actualConfig := provider.loadMarathonConfig()
		fakeClient.AssertExpectations(t)
		if c.expectedNil {
			if actualConfig != nil {
				t.Fatalf("Should have been nil, got %v", actualConfig)
			}
		} else {
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		}
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
				AppID: "multiple-ports",
				Ports: []int{80, 443},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "multiple-ports",
						Ports:  []int{80, 443},
						Labels: &map[string]string{},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "ipAddressOnePort",
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID: "ipAddressOnePort",
						IPAddressPerTask: &marathon.IPAddressPerTask{
							Discovery: &marathon.Discovery{
								Ports: &[]marathon.Port{
									{
										Number: 8880,
										Name:   "p1",
									},
								},
							},
						},
						Labels: &map[string]string{},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "ipAddressTwoPortsUseFirst",
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID: "ipAddressTwoPortsUseFirst",
						IPAddressPerTask: &marathon.IPAddressPerTask{
							Discovery: &marathon.Discovery{
								Ports: &[]marathon.Port{
									{
										Number: 8898,
										Name:   "p1",
									}, {
										Number: 9999,
										Name:   "p1",
									},
								},
							},
						},
						Labels: &map[string]string{},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "ipAddressValidTwoPorts",
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID: "ipAddressValidTwoPorts",
						IPAddressPerTask: &marathon.IPAddressPerTask{
							Discovery: &marathon.Discovery{
								Ports: &[]marathon.Port{
									{
										Number: 8898,
										Name:   "p1",
									}, {
										Number: 9999,
										Name:   "p2",
									},
								},
							},
						},
						Labels: &map[string]string{
							"traefik.portIndex": "0",
						},
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
				AppID: "specify-port-number",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID: "specify-port-number",
						Labels: &map[string]string{
							"traefik.port": "8080",
						},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "specify-port-index",
				Ports: []int{80, 443},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "specify-port-index",
						Ports: []int{80, 443},
						Labels: &map[string]string{
							"traefik.portIndex": "0",
						},
					},
				},
			},
			expected:         true,
			exposedByDefault: true,
		},
		{
			task: marathon.Task{
				AppID: "specify-out-of-range-port-index",
				Ports: []int{80, 443},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:    "specify-out-of-range-port-index",
						Ports: []int{80, 443},
						Labels: &map[string]string{
							"traefik.portIndex": "2",
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
				AppID: "foo",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "foo",
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
				AppID: "foo",
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
						ID:     "foo",
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
				AppID: "single-port",
				Ports: []int{80},
			},
			applications: &marathon.Applications{
				Apps: []marathon.Application{
					{
						ID:     "single-port",
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

	provider := &Marathon{}
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
		provider := &Marathon{
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
		provider := &Marathon{
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

	provider := &Marathon{}
	for _, c := range cases {
		actual := provider.applicationFilter(c.application, c.filteredTasks)
		if actual != c.expected {
			t.Fatalf("expected %v, got %v", c.expected, actual)
		}
	}
}

func TestMarathonGetPort(t *testing.T) {
	provider := &Marathon{}

	cases := []struct {
		applications []marathon.Application
		task         marathon.Task
		expected     string
	}{
		{
			applications: []marathon.Application{},
			task:         marathon.Task{},
			expected:     "",
		},
		{
			applications: []marathon.Application{
				{
					ID:     "test1",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "test2",
			},
			expected: "",
		},
		{
			applications: []marathon.Application{
				{
					ID:     "test1",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "test1",
				Ports: []int{80},
			},
			expected: "80",
		},
		{
			applications: []marathon.Application{
				{
					ID:     "multiple-ports-take-first",
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "multiple-ports-take-first",
				Ports: []int{80, 443},
			},
			expected: "80",
		},
		{
			applications: []marathon.Application{
				{
					ID: "specify-port-number",
					Labels: &map[string]string{
						"traefik.port": "443",
					},
				},
			},
			task: marathon.Task{
				AppID: "specify-port-number",
				Ports: []int{80, 443},
			},
			expected: "443",
		},
		{
			applications: []marathon.Application{
				{
					ID: "specify-port-index",
					Labels: &map[string]string{
						"traefik.portIndex": "1",
					},
				},
			},
			task: marathon.Task{
				AppID: "specify-port-index",
				Ports: []int{80, 443},
			},
			expected: "443",
		}, {
			applications: []marathon.Application{
				{
					ID:     "application-with-port",
					Ports:  []int{9999},
					Labels: &map[string]string{},
				},
			},
			task: marathon.Task{
				AppID: "application-with-port",
				Ports: []int{7777},
			},
			expected: "7777",
		},
	}

	for _, c := range cases {
		actual := provider.getPort(c.task, c.applications)
		if actual != c.expected {
			t.Fatalf("expected %q, got %q", c.expected, actual)
		}
	}
}

func TestMarathonGetWeigh(t *testing.T) {
	provider := &Marathon{}

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
	provider := &Marathon{
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
	provider := &Marathon{}

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
	provider := &Marathon{}

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
	provider := &Marathon{}

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
		provider := &Marathon{
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
	provider := &Marathon{}

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
	providerGroups := &Marathon{GroupsAsSubDomains: true}
	providerNoGroups := &Marathon{GroupsAsSubDomains: false}

	apps := []struct {
		path     string
		expected string
		provider *Marathon
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
