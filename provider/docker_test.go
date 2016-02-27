package provider

import (
	"reflect"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/fsouza/go-dockerclient"
)

func TestDockerGetFrontendName(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "Host-foo-docker-localhost",
		},
		{
			container: docker.Container{
				Name: "bar",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Header",
					},
				},
			},
			expected: "Header-bar-docker-localhost",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.value": "foo.bar",
					},
				},
			},
			expected: "Host-foo-bar",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.value": "foo.bar",
						"traefik.frontend.rule":  "Header",
					},
				},
			},
			expected: "Header-foo-bar",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.value": "[foo.bar]",
						"traefik.frontend.rule":  "Header",
					},
				},
			},
			expected: "Header-foo-bar",
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendName(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetFrontendValue(t *testing.T) {
	provider := &Docker{
		Domain: "docker.localhost",
	}

	containers := []struct {
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "foo.docker.localhost",
		},
		{
			container: docker.Container{
				Name:   "bar",
				Config: &docker.Config{},
			},
			expected: "bar.docker.localhost",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.value": "foo.bar",
					},
				},
			},
			expected: "foo.bar",
		},
	}

	for _, e := range containers {
		actual := provider.getFrontendValue(e.container)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	provider := &Docker{}

	containers := []struct {
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "Host",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "foo",
					},
				},
			},
			expected: "foo",
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "foo",
		},
		{
			container: docker.Container{
				Name:   "bar",
				Config: &docker.Config{},
			},
			expected: "bar",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:            "foo",
				Config:          &docker.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: "",
		},
		{
			container: docker.Container{
				Name:   "bar",
				Config: &docker.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: "80",
		},
		// FIXME handle this better..
		// {
		// 	container: docker.Container{
		// 		Name:   "bar",
		// 		Config: &docker.Config{},
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
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.port": "8080",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "0",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "docker.localhost",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "http",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Name:   "foo",
				Config: &docker.Config{},
			},
			expected: "false",
		},
		{
			container: docker.Container{
				Name: "test",
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.passHostHeader": "true",
					},
				},
			},
			expected: "true",
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
		container docker.Container
		expected  string
	}{
		{
			container: docker.Container{
				Config: &docker.Config{},
			},
			expected: "Label not found:",
		},
		{
			container: docker.Container{
				Config: &docker.Config{
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
		container      docker.Container
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container: docker.Container{
				Config: &docker.Config{},
			},
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: docker.Container{
				Config: &docker.Config{
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
			container: docker.Container{
				Config: &docker.Config{
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
		container docker.Container
		expected  bool
	}{
		{
			container: docker.Container{
				Config:          &docker.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: false,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.enable": "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.rule": "Host",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.value": "foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.Container{
				Config: &docker.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp":  {},
						"443/tcp": {},
					},
				},
			},
			expected: false,
		},
		{
			container: docker.Container{
				Config: &docker.Config{},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.port": "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp":  {},
						"443/tcp": {},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.enable": "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.enable": "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: true,
		},
		{
			container: docker.Container{
				Config: &docker.Config{
					Labels: map[string]string{
						"traefik.frontend.rule":  "Host",
						"traefik.frontend.value": "foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					Ports: map[docker.Port][]docker.PortBinding{
						"80/tcp": {},
					},
				},
			},
			expected: true,
		},
	}

	for _, e := range containers {
		actual := containerFilter(e.container)
		if actual != e.expected {
			t.Fatalf("expected %v, got %v", e.expected, actual)
		}
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
	cases := []struct {
		containers        []docker.Container
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			containers:        []docker.Container{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			containers: []docker.Container{
				{
					Name:   "test",
					Config: &docker.Config{},
					NetworkSettings: &docker.NetworkSettings{
						Ports: map[docker.Port][]docker.PortBinding{
							"80/tcp": {},
						},
						Networks: map[string]docker.ContainerNetwork{
							"bridgde": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`"frontend-Host-test-docker-localhost"`: {
					Backend:     "backend-test",
					EntryPoints: []string{},
					Routes: map[string]types.Route{
						`"route-frontend-Host-test-docker-localhost"`: {
							Rule:  "Host",
							Value: "test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL: "http://127.0.0.1:80",
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
		},
		{
			containers: []docker.Container{
				{
					Name: "test1",
					Config: &docker.Config{
						Labels: map[string]string{
							"traefik.backend":              "foobar",
							"traefik.frontend.entryPoints": "http,https",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						Ports: map[docker.Port][]docker.PortBinding{
							"80/tcp": {},
						},
						Networks: map[string]docker.ContainerNetwork{
							"bridgde": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
				{
					Name: "test2",
					Config: &docker.Config{
						Labels: map[string]string{
							"traefik.backend": "foobar",
						},
					},
					NetworkSettings: &docker.NetworkSettings{
						Ports: map[docker.Port][]docker.PortBinding{
							"80/tcp": {},
						},
						Networks: map[string]docker.ContainerNetwork{
							"bridgde": {
								IPAddress: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				`"frontend-Host-test1-docker-localhost"`: {
					Backend:     "backend-foobar",
					EntryPoints: []string{"http", "https"},
					Routes: map[string]types.Route{
						`"route-frontend-Host-test1-docker-localhost"`: {
							Rule:  "Host",
							Value: "test1.docker.localhost",
						},
					},
				},
				`"frontend-Host-test2-docker-localhost"`: {
					Backend:     "backend-foobar",
					EntryPoints: []string{},
					Routes: map[string]types.Route{
						`"route-frontend-Host-test2-docker-localhost"`: {
							Rule:  "Host",
							Value: "test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL: "http://127.0.0.1:80",
						},
						"server-test2": {
							URL: "http://127.0.0.1:80",
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
