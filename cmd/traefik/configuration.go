package main

import (
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/provider/boltdb"
	"github.com/containous/traefik/provider/consul"
	"github.com/containous/traefik/provider/docker"
	"github.com/containous/traefik/provider/dynamodb"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/etcd"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/provider/marathon"
	"github.com/containous/traefik/provider/mesos"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/provider/web"
	"github.com/containous/traefik/provider/zk"
	"github.com/containous/traefik/types"
)

// TraefikConfiguration holds GlobalConfiguration and other stuff
type TraefikConfiguration struct {
	configuration.GlobalConfiguration `mapstructure:",squash"`
	ConfigFile                        string `short:"c" description:"Configuration file to use (TOML)."`
}

// NewTraefikDefaultPointersConfiguration creates a TraefikConfiguration with pointers default values
func NewTraefikDefaultPointersConfiguration() *TraefikConfiguration {
	//default Docker
	var defaultDocker docker.Provider
	defaultDocker.Watch = true
	defaultDocker.ExposedByDefault = true
	defaultDocker.Endpoint = "unix:///var/run/docker.sock"
	defaultDocker.SwarmMode = false

	// default File
	var defaultFile file.Provider
	defaultFile.Watch = true
	defaultFile.Filename = "" //needs equivalent to  viper.ConfigFileUsed()

	// default Web
	var defaultWeb web.Provider
	defaultWeb.Address = ":8080"
	defaultWeb.Statistics = &types.Statistics{
		RecentErrors: 10,
	}

	// default Metrics
	defaultWeb.Metrics = &types.Metrics{
		Prometheus: &types.Prometheus{
			Buckets: types.Buckets{0.1, 0.3, 1.2, 5},
		},
		Datadog: &types.Datadog{
			Address:      "localhost:8125",
			PushInterval: "10s",
		},
		StatsD: &types.Statsd{
			Address:      "localhost:8125",
			PushInterval: "10s",
		},
	}

	// default Marathon
	var defaultMarathon marathon.Provider
	defaultMarathon.Watch = true
	defaultMarathon.Endpoint = "http://127.0.0.1:8080"
	defaultMarathon.ExposedByDefault = true
	defaultMarathon.Constraints = types.Constraints{}
	defaultMarathon.DialerTimeout = flaeg.Duration(60 * time.Second)
	defaultMarathon.KeepAlive = flaeg.Duration(10 * time.Second)

	// default Consul
	var defaultConsul consul.Provider
	defaultConsul.Watch = true
	defaultConsul.Endpoint = "127.0.0.1:8500"
	defaultConsul.Prefix = "traefik"
	defaultConsul.Constraints = types.Constraints{}

	// default CatalogProvider
	var defaultConsulCatalog consul.CatalogProvider
	defaultConsulCatalog.Endpoint = "127.0.0.1:8500"
	defaultConsulCatalog.ExposedByDefault = true
	defaultConsulCatalog.Constraints = types.Constraints{}
	defaultConsulCatalog.Prefix = "traefik"
	defaultConsulCatalog.FrontEndRule = "Host:{{.ServiceName}}.{{.Domain}}"

	// default Etcd
	var defaultEtcd etcd.Provider
	defaultEtcd.Watch = true
	defaultEtcd.Endpoint = "127.0.0.1:2379"
	defaultEtcd.Prefix = "/traefik"
	defaultEtcd.Constraints = types.Constraints{}

	//default Zookeeper
	var defaultZookeeper zk.Provider
	defaultZookeeper.Watch = true
	defaultZookeeper.Endpoint = "127.0.0.1:2181"
	defaultZookeeper.Prefix = "/traefik"
	defaultZookeeper.Constraints = types.Constraints{}

	//default Boltdb
	var defaultBoltDb boltdb.Provider
	defaultBoltDb.Watch = true
	defaultBoltDb.Endpoint = "127.0.0.1:4001"
	defaultBoltDb.Prefix = "/traefik"
	defaultBoltDb.Constraints = types.Constraints{}

	//default Kubernetes
	var defaultKubernetes kubernetes.Provider
	defaultKubernetes.Watch = true
	defaultKubernetes.Endpoint = ""
	defaultKubernetes.LabelSelector = ""
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

	//default ECS
	var defaultECS ecs.Provider
	defaultECS.Watch = true
	defaultECS.ExposedByDefault = true
	defaultECS.AutoDiscoverClusters = false
	defaultECS.Clusters = ecs.Clusters{"default"}
	defaultECS.RefreshSeconds = 15
	defaultECS.Constraints = types.Constraints{}

	//default Rancher
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

	// default AccessLog
	defaultAccessLog := types.AccessLog{
		Format:   accesslog.CommonFormat,
		FilePath: "",
	}

	defaultConfiguration := configuration.GlobalConfiguration{
		Docker:        &defaultDocker,
		File:          &defaultFile,
		Web:           &defaultWeb,
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
		DynamoDB:      &defaultDynamoDB,
		Retry:         &configuration.Retry{},
		HealthCheck:   &configuration.HealthCheckConfig{},
		AccessLog:     &defaultAccessLog,
	}

	return &TraefikConfiguration{
		GlobalConfiguration: defaultConfiguration,
	}
}

// NewTraefikConfiguration creates a TraefikConfiguration with default values
func NewTraefikConfiguration() *TraefikConfiguration {
	return &TraefikConfiguration{
		GlobalConfiguration: configuration.GlobalConfiguration{
			GraceTimeOut:              flaeg.Duration(10 * time.Second),
			AccessLogsFile:            "",
			TraefikLogsFile:           "",
			LogLevel:                  "ERROR",
			EntryPoints:               map[string]*configuration.EntryPoint{},
			Constraints:               types.Constraints{},
			DefaultEntryPoints:        []string{},
			ProvidersThrottleDuration: flaeg.Duration(2 * time.Second),
			MaxIdleConnsPerHost:       200,
			IdleTimeout:               flaeg.Duration(0),
			HealthCheck: &configuration.HealthCheckConfig{
				Interval: flaeg.Duration(configuration.DefaultHealthCheckInterval),
			},
			RespondingTimeouts: &configuration.RespondingTimeouts{
				IdleTimeout: flaeg.Duration(configuration.DefaultIdleTimeout),
			},
			ForwardingTimeouts: &configuration.ForwardingTimeouts{
				DialTimeout: flaeg.Duration(configuration.DefaultDialTimeout),
			},
			CheckNewVersion: true,
		},
		ConfigFile: "",
	}
}
