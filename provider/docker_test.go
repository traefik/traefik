package provider

import (
	"reflect"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	dockerclient "github.com/docker/engine-api/client"
	docker "github.com/docker/engine-api/types"
	dockertypes "github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/swarm"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
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
					Name: "mycontainer",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"com.docker.compose.project": "foo",
						"com.docker.compose.service": "bar",
					},
				},
			},
			expected: "Host-bar-foo-docker-localhost",
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
		}, {
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"com.docker.compose.project": "foo",
						"com.docker.compose.service": "bar",
					},
				},
			},
			expected: "Host:bar.foo.docker.localhost",
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
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"com.docker.compose.project": "foo",
						"com.docker.compose.service": "bar",
					},
				},
			},
			expected: "bar-foo",
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

func TestSwarmTaskParsing(t *testing.T) {
	cases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedNames map[string]string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
					},
				},
			},
			tasks: []swarm.Task{
				{
					ID:   "id1",
					Slot: 1,
				},
				{
					ID:   "id2",
					Slot: 2,
				},
				{
					ID:   "id3",
					Slot: 3,
				},
			},
			isGlobalSVC: false,
			expectedNames: map[string]string{
				"id1": "container.1",
				"id2": "container.2",
				"id3": "container.3",
			},
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
						Name: "container",
					},
				},
			},
			tasks: []swarm.Task{
				{
					ID: "id1",
				},
				{
					ID: "id2",
				},
				{
					ID: "id3",
				},
			},
			isGlobalSVC: true,
			expectedNames: map[string]string{
				"id1": "container.id1",
				"id2": "container.id2",
				"id3": "container.id3",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for _, e := range cases {
		dockerData := parseService(e.service, e.networks)

		for _, task := range e.tasks {
			taskDockerData := parseTasks(task, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)
			if !reflect.DeepEqual(taskDockerData.Name, e.expectedNames[task.ID]) {
				t.Fatalf("expect %v, got %v", e.expectedNames[task.ID], taskDockerData.Name)
			}
		}
	}
}

type fakeTasksClient struct {
	dockerclient.APIClient
	tasks []swarm.Task
	err   error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}

