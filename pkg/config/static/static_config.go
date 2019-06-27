package static

import (
	"errors"
	"strings"
	"time"

	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/ping"
	acmeprovider "github.com/containous/traefik/pkg/provider/acme"
	"github.com/containous/traefik/pkg/provider/docker"
	"github.com/containous/traefik/pkg/provider/file"
	"github.com/containous/traefik/pkg/provider/kubernetes/crd"
	"github.com/containous/traefik/pkg/provider/kubernetes/ingress"
	"github.com/containous/traefik/pkg/provider/marathon"
	"github.com/containous/traefik/pkg/provider/rancher"
	"github.com/containous/traefik/pkg/provider/rest"
	"github.com/containous/traefik/pkg/tls"
	"github.com/containous/traefik/pkg/tracing/datadog"
	"github.com/containous/traefik/pkg/tracing/haystack"
	"github.com/containous/traefik/pkg/tracing/instana"
	"github.com/containous/traefik/pkg/tracing/jaeger"
	"github.com/containous/traefik/pkg/tracing/zipkin"
	"github.com/containous/traefik/pkg/types"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/go-acme/lego/challenge/dns01"
)

const (
	// DefaultInternalEntryPointName the name of the default internal entry point
	DefaultInternalEntryPointName = "traefik"

	// DefaultGraceTimeout controls how long Traefik serves pending requests
	// prior to shutting down.
	DefaultGraceTimeout = 10 * time.Second

	// DefaultIdleTimeout before closing an idle connection.
	DefaultIdleTimeout = 180 * time.Second

	// DefaultAcmeCAServer is the default ACME API endpoint
	DefaultAcmeCAServer = "https://acme-v02.api.letsencrypt.org/directory"
)

// Configuration is the static configuration
type Configuration struct {
	Global *Global `description:"Global configuration options" export:"true"`

	ServersTransport *ServersTransport `description:"Servers default transport." export:"true"`
	EntryPoints      EntryPoints       `description:"Entry points definition." export:"true"`
	Providers        *Providers        `description:"Providers configuration." export:"true"`

	API     *API           `description:"Enable api/dashboard." export:"true" label:"allowEmpty"`
	Metrics *types.Metrics `description:"Enable a metrics exporter." export:"true"`
	Ping    *ping.Handler  `description:"Enable ping." export:"true" label:"allowEmpty"`
	// Rest    *rest.Provider `description:"Enable Rest backend with default settings" export:"true"`

	Log       *types.TraefikLog `description:"Traefik log settings." export:"true" label:"allowEmpty"`
	AccessLog *types.AccessLog  `description:"Access log settings." export:"true" label:"allowEmpty"`
	Tracing   *Tracing          `description:"OpenTracing configuration." export:"true" label:"allowEmpty"`

	HostResolver *types.HostResolverConfig `description:"Enable CNAME Flattening." export:"true" label:"allowEmpty"`

	ACME *acmeprovider.Configuration `description:"Enable ACME (Let's Encrypt): automatic SSL." export:"true"`
}

// Global holds the global configuration.
type Global struct {
	CheckNewVersion    bool  `description:"Periodically check if a new version has been released." export:"true"`
	SendAnonymousUsage *bool `description:"Periodically send anonymous usage statistics. If the option is not specified, it will be enabled by default." export:"true"`
}

// ServersTransport options to configure communication between Traefik and the servers
type ServersTransport struct {
	InsecureSkipVerify  bool                `description:"Disable SSL certificate verification." export:"true"`
	RootCAs             []tls.FileOrContent `description:"Add cert file for self-signed certificate."`
	MaxIdleConnsPerHost int                 `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts `description:"Timeouts for requests forwarded to the backend servers." export:"true"`
}

// API holds the API configuration
type API struct {
	EntryPoint      string            `description:"The entry point that the API handler will be bound to." export:"true"`
	Dashboard       bool              `description:"Activate dashboard." export:"true"`
	Debug           bool              `description:"Enable additional endpoints for debugging and profiling." export:"true"`
	Statistics      *types.Statistics `description:"Enable more detailed statistics." export:"true" label:"allowEmpty"`
	Middlewares     []string          `description:"Middleware list." export:"true"`
	DashboardAssets *assetfs.AssetFS  `json:"-" label:"-"`
}

// SetDefaults sets the default values.
func (a *API) SetDefaults() {
	a.EntryPoint = "traefik"
	a.Dashboard = true
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  types.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set." export:"true"`
	WriteTimeout types.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set." export:"true"`
	IdleTimeout  types.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout is set." export:"true"`
}

