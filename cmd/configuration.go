package cmd

import (
	"time"

	"github.com/containous/traefik/pkg/provider/rancher"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/ping"
	"github.com/containous/traefik/pkg/provider/docker"
	"github.com/containous/traefik/pkg/provider/file"
	"github.com/containous/traefik/pkg/provider/kubernetes/ingress"
	"github.com/containous/traefik/pkg/provider/marathon"
	"github.com/containous/traefik/pkg/provider/rest"
	"github.com/containous/traefik/pkg/tracing/datadog"
	"github.com/containous/traefik/pkg/tracing/instana"
	"github.com/containous/traefik/pkg/tracing/jaeger"
	"github.com/containous/traefik/pkg/tracing/zipkin"
	"github.com/containous/traefik/pkg/types"
	jaegercli "github.com/uber/jaeger-client-go"
)

// TraefikConfiguration holds GlobalConfiguration and other stuff
type TraefikConfiguration struct {
	static.Configuration `mapstructure:",squash" export:"true"`
	ConfigFile           string `short:"c" description:"Configuration file to use (TOML)." export:"true"`
}

// NewTraefikConfiguration creates a TraefikConfiguration with default values
func NewTraefikConfiguration() *TraefikConfiguration {
	return &TraefikConfiguration{
		Configuration: static.Configuration{
			Global: &static.Global{
				CheckNewVersion: true,
			},
			EntryPoints: make(static.EntryPoints),
			Providers: &static.Providers{
				ProvidersThrottleDuration: parse.Duration(2 * time.Second),
			},
			ServersTransport: &static.ServersTransport{
				MaxIdleConnsPerHost: 200,
			},
		},
		ConfigFile: "",
	}
}

// NewTraefikDefaultPointersConfiguration creates a TraefikConfiguration with pointers default values
func NewTraefikDefaultPointersConfiguration() *TraefikConfiguration {
	// default File
	var defaultFile file.Provider
	defaultFile.Watch = true
	defaultFile.Filename = "" // needs equivalent to  viper.ConfigFileUsed()

	// default Ping
	var defaultPing = ping.Handler{
		EntryPoint: "traefik",
	}

	// default TraefikLog
	defaultTraefikLog := types.TraefikLog{
		Format:   "common",
		FilePath: "",
	}

	// default AccessLog
	defaultAccessLog := types.AccessLog{
		Format:   accesslog.CommonFormat,
		FilePath: "",
		Filters:  &types.AccessLogFilters{},
		Fields: &types.AccessLogFields{
			DefaultMode: types.AccessLogKeep,
			Headers: &types.FieldHeaders{
				DefaultMode: types.AccessLogKeep,
			},
		},
	}

	// default Tracing
	defaultTracing := static.Tracing{
		Backend:       "jaeger",
		ServiceName:   "traefik",
		SpanNameLimit: 0,
		Jaeger: &jaeger.Config{
			SamplingServerURL:      "http://localhost:5778/sampling",
			SamplingType:           "const",
			SamplingParam:          1.0,
			LocalAgentHostPort:     "127.0.0.1:6831",
			Propagation:            "jaeger",
			Gen128Bit:              false,
			TraceContextHeaderName: jaegercli.TraceContextHeaderName,
		},
		Zipkin: &zipkin.Config{
			HTTPEndpoint: "http://localhost:9411/api/v1/spans",
			SameSpan:     false,
			ID128Bit:     true,
			Debug:        false,
			SampleRate:   1.0,
		},
		DataDog: &datadog.Config{
			LocalAgentHostPort: "localhost:8126",
			GlobalTag:          "",
			Debug:              false,
			PrioritySampling:   false,
		},
		Instana: &instana.Config{
			LocalAgentHost: "localhost",
			LocalAgentPort: 42699,
			LogLevel:       "info",
		},
	}

	// default ApiConfiguration
	defaultAPI := static.API{
		EntryPoint: "traefik",
		Dashboard:  true,
	}
	defaultAPI.Statistics = &types.Statistics{
		RecentErrors: 10,
	}

	// default Metrics
	defaultMetrics := types.Metrics{
		Prometheus: &types.Prometheus{
			Buckets:    types.Buckets{0.1, 0.3, 1.2, 5},
			EntryPoint: static.DefaultInternalEntryPointName,
		},
		Datadog: &types.Datadog{
			Address:      "localhost:8125",
			PushInterval: "10s",
		},
		StatsD: &types.Statsd{
			Address:      "localhost:8125",
			PushInterval: "10s",
		},
		InfluxDB: &types.InfluxDB{
			Address:      "localhost:8089",
			Protocol:     "udp",
			PushInterval: "10s",
		},
	}

	defaultResolver := types.HostResolverConfig{
		CnameFlattening: false,
		ResolvConfig:    "/etc/resolv.conf",
		ResolvDepth:     5,
	}

	var defaultDocker docker.Provider
	defaultDocker.Watch = true
	defaultDocker.ExposedByDefault = true
	defaultDocker.Endpoint = "unix:///var/run/docker.sock"
	defaultDocker.SwarmMode = false
	defaultDocker.SwarmModeRefreshSeconds = 15
	defaultDocker.DefaultRule = docker.DefaultTemplateRule

	// default Rest
	var defaultRest rest.Provider
	defaultRest.EntryPoint = static.DefaultInternalEntryPointName

	// default Marathon
	var defaultMarathon marathon.Provider
	defaultMarathon.Watch = true
	defaultMarathon.Endpoint = "http://127.0.0.1:8080"
	defaultMarathon.ExposedByDefault = true
	defaultMarathon.DialerTimeout = parse.Duration(5 * time.Second)
	defaultMarathon.ResponseHeaderTimeout = parse.Duration(60 * time.Second)
	defaultMarathon.TLSHandshakeTimeout = parse.Duration(5 * time.Second)
	defaultMarathon.KeepAlive = parse.Duration(10 * time.Second)
	defaultMarathon.DefaultRule = marathon.DefaultTemplateRule

	// default Kubernetes
	var defaultKubernetes ingress.Provider
	defaultKubernetes.Watch = true

	// default Rancher
	var defaultRancher rancher.Provider
	defaultRancher.Watch = true
	defaultRancher.ExposedByDefault = true
	defaultRancher.EnableServiceHealthFilter = true
	defaultRancher.RefreshSeconds = 15
	defaultRancher.DefaultRule = rancher.DefaultTemplateRule

	defaultProviders := static.Providers{
		File:       &defaultFile,
		Docker:     &defaultDocker,
		Rest:       &defaultRest,
		Marathon:   &defaultMarathon,
		Kubernetes: &defaultKubernetes,
	}

	return &TraefikConfiguration{
		Configuration: static.Configuration{
			Providers:    &defaultProviders,
			Log:          &defaultTraefikLog,
			AccessLog:    &defaultAccessLog,
			Ping:         &defaultPing,
			API:          &defaultAPI,
			Metrics:      &defaultMetrics,
			Tracing:      &defaultTracing,
			HostResolver: &defaultResolver,
		},
	}
}
