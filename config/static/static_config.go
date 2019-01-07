package static

import (
	"errors"
	"strings"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/old/provider/boltdb"
	"github.com/containous/traefik/old/provider/consul"
	"github.com/containous/traefik/old/provider/consulcatalog"
	"github.com/containous/traefik/old/provider/docker"
	"github.com/containous/traefik/old/provider/dynamodb"
	"github.com/containous/traefik/old/provider/ecs"
	"github.com/containous/traefik/old/provider/etcd"
	"github.com/containous/traefik/old/provider/eureka"
	"github.com/containous/traefik/old/provider/kubernetes"
	"github.com/containous/traefik/old/provider/marathon"
	"github.com/containous/traefik/old/provider/mesos"
	"github.com/containous/traefik/old/provider/rancher"
	"github.com/containous/traefik/old/provider/zk"
	"github.com/containous/traefik/ping"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/rest"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	"github.com/containous/traefik/types"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/xenolf/lego/challenge/dns01"
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

	ServersTransport *ServersTransport `description:"Servers default transport" export:"true"`
	EntryPoints      EntryPoints       `description:"Entrypoints definition using format: --entryPoints='Name:http Address::8000 Redirect.EntryPoint:https' --entryPoints='Name:https Address::4442 TLS:tests/traefik.crt,tests/traefik.key;prod/traefik.crt,prod/traefik.key'" export:"true"`
	Providers        *Providers        `description:"Providers configuration" export:"true"`

	API     *API           `description:"Enable api/dashboard" export:"true"`
	Metrics *types.Metrics `description:"Enable a metrics exporter" export:"true"`
	Ping    *ping.Handler  `description:"Enable ping" export:"true"`
	// Rest    *rest.Provider `description:"Enable Rest backend with default settings" export:"true"`

	Log       *types.TraefikLog
	AccessLog *types.AccessLog `description:"Access log settings" export:"true"`
	Tracing   *Tracing         `description:"OpenTracing configuration" export:"true"`

	HostResolver *HostResolverConfig `description:"Enable CNAME Flattening" export:"true"`

	ACME *acme.ACME `description:"Enable ACME (Let's Encrypt): automatic SSL" export:"true"`
}

// Global holds the global configuration.
type Global struct {
	Debug              bool `short:"d" description:"Enable debug mode" export:"true"`
	CheckNewVersion    bool `description:"Periodically check if a new version has been released" export:"true"`
	SendAnonymousUsage bool `description:"send periodically anonymous usage statistics" export:"true"`
}

// ServersTransport options to configure communication between Traefik and the servers
type ServersTransport struct {
	InsecureSkipVerify  bool                `description:"Disable SSL certificate verification" export:"true"`
	RootCAs             tls.FilesOrContents `description:"Add cert file for self-signed certificate"`
	MaxIdleConnsPerHost int                 `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts `description:"Timeouts for requests forwarded to the backend servers" export:"true"`
}

// API holds the API configuration
type API struct {
	EntryPoint      string            `description:"EntryPoint" export:"true"`
	Dashboard       bool              `description:"Activate dashboard" export:"true"`
	Statistics      *types.Statistics `description:"Enable more detailed statistics" export:"true"`
	Middlewares     []string          `description:"Middleware list" export:"true"`
	DashboardAssets *assetfs.AssetFS  `json:"-"`
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  parse.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set" export:"true"`
	WriteTimeout parse.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set" export:"true"`
	IdleTimeout  parse.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. Defaults to 180 seconds. If zero, no timeout is set" export:"true"`
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           parse.Duration `description:"The amount of time to wait until a connection to a backend server can be established. Defaults to 30 seconds. If zero, no timeout exists" export:"true"`
	ResponseHeaderTimeout parse.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists" export:"true"`
}

// LifeCycle contains configurations relevant to the lifecycle (such as the shutdown phase) of Traefik.
type LifeCycle struct {
	RequestAcceptGraceTimeout parse.Duration `description:"Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure"`
	GraceTimeOut              parse.Duration `description:"Duration to give active requests a chance to finish before Traefik stops"`
}

// Tracing holds the tracing configuration.
type Tracing struct {
	Backend       string          `description:"Selects the tracking backend ('jaeger','zipkin', 'datadog')." export:"true"`
	ServiceName   string          `description:"Set the name for this service" export:"true"`
	SpanNameLimit int             `description:"Set the maximum character limit for Span names (default 0 = no limit)" export:"true"`
	Jaeger        *jaeger.Config  `description:"Settings for jaeger"`
	Zipkin        *zipkin.Config  `description:"Settings for zipkin"`
	DataDog       *datadog.Config `description:"Settings for DataDog"`
}

// HostResolverConfig contain configuration for CNAME Flattening.
type HostResolverConfig struct {
	CnameFlattening bool   `description:"A flag to enable/disable CNAME flattening" export:"true"`
	ResolvConfig    string `description:"resolv.conf used for DNS resolving" export:"true"`
	ResolvDepth     int    `description:"The maximal depth of DNS recursive resolving" export:"true"`
}

