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

	swarm "github.com/docker/engine-api/types/swarm"
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
		dockerData := parseContainer(e.container)
		actual := provider.getFrontendName(dockerData)
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
		dockerData := parseContainer(e.container)
		actual := provider.getFrontendRule(dockerData)
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
		dockerData := parseContainer(e.container)
		actual := provider.getBackend(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetIPAddress(t *testing.T) { // TODO
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"nottestnet": {
							IPAddress: "10.11.12.13",
						},
					},
				},
			},
			expected: "10.11.12.13",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.docker.network": "testnet2",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "10.11.12.14",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "bar",
					HostConfig: &container.HostConfig{
						NetworkMode: "host",
					},
				},
				Config: &container.Config{
					Labels: map[string]string{},
				},
				NetworkSettings: &docker.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"testnet1": {
							IPAddress: "10.11.12.13",
						},
						"testnet2": {
							IPAddress: "10.11.12.14",
						},
					},
				},
			},
			expected: "127.0.0.1",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getIPAddress(dockerData)
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
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "8080",
		},
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
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test-multi-ports",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"8080/tcp": {},
							"80/tcp":   {},
						},
					},
				},
			},
			expected: "8080",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getPort(dockerData)
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
			expected: "0",
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
		dockerData := parseContainer(e.container)
		actual := provider.getWeight(dockerData)
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
		dockerData := parseContainer(e.container)
		actual := provider.getDomain(dockerData)
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
		dockerData := parseContainer(e.container)
		actual := provider.getProtocol(dockerData)
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
		dockerData := parseContainer(e.container)
		actual := provider.getPassHostHeader(dockerData)
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
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
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
		dockerData := parseContainer(e.container)
		label, err := getLabel(dockerData, "foo")
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
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
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
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
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
		dockerData := parseContainer(e.container)
		labels, err := getLabels(dockerData, []string{"foo", "bar"})
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
	provider := Docker{}
	containers := []struct {
		container        docker.ContainerJSON
		exposedByDefault bool
		expected         bool
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			exposedByDefault: true,
			expected:         false,
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
			exposedByDefault: true,
			expected:         false,
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
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container-multi-ports",
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: true,
			expected:         true,
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
			exposedByDefault: false,
			expected:         false,
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
			exposedByDefault: false,
			expected:         true,
		},
	}

	for _, e := range containers {
		provider.ExposedByDefault = e.exposedByDefault
		dockerData := parseContainer(e.container)
		actual := provider.containerFilter(dockerData)
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
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
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
							Weight: 0,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
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
							"traefik.backend":                           "foobar",
							"traefik.frontend.entryPoints":              "http,https",
							"traefik.backend.maxconn.amount":            "1000",
							"traefik.backend.maxconn.extractorfunc":     "somethingelse",
							"traefik.backend.loadbalancer.method":       "drr",
							"traefik.backend.circuitbreaker.expression": "NetworkErrorRatio() > 0.5",
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
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
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
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "somethingelse",
					},
				},
			},
		},
	}

	provider := &Docker{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}

	for _, c := range cases {
		var dockerDataList []dockerData
		for _, container := range c.containers {
			dockerData := parseContainer(container)
			dockerDataList = append(dockerDataList, dockerData)
		}

		actualConfig := provider.loadDockerConfig(dockerDataList)
		// Compare backends
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}