func TestListTasks(t *testing.T) {
	cases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedTasks []string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarm.Service{
				Spec: swarm.ServiceSpec{
					Annotations: swarm.Annotations{
						Name: "container",
					},
				},
			},
			tasks: []swarm.Task{
				{
					ID:   "id1",
					Slot: 1,
					Status: swarm.TaskStatus{
						State: swarm.TaskStateRunning,
					},
				},
				{
					ID:   "id2",
					Slot: 2,
					Status: swarm.TaskStatus{
						State: swarm.TaskStatePending,
					},
				},
				{
					ID:   "id3",
					Slot: 3,
				},
				{
					ID:   "id4",
					Slot: 4,
					Status: swarm.TaskStatus{
						State: swarm.TaskStateRunning,
					},
				},
				{
					ID:   "id5",
					Slot: 5,
					Status: swarm.TaskStatus{
						State: swarm.TaskStateFailed,
					},
				},
			},
			isGlobalSVC: false,
			expectedTasks: []string{
				"container.1",
				"container.4",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for _, e := range cases {
		dockerData := parseService(e.service, e.networks)
		dockerClient := &fakeTasksClient{tasks: e.tasks}
		taskDockerData, _ := listTasks(context.Background(), dockerClient, e.service.ID, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)

		if len(e.expectedTasks) != len(taskDockerData) {
			t.Fatalf("expected tasks %v, got %v", spew.Sprint(e.expectedTasks), spew.Sprint(taskDockerData))
		}

		for i, taskID := range e.expectedTasks {
			if taskDockerData[i].Name != taskID {
				t.Fatalf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
			}
		}
	}
}

func TestDockerGetServiceProtocol(t *testing.T) {
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
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.protocol": "https",
					},
				},
			},
			expected: "https",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.protocol": "https",
					},
				},
			},
			expected: "https",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServiceProtocol(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServiceWeight(t *testing.T) {
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
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.weight": "200",
					},
				},
			},
			expected: "200",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.weight": "31337",
					},
				},
			},
			expected: "31337",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServiceWeight(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServicePort(t *testing.T) {
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
			expected: "",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.port": "2500",
					},
				},
			},
			expected: "2500",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.port": "1234",
					},
				},
			},
			expected: "1234",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServicePort(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServiceFrontendRule(t *testing.T) {
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
			expected: "Host:foo.",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Path:/helloworld",
					},
				},
			},
			expected: "Path:/helloworld",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.frontend.rule": "Path:/mycustomservicepath",
					},
				},
			},
			expected: "Path:/mycustomservicepath",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServiceFrontendRule(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServiceBackend(t *testing.T) {
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
			expected: "foo-myservice",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.backend": "another-backend",
					},
				},
			},
			expected: "another-backend-myservice",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.frontend.backend": "custom-backend",
					},
				},
			},
			expected: "custom-backend",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServiceBackend(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServicePriority(t *testing.T) {
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
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.priority": "33",
					},
				},
			},
			expected: "33",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.frontend.priority": "2503",
					},
				},
			},
			expected: "2503",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServicePriority(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServicePassHostHeader(t *testing.T) {
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
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.passHostHeader": "false",
					},
				},
			},
			expected: "false",
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.frontend.passHostHeader": "false",
					},
				},
			},
			expected: "false",
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServicePassHostHeader(dockerData, "myservice")
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetServiceEntryPoints(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.ContainerJSON
		expected  []string
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "foo",
				},
				Config: &container.Config{},
			},
			expected: []string{},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "another",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.frontend.entryPoints": "http,https",
					},
				},
			},
			expected: []string{"http", "https"},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "test",
				},
				Config: &container.Config{
					Labels: map[string]string{
						"traefik.myservice.frontend.entryPoints": "http,https",
					},
				},
			},
			expected: []string{"http", "https"},
		},
	}

	for _, e := range containers {
		dockerData := parseContainer(e.container)
		actual := provider.getServiceEntryPoints(dockerData, "myservice")
		if !reflect.DeepEqual(actual, e.expected) {
			t.Fatalf("expected %q, got %q for container %q", e.expected, actual, dockerData.Name)
		}
	}
}

func TestDockerLoadDockerServiceConfig(t *testing.T) {
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
						Name: "foo",
					},
					Config: &container.Config{
						Labels: map[string]string{
							"traefik.service.port":                 "2503",
							"traefik.service.frontend.entryPoints": "http,https",
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
				"frontend-foo-service": {
					Backend:        "backend-foo-service",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Host:foo.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-service": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "http://127.0.0.1:2503",
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
							"traefik.service.port":                    "2503",
							"traefik.service.protocol":                "https",
							"traefik.service.weight":                  "80",
							"traefik.service.frontend.backend":        "foobar",
							"traefik.service.frontend.passHostHeader": "false",
							"traefik.service.frontend.rule":           "Path:/mypath",
							"traefik.service.frontend.priority":       "5000",
							"traefik.service.frontend.entryPoints":    "http,https,ws",
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
							"traefik.anotherservice.port":          "8079",
							"traefik.anotherservice.weight":        "33",
							"traefik.anotherservice.frontend.rule": "Path:/anotherpath",
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
				"frontend-foobar": {
					Backend:        "backend-foobar",
					PassHostHeader: false,
					Priority:       5000,
					EntryPoints:    []string{"http", "https", "ws"},
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Path:/mypath",
						},
					},
				},
				"frontend-test2-anotherservice": {
					Backend:        "backend-test2-anotherservice",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"service-anotherservice": {
							Rule: "Path:/anotherpath",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-test2-anotherservice": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "http://127.0.0.1:8079",
							Weight: 33,
						},
					},
					CircuitBreaker: nil,
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