// Providers contains providers configuration
type Providers struct {
	ProvidersThrottleDuration parse.Duration          `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." export:"true"`
	Docker                    *docker.Provider        `description:"Enable Docker backend with default settings" export:"true"`
	File                      *file.Provider          `description:"Enable File backend with default settings" export:"true"`
	Marathon                  *marathon.Provider      `description:"Enable Marathon backend with default settings" export:"true"`
	Consul                    *consul.Provider        `description:"Enable Consul backend with default settings" export:"true"`
	ConsulCatalog             *consulcatalog.Provider `description:"Enable Consul catalog backend with default settings" export:"true"`
	Etcd                      *etcd.Provider          `description:"Enable Etcd backend with default settings" export:"true"`
	Zookeeper                 *zk.Provider            `description:"Enable Zookeeper backend with default settings" export:"true"`
	Boltdb                    *boltdb.Provider        `description:"Enable Boltdb backend with default settings" export:"true"`
	Kubernetes                *kubernetes.Provider    `description:"Enable Kubernetes backend with default settings" export:"true"`
	Mesos                     *mesos.Provider         `description:"Enable Mesos backend with default settings" export:"true"`
	Eureka                    *eureka.Provider        `description:"Enable Eureka backend with default settings" export:"true"`
	ECS                       *ecs.Provider           `description:"Enable ECS backend with default settings" export:"true"`
	Rancher                   *rancher.Provider       `description:"Enable Rancher backend with default settings" export:"true"`
	DynamoDB                  *dynamodb.Provider      `description:"Enable DynamoDB backend with default settings" export:"true"`
	Rest                      *rest.Provider          `description:"Enable Rest backend with default settings" export:"true"`
}

// SetEffectiveConfiguration adds missing configuration parameters derived from existing ones.
// It also takes care of maintaining backwards compatibility.
func (c *Configuration) SetEffectiveConfiguration(configFile string) {
	if len(c.EntryPoints) == 0 {
		c.EntryPoints = EntryPoints{
			"http": &EntryPoint{
				Address: ":80",
			},
		}
	}

	if (c.API != nil && c.API.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Ping != nil && c.Ping.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Metrics != nil && c.Metrics.Prometheus != nil && c.Metrics.Prometheus.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Providers.Rest != nil && c.Providers.Rest.EntryPoint == DefaultInternalEntryPointName) {
		if _, ok := c.EntryPoints[DefaultInternalEntryPointName]; !ok {
			c.EntryPoints[DefaultInternalEntryPointName] = &EntryPoint{Address: ":8080"}
		}
	}

	for _, entryPoint := range c.EntryPoints {
		if entryPoint.Transport == nil {
			entryPoint.Transport = &EntryPointsTransport{}
		}

		// Make sure LifeCycle isn't nil to spare nil checks elsewhere.
		if entryPoint.Transport.LifeCycle == nil {
			entryPoint.Transport.LifeCycle = &LifeCycle{
				GraceTimeOut: parse.Duration(DefaultGraceTimeout),
			}
			entryPoint.Transport.RespondingTimeouts = &RespondingTimeouts{
				IdleTimeout: parse.Duration(DefaultIdleTimeout),
			}

		}
	}

	if c.Providers.Rancher != nil {
		// Ensure backwards compatibility for now
		if len(c.Providers.Rancher.AccessKey) > 0 ||
			len(c.Providers.Rancher.Endpoint) > 0 ||
			len(c.Providers.Rancher.SecretKey) > 0 {

			if c.Providers.Rancher.API == nil {
				c.Providers.Rancher.API = &rancher.APIConfiguration{
					AccessKey: c.Providers.Rancher.AccessKey,
					SecretKey: c.Providers.Rancher.SecretKey,
					Endpoint:  c.Providers.Rancher.Endpoint,
				}
			}
			log.Warn("Deprecated configuration found: rancher.[accesskey|secretkey|endpoint]. " +
				"Please use rancher.api.[accesskey|secretkey|endpoint] instead.")
		}

		if c.Providers.Rancher.Metadata != nil && len(c.Providers.Rancher.Metadata.Prefix) == 0 {
			c.Providers.Rancher.Metadata.Prefix = "latest"
		}
	}

	if c.Providers.File != nil {
		c.Providers.File.TraefikFile = configFile
	}

	c.initACMEProvider()
	c.initTracing()
}

