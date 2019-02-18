package cmd

import (
	"time"

	"github.com/containous/flaeg"
	servicefabric "github.com/containous/traefik-extra-service-fabric"
	"github.com/containous/traefik/api"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/middlewares/tracing/datadog"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	"github.com/containous/traefik/ping"
	"github.com/containous/traefik/provider/boltdb"
	"github.com/containous/traefik/provider/consul"
	"github.com/containous/traefik/provider/consulcatalog"
	"github.com/containous/traefik/provider/docker"
	"github.com/containous/traefik/provider/dynamodb"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/etcd"
	"github.com/containous/traefik/provider/eureka"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/provider/marathon"
	"github.com/containous/traefik/provider/mesos"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/provider/rest"
	"github.com/containous/traefik/provider/zk"
	"github.com/containous/traefik/types"
	sf "github.com/jjcollinge/servicefabric"
)

// TraefikConfiguration holds GlobalConfiguration and other stuff
type TraefikConfiguration struct {
	configuration.GlobalConfiguration `mapstructure:",squash" export:"true"`
	ConfigFile                        string `short:"c" description:"Configuration file to use (TOML)." export:"true"`
}

// NewTraefikDefaultPointersConfiguration creates a TraefikConfiguration with pointers default values
func NewTraefikDefaultPointersConfiguration() *TraefikConfiguration {
	// default Docker
	var defaultDocker docker.Provider
	defaultDocker.Watch = true
	defaultDocker.ExposedByDefault = true
	defaultDocker.Endpoint = "unix:///var/run/docker.sock"
	defaultDocker.SwarmMode = false
	defaultDocker.SwarmModeRefreshSeconds = 15

	// default File
	var defaultFile file.Provider
	defaultFile.Watch = true
	defaultFile.Filename = "" // needs equivalent to  viper.ConfigFileUsed()

	// default Rest
	var defaultRest rest.Provider
	defaultRest.EntryPoint = configuration.DefaultInternalEntryPointName

	// TODO: Deprecated - Web provider, use REST provider instead
	var defaultWeb configuration.WebCompatibility
	defaultWeb.Address = ":8080"
	defaultWeb.Statistics = &types.Statistics{
		RecentErrors: 10,
	}

	// TODO: Deprecated - default Metrics
	defaultWeb.Metrics = &types.Metrics{
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

	// default Marathon
	var defaultMarathon marathon.Provider
	defaultMarathon.Watch = true
	defaultMarathon.Endpoint = "http://127.0.0.1:8080"
	defaultMarathon.ExposedByDefault = true
	defaultMarathon.Constraints = types.Constraints{}
	defaultMarathon.DialerTimeout = flaeg.Duration(5 * time.Second)
	defaultMarathon.ResponseHeaderTimeout = flaeg.Duration(60 * time.Second)
	defaultMarathon.TLSHandshakeTimeout = flaeg.Duration(5 * time.Second)
	defaultMarathon.KeepAlive = flaeg.Duration(10 * time.Second)

	// default Consul
	var defaultConsul consul.Provider
	defaultConsul.Watch = true
	defaultConsul.Endpoint = "127.0.0.1:8500"
	defaultConsul.Prefix = "traefik"
	defaultConsul.Constraints = types.Constraints{}

	// default CatalogProvider
	var defaultConsulCatalog consulcatalog.Provider
	defaultConsulCatalog.Endpoint = "127.0.0.1:8500"
	defaultConsulCatalog.ExposedByDefault = true
	defaultConsulCatalog.Constraints = types.Constraints{}
	defaultConsulCatalog.Prefix = "traefik"
	defaultConsulCatalog.FrontEndRule = "Host:{{.ServiceName}}.{{.Domain}}"
	defaultConsulCatalog.Stale = false

	// default Etcd
	var defaultEtcd etcd.Provider
	defaultEtcd.Watch = true
	defaultEtcd.Endpoint = "127.0.0.1:2379"
	defaultEtcd.Prefix = "/traefik"
	defaultEtcd.Constraints = types.Constraints{}

	// default Zookeeper
	var defaultZookeeper zk.Provider
	defaultZookeeper.Watch = true
	defaultZookeeper.Endpoint = "127.0.0.1:2181"
	defaultZookeeper.Prefix = "traefik"
	defaultZookeeper.Constraints = types.Constraints{}

	// default Boltdb
	var defaultBoltDb boltdb.Provider
	defaultBoltDb.Watch = true
	defaultBoltDb.Endpoint = "127.0.0.1:4001"
	defaultBoltDb.Prefix = "/traefik"
	defaultBoltDb.Constraints = types.Constraints{}

	// default Kubernetes
	var defaultKubernetes kubernetes.Provider
	defaultKubernetes.Watch = true
	defaultKubernetes.Constraints = types.Constraints{}

	// default Mesos
	var defaultMesos mesos.Provider
	defaultMesos.Watch = true
	defaultMesos.Endpoint = "http://127.0.0.1:5050"
	defaultMesos.ExposedByDefault = true
	defaultMesos.Constraints = types.Constraints{}
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
	defaultECS.Constraints = types.Constraints{}

	// default Rancher
	var defaultRancher rancher.Provider
	defaultRancher.Watch = true
	defaultRancher.ExposedByDefault = true
	defaultRancher.RefreshSeconds = 15

	// default DynamoDB
	var defaultDynamoDB dynamodb.Provider
	defaultDynamoDB.Constraints = types.Constraints{}
	defaultDynamoDB.RefreshSeconds = 15
	defaultDynamoDB.TableName = "traefik"
	defaultDynamoDB.Watch = true

	// default Eureka
	var defaultEureka eureka.Provider
	defaultEureka.RefreshSeconds = flaeg.Duration(30 * time.Second)

	// default ServiceFabric
	var defaultServiceFabric servicefabric.Provider
	defaultServiceFabric.APIVersion = sf.DefaultAPIVersion
	defaultServiceFabric.RefreshSeconds = 10

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

	// default HealthCheckConfig
	healthCheck := configuration.HealthCheckConfig{
		Interval: flaeg.Duration(configuration.DefaultHealthCheckInterval),
	}

	// default RespondingTimeouts
	respondingTimeouts := configuration.RespondingTimeouts{
		IdleTimeout: flaeg.Duration(configuration.DefaultIdleTimeout),
	}

	// default ForwardingTimeouts
	forwardingTimeouts := configuration.ForwardingTimeouts{
		DialTimeout: flaeg.Duration(configuration.DefaultDialTimeout),
	}

	// default Tracing
	defaultTracing := tracing.Tracing{
		Backend:       "jaeger",
		ServiceName:   "traefik",
		SpanNameLimit: 0,
		Jaeger: &jaeger.Config{
			SamplingServerURL:      "http://localhost:5778/sampling",
			SamplingType:           "const",
			SamplingParam:          1.0,
			LocalAgentHostPort:     "127.0.0.1:6831",
			TraceContextHeaderName: "uber-trace-id",
		},
		Zipkin: &zipkin.Config{
			HTTPEndpoint: "http://localhost:9411/api/v1/spans",
			SameSpan:     false,
			ID128Bit:     true,
			Debug:        false,
		},
		DataDog: &datadog.Config{
			LocalAgentHostPort: "localhost:8126",
			GlobalTag:          "",
			Debug:              false,
			PrioritySampling:   false,
		},
	}

	// default LifeCycle
	defaultLifeCycle := configuration.LifeCycle{
		GraceTimeOut: flaeg.Duration(configuration.DefaultGraceTimeout),
	}

	// default ApiConfiguration
	defaultAPI := api.Handler{
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

	defaultResolver := configuration.HostResolverConfig{
		CnameFlattening: false,
		ResolvConfig:    "/etc/resolv.conf",
		ResolvDepth:     5,
	}

	defaultConfiguration := configuration.GlobalConfiguration{
		Docker:             &defaultDocker,
		File:               &defaultFile,
		Web:                &defaultWeb,
		Rest:               &defaultRest,
		Marathon:           &defaultMarathon,
		Consul:             &defaultConsul,
		ConsulCatalog:      &defaultConsulCatalog,
		Etcd:               &defaultEtcd,
		Zookeeper:          &defaultZookeeper,
		Boltdb:             &defaultBoltDb,
		Kubernetes:         &defaultKubernetes,
		Mesos:              &defaultMesos,
		ECS:                &defaultECS,
		Rancher:            &defaultRancher,
		Eureka:             &defaultEureka,
		DynamoDB:           &defaultDynamoDB,
		Retry:              &configuration.Retry{},
		HealthCheck:        &healthCheck,
		RespondingTimeouts: &respondingTimeouts,
		ForwardingTimeouts: &forwardingTimeouts,
		TraefikLog:         &defaultTraefikLog,
		AccessLog:          &defaultAccessLog,
		LifeCycle:          &defaultLifeCycle,
		Ping:               &defaultPing,
		API:                &defaultAPI,
		Metrics:            &defaultMetrics,
		Tracing:            &defaultTracing,
		HostResolver:       &defaultResolver,
	}

	return &TraefikConfiguration{
		GlobalConfiguration: defaultConfiguration,
	}
}

// NewTraefikConfiguration creates a TraefikConfiguration with default values
func NewTraefikConfiguration() *TraefikConfiguration {
	return &TraefikConfiguration{
		GlobalConfiguration: configuration.GlobalConfiguration{
			AccessLogsFile:            "",
			TraefikLogsFile:           "",
			EntryPoints:               map[string]*configuration.EntryPoint{},
			Constraints:               types.Constraints{},
			DefaultEntryPoints:        []string{"http"},
			ProvidersThrottleDuration: flaeg.Duration(2 * time.Second),
			MaxIdleConnsPerHost:       200,
			IdleTimeout:               flaeg.Duration(0),
			HealthCheck: &configuration.HealthCheckConfig{
				Interval: flaeg.Duration(configuration.DefaultHealthCheckInterval),
			},
			LifeCycle: &configuration.LifeCycle{
				GraceTimeOut: flaeg.Duration(configuration.DefaultGraceTimeout),
			},
			CheckNewVersion: true,
		},
		ConfigFile: "",
	}
}