func TestSwarmGetFrontendName(t *testing.T) {
	provider := &Docker{
		Domain:    "docker.localhost",
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "foo",
					},
				},
			},
			expected: "Host-foo-docker-localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
						},
					},
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.frontend.rule": "Host:foo.bar",
						},
					},
				},
			},
			expected: "Host-foo-bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.frontend.rule": "Path:/test",
						},
					},
				},
			},
			expected: "Path-test",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.frontend.rule": "PathPrefix:/test2",
						},
					},
				},
			},
			expected: "PathPrefix-test2",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getFrontendName(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetFrontendRule(t *testing.T) {
	provider := &Docker{
		Domain:    "docker.localhost",
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "foo",
					},
				},
			},
			expected: "Host:foo.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
				},
			},
			expected: "Host:bar.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.frontend.rule": "Host:foo.bar",
						},
					},
				},
			},
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.frontend.rule": "Path:/test",
						},
					},
				},
			},
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getFrontendRule(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetBackend(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "foo",
					},
				},
			},
			expected: "foo",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
				},
			},
			expected: "bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"traefik.backend": "foobar",
						},
					},
				},
			},
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getBackend(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeVIP,
					},
				},
				Endpoint: swarm.Endpoint{
					VirtualIPs: []swarm.EndpointVirtualIP{
						{
							NetworkID: "1",
							Addr:      "10.11.12.13/24",
						},
					},
				},
			},
			expected: "10.11.12.13",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.docker.network": "barnet",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeVIP,
					},
				},
				Endpoint: swarm.Endpoint{
					VirtualIPs: []swarm.EndpointVirtualIP{
						{
							NetworkID: "1",
							Addr:      "10.11.12.13/24",
						},
						{
							NetworkID: "2",
							Addr:      "10.11.12.99/24",
						},
					},
				},
			},
			expected: "10.11.12.99",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getIPAddress(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetPort(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.port": "8080",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "8080",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getPort(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetWeight(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.weight": "10",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "10",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getWeight(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetDomain(t *testing.T) {
	provider := &Docker{
		Domain:    "docker.localhost",
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "foo",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.domain": "foo.bar",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getDomain(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetProtocol(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "http",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.protocol": "https",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "https",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getProtocol(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetPassHostHeader(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "true",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
						Labels: map[string]string{
							"traefik.frontend.passHostHeader": "false",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "false",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		actual := provider.getPassHostHeader(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestSwarmGetLabel(t *testing.T) {

	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "bar",
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "Label not found:",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					EndpointSpec: &swarm.EndpointSpec{
						Mode: swarm.ResolutionModeDNSRR,
					},
				},
			},
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		label, err := getLabel(dockerData, "foo")
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

func TestSwarmGetLabels(t *testing.T) {
	services := []struct {
		service        swarm.Service
		expectedLabels map[string]string
		expectedError  string
		networks       map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
					},
				},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
			networks:       map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"foo": "fooz",
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
			networks:      map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "test",
						Labels: map[string]string{
							"foo": "fooz",
							"bar": "barz",
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
			networks:      map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		labels, err := getLabels(dockerData, []string{"foo", "bar"})
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

func TestSwarmTraefikFilter(t *testing.T) {
	provider := &Docker{
		SwarmMode: true,
	}
	services := []struct {
		service          swarm.Service
		exposedByDefault bool
		expected         bool
		networks         map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
					},
				},
			},
			exposedByDefault: true,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.enable": "false",
							"traefik.port":   "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.frontend.rule": "Host:foo.bar",
							"traefik.port":          "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.port": "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.enable": "true",
							"traefik.port":   "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.enable": "anything",
							"traefik.port":   "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.frontend.rule": "Host:foo.bar",
							"traefik.port":          "80",
						},
					},
				},
			},
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.port": "80",
						},
					},
				},
			},
			exposedByDefault: false,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
						Labels: map[string]string{
							"traefik.enable": "true",
							"traefik.port":   "80",
						},
					},
				},
			},
			exposedByDefault: false,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
	}

	for _, e := range services {
		dockerData := parseService(e.service, e.networks)
		provider.ExposedByDefault = e.exposedByDefault
		actual := provider.containerFilter(dockerData)
		if actual != e.expected {
			t.Fatalf("expected %v for %+v, got %+v", e.expected, e, actual)
		}
	}
}

func TestSwarmLoadDockerConfig(t *testing.T) {
	cases := []struct {
		services          []swarm.Service
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
		networks          map[string]*docker.NetworkResource
	}{
		{
			services:          []swarm.Service{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
			networks:          map[string]*docker.NetworkResource{},
		},
		{
			services: []swarm.Service{
				{
					Spec: swarm.ServiceSpec{
						Annotations: swarm.Annotations{
							Name: "test",
							Labels: map[string]string{
								"traefik.port": "80",
							},
						},
						EndpointSpec: &swarm.EndpointSpec{
							Mode: swarm.ResolutionModeVIP,
						},
					},
					Endpoint: swarm.Endpoint{
						VirtualIPs: []swarm.EndpointVirtualIP{
							{
								Addr:      "127.0.0.1/24",
								NetworkID: "1",
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
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			services: []swarm.Service{
				{
					Spec: swarm.ServiceSpec{
						Annotations: swarm.Annotations{
							Name: "test1",
							Labels: map[string]string{
								"traefik.port":                 "80",
								"traefik.backend":              "foobar",
								"traefik.frontend.entryPoints": "http,https",
							},
						},
						EndpointSpec: &swarm.EndpointSpec{
							Mode: swarm.ResolutionModeVIP,
						},
					},
					Endpoint: swarm.Endpoint{
						VirtualIPs: []swarm.EndpointVirtualIP{
							{
								Addr:      "127.0.0.1/24",
								NetworkID: "1",
							},
						},
					},
				},
				{
					Spec: swarm.ServiceSpec{
						Annotations: swarm.Annotations{
							Name: "test2",
							Labels: map[string]string{
								"traefik.port":    "80",
								"traefik.backend": "foobar",
							},
						},
						EndpointSpec: &swarm.EndpointSpec{
							Mode: swarm.ResolutionModeVIP,
						},
					},
					Endpoint: swarm.Endpoint{
						VirtualIPs: []swarm.EndpointVirtualIP{
							{
								Addr:      "127.0.0.1/24",
								NetworkID: "1",
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
							Weight: 0,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	provider := &Docker{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
		SwarmMode:        true,
	}

	for _, c := range cases {
		var dockerDataList []dockerData
		for _, service := range c.services {
			dockerData := parseService(service, c.networks)
			dockerDataList = append(dockerDataList, dockerData)
		}

		actualConfig := provider.loadDockerConfig(dockerDataList)
		// Compare backends
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
