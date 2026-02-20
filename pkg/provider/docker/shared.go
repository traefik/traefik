package docker

import (
	"cmp"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"text/template"
	"time"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/go-connections/sockets"
	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
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
	res, err := dockerClient.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("Failed to inspect container %s", containerID)
		return dockerData{}
	}

	// Always parse all containers (running and stopped)
	// The allowNonRunning filtering will be applied later in service configuration
	if res.Container.State != nil {
		return parseContainer(res.Container)
	}

	return dockerData{}
}

func parseContainer(container containertypes.InspectResponse) dockerData {
	dData := dockerData{
		ID:              container.ID,
		ServiceName:     container.Name, // Default ServiceName to be the container's Name.
		Name:            container.Name,
		Status:          container.State.Status,
		NetworkSettings: networkSettings{},
	}

	if container.HostConfig != nil {
		dData.NetworkSettings.NetworkMode = container.HostConfig.NetworkMode
	}

	if container.State != nil && container.State.Health != nil {
		dData.Health = container.State.Health.Status
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
				var addr string
				if containerNetwork.IPAddress.IsValid() {
					addr = containerNetwork.IPAddress.String()
				} else if containerNetwork.GlobalIPv6Address.IsValid() {
					addr = containerNetwork.GlobalIPv6Address.String()
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
		client.FromEnv,
		client.WithHTTPHeaders(httpHeaders))

	return client.New(opts...)
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
	if len(container.NetworkSettings.Ports) == 0 {
		return ""
	}

	var ports []networktypes.Port
	for port := range container.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	slices.SortFunc(ports, func(a, b networktypes.Port) int {
		return cmp.Compare(a.Num(), b.Num())
	})

	return strconv.Itoa(int(ports[0].Num()))
}

func getServiceName(container dockerData) string {
	serviceName := container.ServiceName

	if values, err := getStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		serviceName = values[labelDockerComposeService] + "_" + values[labelDockerComposeProject]
	}

	return provider.Normalize(serviceName)
}