func (c *Configuration) initTracing() {
	if c.Tracing != nil {
		switch c.Tracing.Backend {
		case jaeger.Name:
			if c.Tracing.Jaeger == nil {
				c.Tracing.Jaeger = &jaeger.Config{
					SamplingServerURL:  "http://localhost:5778/sampling",
					SamplingType:       "const",
					SamplingParam:      1.0,
					LocalAgentHostPort: "127.0.0.1:6831",
					Propagation:        "jaeger",
					Gen128Bit:          false,
				}
			}
			if c.Tracing.Zipkin != nil {
				log.Warn("Zipkin configuration will be ignored")
				c.Tracing.Zipkin = nil
			}
			if c.Tracing.DataDog != nil {
				log.Warn("DataDog configuration will be ignored")
				c.Tracing.DataDog = nil
			}
		case zipkin.Name:
			if c.Tracing.Zipkin == nil {
				c.Tracing.Zipkin = &zipkin.Config{
					HTTPEndpoint: "http://localhost:9411/api/v1/spans",
					SameSpan:     false,
					ID128Bit:     true,
					Debug:        false,
					SampleRate:   1.0,
				}
			}
			if c.Tracing.Jaeger != nil {
				log.Warn("Jaeger configuration will be ignored")
				c.Tracing.Jaeger = nil
			}
			if c.Tracing.DataDog != nil {
				log.Warn("DataDog configuration will be ignored")
				c.Tracing.DataDog = nil
			}
		case datadog.Name:
			if c.Tracing.DataDog == nil {
				c.Tracing.DataDog = &datadog.Config{
					LocalAgentHostPort: "localhost:8126",
					GlobalTag:          "",
					Debug:              false,
				}
			}
			if c.Tracing.Zipkin != nil {
				log.Warn("Zipkin configuration will be ignored")
				c.Tracing.Zipkin = nil
			}
			if c.Tracing.Jaeger != nil {
				log.Warn("Jaeger configuration will be ignored")
				c.Tracing.Jaeger = nil
			}
		default:
			log.Warnf("Unknown tracer %q", c.Tracing.Backend)
			return
		}
	}
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

		if c.ACME.OnDemand {
			log.Warn("ACME.OnDemand is deprecated")
		}
	}
}

// InitACMEProvider create an acme provider from the ACME part of globalConfiguration
func (c *Configuration) InitACMEProvider() (*acmeprovider.Provider, error) {
	if c.ACME != nil {
		if len(c.ACME.Storage) == 0 {
			// Delete the ACME configuration to avoid starting ACME in cluster mode
			c.ACME = nil
			return nil, errors.New("unable to initialize ACME provider with no storage location for the certificates")
		}
		provider := &acmeprovider.Provider{}
		provider.Configuration = convertACMEChallenge(c.ACME)

		store := acmeprovider.NewLocalStore(provider.Storage)
		provider.Store = store
		acme.ConvertToNewFormat(provider.Storage)
		c.ACME = nil
		return provider, nil
	}
	return nil, nil
}

// ValidateConfiguration validate that configuration is coherent
func (c *Configuration) ValidateConfiguration() {
	if c.ACME != nil {
		if _, ok := c.EntryPoints[c.ACME.EntryPoint]; !ok {
			log.Fatalf("Unknown entrypoint %q for ACME configuration", c.ACME.EntryPoint)
		} else {
			if c.EntryPoints[c.ACME.EntryPoint].TLS == nil {
				log.Fatalf("Entrypoint %q has no TLS configuration for ACME configuration", c.ACME.EntryPoint)
			}
		}
	}
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

// Deprecated
func convertACMEChallenge(oldACMEChallenge *acme.ACME) *acmeprovider.Configuration {
	conf := &acmeprovider.Configuration{
		KeyType:     oldACMEChallenge.KeyType,
		OnHostRule:  oldACMEChallenge.OnHostRule,
		OnDemand:    oldACMEChallenge.OnDemand,
		Email:       oldACMEChallenge.Email,
		Storage:     oldACMEChallenge.Storage,
		ACMELogging: oldACMEChallenge.ACMELogging,
		CAServer:    oldACMEChallenge.CAServer,
		EntryPoint:  oldACMEChallenge.EntryPoint,
	}

	for _, domain := range oldACMEChallenge.Domains {
		if domain.Main != dns01.UnFqdn(domain.Main) {
			log.Warnf("FQDN detected, please remove the trailing dot: %s", domain.Main)
		}
		for _, san := range domain.SANs {
			if san != dns01.UnFqdn(san) {
				log.Warnf("FQDN detected, please remove the trailing dot: %s", san)
			}
		}
		conf.Domains = append(conf.Domains, domain)
	}
	if oldACMEChallenge.HTTPChallenge != nil {
		conf.HTTPChallenge = &acmeprovider.HTTPChallenge{
			EntryPoint: oldACMEChallenge.HTTPChallenge.EntryPoint,
		}
	}

	if oldACMEChallenge.DNSChallenge != nil {
		conf.DNSChallenge = &acmeprovider.DNSChallenge{
			Provider:         oldACMEChallenge.DNSChallenge.Provider,
			DelayBeforeCheck: oldACMEChallenge.DNSChallenge.DelayBeforeCheck,
		}
	}

	if oldACMEChallenge.TLSChallenge != nil {
		conf.TLSChallenge = &acmeprovider.TLSChallenge{}
	}

	return conf
}
