package docker

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/docker/cli/cli/connhelper"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-connections/sockets"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/traefik/traefik/v3/pkg/version"
)

// DefaultTemplateRule The default template for the default rule.
const DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

type Shared struct {
	ExposedByDefault   bool   `description:"Expose containers by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	Constraints        string `description:"Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	AllowEmptyServices bool   `description:"Disregards the Docker containers health checks with respect to the creation or removal of the corresponding services." json:"allowEmptyServices,omitempty" toml:"allowEmptyServices,omitempty" yaml:"allowEmptyServices,omitempty" export:"true"`
	Network            string `description:"Default Docker network used." json:"network,omitempty" toml:"network,omitempty" yaml:"network,omitempty" export:"true"`
	UseBindPortIP      bool   `description:"Use the ip address from the bound port, rather than from the inner network." json:"useBindPortIP,omitempty" toml:"useBindPortIP,omitempty" yaml:"useBindPortIP,omitempty" export:"true"`

	Watch       bool   `description:"Watch Docker events." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	DefaultRule string `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`

	defaultRuleTpl *template.Template
}

func inspectContainers(ctx context.Context, dockerClient client.ContainerAPIClient, containerID string) dockerData {
	containerInspected, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("Failed to inspect container %s", containerID)
		return dockerData{}
	}

	// This condition is here to avoid to have empty IP https://github.com/traefik/traefik/issues/2459
	// We register only container which are running
	if containerInspected.ContainerJSONBase != nil && containerInspected.ContainerJSONBase.State != nil && containerInspected.ContainerJSONBase.State.Running {
		return parseContainer(containerInspected)
	}

	return dockerData{}
}

func parseContainer(container dockertypes.ContainerJSON) dockerData {
	dData := dockerData{
		NetworkSettings: networkSettings{},
	}

	if container.ContainerJSONBase != nil {
		dData.ID = container.ContainerJSONBase.ID
		dData.Name = container.ContainerJSONBase.Name
		dData.ServiceName = dData.Name // Default ServiceName to be the container's Name.
		dData.Node = container.ContainerJSONBase.Node

		if container.ContainerJSONBase.HostConfig != nil {
			dData.NetworkSettings.NetworkMode = container.ContainerJSONBase.HostConfig.NetworkMode
		}

		if container.State != nil && container.State.Health != nil {
			dData.Health = container.State.Health.Status
		}
	}

	if container.Config != nil && container.Config.Labels != nil {
		dData.Labels = container.Config.Labels
	}

	if container.NetworkSettings != nil {
		if container.NetworkSettings.Ports != nil {
			dData.NetworkSettings.Ports = container.NetworkSettings.Ports
		}
		if container.NetworkSettings.Networks != nil {
			dData.NetworkSettings.Networks = make(map[string]*networkData)
			for name, containerNetwork := range container.NetworkSettings.Networks {
				addr := containerNetwork.IPAddress
				if addr == "" {
					addr = containerNetwork.GlobalIPv6Address
				}

				dData.NetworkSettings.Networks[name] = &networkData{
					ID:   containerNetwork.NetworkID,
					Name: name,
					Addr: addr,
				}
			}
		}
	}
	return dData
}

type ClientConfig struct {
	apiVersion string

	Username          string           `description:"Username for Basic HTTP authentication." json:"username,omitempty" toml:"username,omitempty" yaml:"username,omitempty"`
	Password          string           `description:"Password for Basic HTTP authentication." json:"password,omitempty" toml:"password,omitempty" yaml:"password,omitempty"`
	Endpoint          string           `description:"Docker server endpoint. Can be a TCP or a Unix socket endpoint." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	TLS               *types.ClientTLS `description:"Enable Docker TLS support." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" export:"true"`
	HTTPClientTimeout ptypes.Duration  `description:"Client timeout for HTTP connections." json:"httpClientTimeout,omitempty" toml:"httpClientTimeout,omitempty" yaml:"httpClientTimeout,omitempty" export:"true"`
}

func createClient(ctx context.Context, cfg ClientConfig) (*client.Client, error) {
	opts, err := getClientOpts(ctx, cfg)
	if err != nil {
		return nil, err
	}

	httpHeaders := map[string]string{
		"User-Agent": "Traefik " + version.Version,
	}
	if cfg.Username != "" && cfg.Password != "" {
		httpHeaders["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(cfg.Username+":"+cfg.Password))
	}

	opts = append(opts,
		client.WithHTTPHeaders(httpHeaders),
		client.WithVersion(cfg.apiVersion))

	return client.NewClientWithOpts(opts...)
}

func getClientOpts(ctx context.Context, cfg ClientConfig) ([]client.Opt, error) {
	helper, err := connhelper.GetConnectionHelper(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	// SSH
	if helper != nil {
		// https://github.com/docker/cli/blob/ebca1413117a3fcb81c89d6be226dcec74e5289f/cli/context/docker/load.go#L112-L123

		httpClient := &http.Client{
			Transport: &http.Transport{
				DialContext: helper.Dialer,
			},
		}

		return []client.Opt{
			client.WithHTTPClient(httpClient),
			client.WithTimeout(time.Duration(cfg.HTTPClientTimeout)),
			client.WithHost(helper.Host), // To avoid 400 Bad Request: malformed Host header daemon error
			client.WithDialContext(helper.Dialer),
		}, nil
	}

	opts := []client.Opt{
		client.WithHost(cfg.Endpoint),
		client.WithTimeout(time.Duration(cfg.HTTPClientTimeout)),
	}

	if cfg.TLS != nil {
		conf, err := cfg.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to create client TLS configuration: %w", err)
		}

		hostURL, err := client.ParseHostURL(cfg.Endpoint)
		if err != nil {
			return nil, err
		}

		tr := &http.Transport{
			TLSClientConfig: conf,
		}

		if err := sockets.ConfigureTransport(tr, hostURL.Scheme, hostURL.Host); err != nil {
			return nil, err
		}

		opts = append(opts, client.WithHTTPClient(&http.Client{Transport: tr, Timeout: time.Duration(cfg.HTTPClientTimeout)}))
	}

	return opts, nil
}

func getPort(container dockerData, serverPort string) string {
	if len(serverPort) > 0 {
		return serverPort
	}

	var ports []nat.Port
	for port := range container.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	less := func(i, j nat.Port) bool {
		return i.Int() < j.Int()
	}
	nat.Sort(ports, less)

	if len(ports) > 0 {
		return ports[0].Port()
	}

	return ""
}

func getServiceName(container dockerData) string {
	serviceName := container.ServiceName

	if values, err := getStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		serviceName = values[labelDockerComposeService] + "_" + values[labelDockerComposeProject]
	}

	return provider.Normalize(serviceName)
}
