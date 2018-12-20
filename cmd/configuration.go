package cmd

import (
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/old/configuration"
	"github.com/containous/traefik/old/middlewares/accesslog"
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
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/rest"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	"github.com/containous/traefik/types"
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
				CheckNewVersion:    true,
				SendAnonymousUsage: false,
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
			SamplingServerURL:  "http://localhost:5778/sampling",
			SamplingType:       "const",
			SamplingParam:      1.0,
			LocalAgentHostPort: "127.0.0.1:6831",
			Propagation:        "jaeger",
			Gen128Bit:          false,
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
			EntryPoint: configuration.DefaultInternalEntryPointName,
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

	defaultResolver := static.HostResolverConfig{
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

	// default Rest
	var defaultRest rest.Provider
	defaultRest.EntryPoint = configuration.DefaultInternalEntryPointName

	// default Marathon
	var defaultMarathon marathon.Provider
	defaultMarathon.Watch = true
	defaultMarathon.Endpoint = "http://127.0.0.1:8080"
	defaultMarathon.ExposedByDefault = true
	defaultMarathon.DialerTimeout = parse.Duration(5 * time.Second)
	defaultMarathon.ResponseHeaderTimeout = parse.Duration(60 * time.Second)
	defaultMarathon.TLSHandshakeTimeout = parse.Duration(5 * time.Second)
	defaultMarathon.KeepAlive = parse.Duration(10 * time.Second)

	// default Consul
	var defaultConsul consul.Provider
	defaultConsul.Watch = true
	defaultConsul.Endpoint = "127.0.0.1:8500"
	defaultConsul.Prefix = "traefik"

	// default CatalogProvider
	var defaultConsulCatalog consulcatalog.Provider
	defaultConsulCatalog.Endpoint = "127.0.0.1:8500"
	defaultConsulCatalog.ExposedByDefault = true
	defaultConsulCatalog.Prefix = "traefik"
	defaultConsulCatalog.FrontEndRule = "Host:{{.ServiceName}}.{{.Domain}}"
	defaultConsulCatalog.Stale = false

	// default Etcd
	var defaultEtcd etcd.Provider
	defaultEtcd.Watch = true
	defaultEtcd.Endpoint = "127.0.0.1:2379"
	defaultEtcd.Prefix = "/traefik"

	// default Zookeeper
	var defaultZookeeper zk.Provider
	defaultZookeeper.Watch = true
	defaultZookeeper.Endpoint = "127.0.0.1:2181"
	defaultZookeeper.Prefix = "traefik"

	// default Boltdb
	var defaultBoltDb boltdb.Provider
	defaultBoltDb.Watch = true
	defaultBoltDb.Endpoint = "127.0.0.1:4001"
	defaultBoltDb.Prefix = "/traefik"

	// default Kubernetes
	var defaultKubernetes kubernetes.Provider
	defaultKubernetes.Watch = true

	// default Mesos
	var defaultMesos mesos.Provider
	defaultMesos.Watch = true
	defaultMesos.Endpoint = "http://127.0.0.1:5050"
	defaultMesos.ExposedByDefault = true
	defaultMesos.RefreshSeconds = 30
	defaultMesos.ZkDetectionTimeout = 30
	defaultMesos.StateTimeoutSecond = 30

	// default ECS
	var defaultECS ecs.Provider
	defaultECS.Watch = true
	defaultECS.ExposedByDefault = true
	defaultECS.AutoDiscoverClusters = false
	defaultECS.Clusters = ecs.Clusters{"default"}
	defaultECS.RefreshSeconds = 15

	// default Rancher
	var defaultRancher rancher.Provider
	defaultRancher.Watch = true
	defaultRancher.ExposedByDefault = true
	defaultRancher.RefreshSeconds = 15

	// default DynamoDB
	var defaultDynamoDB dynamodb.Provider
	defaultDynamoDB.RefreshSeconds = 15
	defaultDynamoDB.TableName = "traefik"
	defaultDynamoDB.Watch = true

	// default Eureka
	var defaultEureka eureka.Provider
	defaultEureka.RefreshSeconds = parse.Duration(30 * time.Second)

	defaultProviders := static.Providers{
		File:          &defaultFile,
		Docker:        &defaultDocker,
		Rest:          &defaultRest,
		Marathon:      &defaultMarathon,
		Consul:        &defaultConsul,
		ConsulCatalog: &defaultConsulCatalog,
		Etcd:          &defaultEtcd,
		Zookeeper:     &defaultZookeeper,
		Boltdb:        &defaultBoltDb,
		Kubernetes:    &defaultKubernetes,
		Mesos:         &defaultMesos,
		ECS:           &defaultECS,
		Rancher:       &defaultRancher,
		Eureka:        &defaultEureka,
		DynamoDB:      &defaultDynamoDB,
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
