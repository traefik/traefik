package rancher

import (
	"github.com/containous/traefik/types"
	"reflect"
	"strings"
	"testing"
)

func TestRancherGetFrontendName(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "foo",
			},
			expected: "Host-foo-rancher-localhost",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
				},
			},

			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "Host:foo.bar",
				},
			},

			expected: "Host-foo-bar",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "Path:/test",
				},
			},

			expected: "Path-test",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "PathPrefix:/test2",
				},
			},

			expected: "PathPrefix-test2",
		},
	}

	for _, e := range services {
		actual := provider.getFrontendName(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetFrontendRule(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "foo",
			},
			expected: "Host:foo.rancher.localhost",
		},
		{
			service: rancherData{
				Name: "foo/bar",
			},
			expected: "Host:foo.bar.rancher.localhost",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "Host:foo.bar.com",
				},
			},

			expected: "Host:foo.bar.com",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "Path:/test",
				},
			},

			expected: "Path:/test",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.rule": "PathPrefix:/test2",
				},
			},

			expected: "PathPrefix:/test2",
		},
	}

	for _, e := range services {
		actual := provider.getFrontendRule(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetBackend(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "test-service",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.backend": "foobar",
				},
			},

			expected: "foobar",
		},
	}

	for _, e := range services {
		actual := provider.getBackend(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetWeight(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "0",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.weight": "5",
				},
			},

			expected: "5",
		},
	}

	for _, e := range services {
		actual := provider.getWeight(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetPort(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.port": "1337",
				},
			},

			expected: "1337",
		},
	}

	for _, e := range services {
		actual := provider.getPort(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetDomain(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "rancher.localhost",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.domain": "foo.bar",
				},
			},

			expected: "foo.bar",
		},
	}

	for _, e := range services {
		actual := provider.getDomain(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetProtocol(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "http",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.protocol": "https",
				},
			},

			expected: "https",
		},
	}

	for _, e := range services {
		actual := provider.getProtocol(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetPassHostHeader(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
	}

	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "true",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"traefik.frontend.passHostHeader": "false",
				},
			},

			expected: "false",
		},
	}

	for _, e := range services {
		actual := provider.getPassHostHeader(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestRancherGetLabel(t *testing.T) {
	services := []struct {
		service  rancherData
		expected string
	}{
		{
			service: rancherData{
				Name: "test-service",
			},
			expected: "Label not found",
		},
		{
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"foo": "bar",
				},
			},

			expected: "",
		},
	}

	for _, e := range services {
		label, err := getServiceLabel(e.service, "foo")
		if e.expected != "" {
			if err == nil || !strings.Contains(err.Error(), e.expected) {
				t.Fatalf("expected an error with %q, got %v", e.expected, err)
			}
		} else {
			if label != "bar" {
				t.Fatalf("expected label 'bar', got %s", label)
			}
		}
	}
}

func TestRancherLoadRancherConfig(t *testing.T) {
	cases := []struct {
		services          []rancherData
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			services:          []rancherData{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						"traefik.port":                "80",
						"traefik.frontend.auth.basic": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Priority:       0,

					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	provider := &Provider{
		Domain:           "rancher.localhost",
		ExposedByDefault: true,
	}

	for _, c := range cases {
		var rancherDataList []rancherData
		for _, service := range c.services {
			rancherDataList = append(rancherDataList, service)
		}

		actualConfig := provider.loadRancherConfig(rancherDataList)

		// Compare backends
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