// SetDefaults sets the default values.
func (a *RespondingTimeouts) SetDefaults() {
	a.IdleTimeout = types.Duration(DefaultIdleTimeout)
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           types.Duration `description:"The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." export:"true"`
	ResponseHeaderTimeout types.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists." export:"true"`
	IdleConnTimeout       types.Duration `description:"The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself" export:"true"`
}

// SetDefaults sets the default values.
func (f *ForwardingTimeouts) SetDefaults() {
	f.DialTimeout = types.Duration(30 * time.Second)
	f.IdleConnTimeout = types.Duration(90 * time.Second)
}

// LifeCycle contains configurations relevant to the lifecycle (such as the shutdown phase) of Traefik.
type LifeCycle struct {
	RequestAcceptGraceTimeout types.Duration `description:"Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure."`
	GraceTimeOut              types.Duration `description:"Duration to give active requests a chance to finish before Traefik stops."`
}

// SetDefaults sets the default values.
func (a *LifeCycle) SetDefaults() {
	a.GraceTimeOut = types.Duration(DefaultGraceTimeout)
}

// Tracing holds the tracing configuration.
type Tracing struct {
	ServiceName   string           `description:"Set the name for this service." export:"true"`
	SpanNameLimit int              `description:"Set the maximum character limit for Span names (default 0 = no limit)." export:"true"`
	Jaeger        *jaeger.Config   `description:"Settings for jaeger." label:"allowEmpty"`
	Zipkin        *zipkin.Config   `description:"Settings for zipkin." label:"allowEmpty"`
	DataDog       *datadog.Config  `description:"Settings for DataDog." label:"allowEmpty"`
	Instana       *instana.Config  `description:"Settings for Instana." label:"allowEmpty"`
	Haystack      *haystack.Config `description:"Settings for Haystack." label:"allowEmpty"`
}

// SetDefaults sets the default values.
func (t *Tracing) SetDefaults() {
	t.ServiceName = "traefik"
	t.SpanNameLimit = 0
}

// Providers contains providers configuration
type Providers struct {
	ProvidersThrottleDuration types.Duration     `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." export:"true"`
	Docker                    *docker.Provider   `description:"Enable Docker backend with default settings." export:"true" label:"allowEmpty"`
	File                      *file.Provider     `description:"Enable File backend with default settings." export:"true" label:"allowEmpty"`
	Marathon                  *marathon.Provider `description:"Enable Marathon backend with default settings." export:"true" label:"allowEmpty"`
	Kubernetes                *ingress.Provider  `description:"Enable Kubernetes backend with default settings." export:"true" label:"allowEmpty"`
	KubernetesCRD             *crd.Provider      `description:"Enable Kubernetes backend with default settings." export:"true" label:"allowEmpty"`
	Rest                      *rest.Provider     `description:"Enable Rest backend with default settings." export:"true" label:"allowEmpty"`
	Rancher                   *rancher.Provider  `description:"Enable Rancher backend with default settings." export:"true" label:"allowEmpty"`
}

