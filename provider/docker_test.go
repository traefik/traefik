package provider

import (
	"reflect"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
)

func TestDockerGetFrontendName(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Host-foo-docker-localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
					},
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: "Host-foo-bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: "Path-test",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "PathPrefix:/test2",
					},
				},
			},
			expected: "PathPrefix-test2",
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendName(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Host:foo.docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: "Host:bar.docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
			},
			expected: "Host:foo.bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/test",
					},
				},
			},
			expected: "Path:/test",
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendRule(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetBackend(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "foo",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
			},
			expected: "bar",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.backend": "foobar",
					},
				},
			},
			expected: "foobar",
		},
	}

	for _, e := range containers {
		actual := provider.getBackend(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetPort(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "80",
		},
		// FIXME handle this better..
		// {
		// 	container: docker.ContainerJSON{
		// 		Name:   "bar",
		// 		Config: &container.Config{},
		// 		NetworkSettings: &docker.NetworkSettings{
		// 			Ports: map[docker.Port][]docker.PortBinding{
		// 				"80/tcp":  []docker.PortBinding{},
		// 				"443/tcp": []docker.PortBinding{},
		// 			},
		// 		},
		// 	},
		// 	expected: "80",
		// },
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: "8080",
		},
	}

	for _, e := range containers {
		actual := provider.getPort(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetWeight(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "1",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.weight": "10",
					},
				},
			},
			expected: "10",
		},
	}

	for _, e := range containers {
		actual := provider.getWeight(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetDomain(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "docker.localhost",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
				},
			},
			expected: "foo.bar",
		},
	}

	for _, e := range containers {
		actual := provider.getDomain(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetProtocol(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "http",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			expected: "https",
		},
	}

	for _, e := range containers {
		actual := provider.getProtocol(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetPassHostHeader(t *testing.T) {
	provider := &Docker{}
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "true",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.passHostHeader": "false",
					},
				},
			},
			expected: "false",
		},
	}

	for _, e := range containers {
		actual := provider.getPassHostHeader(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetLabel(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				Config: &container.Config{},
			},
			expected: "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
			expected: "",
		},
	}

	for _, e := range containers {
		label, err := getLabel(e.container, "foo")
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

func TestDockerGetLabels(t *testing.T) {
	containers := []struct {
		container      docker.ContainerJSON
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container: docker.ContainerJSON{
				Config: &container.Config{},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
		},
		{
			container: docker.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"foo": "fooz",
						"bar": "barz",
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
		},
	}

	for _, e := range containers {
		labels, err := getLabels(e.container, []string{"foo", "bar"})
		if !reflect.DeepEqual(labels, e.expectedLabels) {
			t.Fatalf("expect %v, got %v", e.expectedLabels, labels)
		}
		if e.expectedError != "" {
			if err == nil || !strings.Contains(err.Error(), e.expectedError) {
				t.Fatalf("expected an error with %q, got %v", e.expectedError, err)
			}
		}
	}
}

func TestDockerTraefikFilter(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  bool
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.enable": "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, e := range containers {
		actual := containerFilter(e.container)
		if actual != e.expected {
			t.Fatalf("expected %v for %+v, got %+v", e.expected, e, actual)
		}
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
	cases := []struct {
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test",
					},
					Config: &container.Config{},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost": {
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
							Weight: 1,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test1",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend":              "foobar",
							"traefik.frontend.entryPoints": "http,https",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
				{
					ContainerJSONBase: &docker.ContainerJSONBase{
						Name: "test2",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.backend": "foobar",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						NetworkSettingsBase: docker.NetworkSettingsBase{
							Ports: nat.PortMap{
								"80/tcp": {},
							},
						},
						Networks: map[string]*network.EndpointSettings{
							"bridge": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost": {
							Rule: "Host:test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 1,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 1,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
		},
	}

	provider := &Docker{
		Domain: "docker.localhost",
	}

	for _, c := range cases {
		actualConfig := provider.loadDockerConfig(c.containers)
		// Compare backends
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