// SetEffectiveConfiguration adds missing configuration parameters derived from existing ones.
// It also takes care of maintaining backwards compatibility.
func (c *Configuration) SetEffectiveConfiguration(configFile string) {
	if len(c.EntryPoints) == 0 {
		ep := &EntryPoint{Address: ":80"}
		ep.SetDefaults()
		c.EntryPoints = EntryPoints{
			"http": ep,
		}
	}

	if (c.API != nil && c.API.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Ping != nil && c.Ping.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Metrics != nil && c.Metrics.Prometheus != nil && c.Metrics.Prometheus.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Providers.Rest != nil && c.Providers.Rest.EntryPoint == DefaultInternalEntryPointName) {
		if _, ok := c.EntryPoints[DefaultInternalEntryPointName]; !ok {
			ep := &EntryPoint{Address: ":8080"}
			ep.SetDefaults()
			c.EntryPoints[DefaultInternalEntryPointName] = ep
		}
	}

	if c.Providers.Docker != nil {
		if c.Providers.Docker.SwarmModeRefreshSeconds <= 0 {
			c.Providers.Docker.SwarmModeRefreshSeconds = types.Duration(15 * time.Second)
		}
	}

	if c.Providers.File != nil {
		c.Providers.File.TraefikFile = configFile
	}

	if c.Providers.Rancher != nil {
		if c.Providers.Rancher.RefreshSeconds <= 0 {
			c.Providers.Rancher.RefreshSeconds = 15
		}
	}

	c.initACMEProvider()
}

// FIXME handle on new configuration ACME struct
func (c *Configuration) initACMEProvider() {
	if c.ACME != nil {
		c.ACME.CAServer = getSafeACMECAServer(c.ACME.CAServer)

		if c.ACME.DNSChallenge != nil && c.ACME.HTTPChallenge != nil {
			log.Warn("Unable to use DNS challenge and HTTP challenge at the same time. Fallback to DNS challenge.")
			c.ACME.HTTPChallenge = nil
		}

		if c.ACME.DNSChallenge != nil && c.ACME.TLSChallenge != nil {
			log.Warn("Unable to use DNS challenge and TLS challenge at the same time. Fallback to DNS challenge.")
			c.ACME.TLSChallenge = nil
		}

		if c.ACME.HTTPChallenge != nil && c.ACME.TLSChallenge != nil {
			log.Warn("Unable to use HTTP challenge and TLS challenge at the same time. Fallback to TLS challenge.")
			c.ACME.HTTPChallenge = nil
		}
	}
}

// InitACMEProvider create an acme provider from the ACME part of globalConfiguration
func (c *Configuration) InitACMEProvider() (*acmeprovider.Provider, error) {
	if c.ACME != nil {
		if len(c.ACME.Storage) == 0 {
			return nil, errors.New("unable to initialize ACME provider with no storage location for the certificates")
		}
		return &acmeprovider.Provider{
			Configuration: c.ACME,
		}, nil
	}
	return nil, nil
}

// ValidateConfiguration validate that configuration is coherent
func (c *Configuration) ValidateConfiguration() {
	if c.ACME != nil {
		for _, domain := range c.ACME.Domains {
			if domain.Main != dns01.UnFqdn(domain.Main) {
				log.Warnf("FQDN detected, please remove the trailing dot: %s", domain.Main)
			}
			for _, san := range domain.SANs {
				if san != dns01.UnFqdn(san) {
					log.Warnf("FQDN detected, please remove the trailing dot: %s", san)
				}
			}
		}
	}
	// FIXME Validate store config?
	// if c.ACME != nil {
	// if _, ok := c.EntryPoints[c.ACME.EntryPoint]; !ok {
	// 	log.Fatalf("Unknown entrypoint %q for ACME configuration", c.ACME.EntryPoint)
	// }
	// else if c.EntryPoints[c.ACME.EntryPoint].TLS == nil {
	// 	log.Fatalf("Entrypoint %q has no TLS configuration for ACME configuration", c.ACME.EntryPoint)
	// }
	// }
}

func getSafeACMECAServer(caServerSrc string) string {
	if len(caServerSrc) == 0 {
		return DefaultAcmeCAServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-v01.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "v01", "v02", 1)
		log.Warnf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-staging.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "https://acme-staging.api.letsencrypt.org", "https://acme-staging-v02.api.letsencrypt.org", 1)
		log.Warnf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	return caServerSrc
}
