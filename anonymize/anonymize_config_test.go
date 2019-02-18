package anonymize

import (
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/api"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/provider"
	acmeprovider "github.com/containous/traefik/provider/acme"
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
	"github.com/containous/traefik/provider/kv"
	"github.com/containous/traefik/provider/marathon"
	"github.com/containous/traefik/provider/mesos"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/provider/zk"
	"github.com/containous/traefik/safe"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/thoas/stats"
)

func TestDo_globalConfiguration(t *testing.T) {

	config := &configuration.GlobalConfiguration{}

	config.GraceTimeOut = flaeg.Duration(666 * time.Second)
	config.Debug = true
	config.CheckNewVersion = true
	config.AccessLogsFile = "AccessLogsFile"
	config.AccessLog = &types.AccessLog{
		FilePath: "AccessLog FilePath",
		Format:   "AccessLog Format",
	}
	config.TraefikLogsFile = "TraefikLogsFile"
	config.LogLevel = "LogLevel"
	config.EntryPoints = configuration.EntryPoints{
		"foo": {
			Address: "foo Address",
			TLS: &traefiktls.TLS{
				MinVersion:   "foo MinVersion",
				CipherSuites: []string{"foo CipherSuites 1", "foo CipherSuites 2", "foo CipherSuites 3"},
				Certificates: traefiktls.Certificates{
					{CertFile: "CertFile 1", KeyFile: "KeyFile 1"},
					{CertFile: "CertFile 2", KeyFile: "KeyFile 2"},
				},
				ClientCA: traefiktls.ClientCA{
					Files:    traefiktls.FilesOrContents{"foo ClientCAFiles 1", "foo ClientCAFiles 2", "foo ClientCAFiles 3"},
					Optional: false,
				},
			},
			Redirect: &types.Redirect{
				Replacement: "foo Replacement",
				Regex:       "foo Regex",
				EntryPoint:  "foo EntryPoint",
			},
			Auth: &types.Auth{
				Basic: &types.Basic{
					UsersFile: "foo Basic UsersFile",
					Users:     types.Users{"foo Basic Users 1", "foo Basic Users 2", "foo Basic Users 3"},
				},
				Digest: &types.Digest{
					UsersFile: "foo Digest UsersFile",
					Users:     types.Users{"foo Digest Users 1", "foo Digest Users 2", "foo Digest Users 3"},
				},
				Forward: &types.Forward{
					Address: "foo Address",
					TLS: &types.ClientTLS{
						CA:                 "foo CA",
						Cert:               "foo Cert",
						Key:                "foo Key",
						InsecureSkipVerify: true,
					},
					TrustForwardHeader: true,
				},
			},
			WhitelistSourceRange: []string{"foo WhitelistSourceRange 1", "foo WhitelistSourceRange 2", "foo WhitelistSourceRange 3"},
			Compress:             true,
			ProxyProtocol: &configuration.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
		},
		"fii": {
			Address: "fii Address",
			TLS: &traefiktls.TLS{
				MinVersion:   "fii MinVersion",
				CipherSuites: []string{"fii CipherSuites 1", "fii CipherSuites 2", "fii CipherSuites 3"},
				Certificates: traefiktls.Certificates{
					{CertFile: "CertFile 1", KeyFile: "KeyFile 1"},
					{CertFile: "CertFile 2", KeyFile: "KeyFile 2"},
				},
				ClientCA: traefiktls.ClientCA{
					Files:    traefiktls.FilesOrContents{"fii ClientCAFiles 1", "fii ClientCAFiles 2", "fii ClientCAFiles 3"},
					Optional: false,
				},
			},
			Redirect: &types.Redirect{
				Replacement: "fii Replacement",
				Regex:       "fii Regex",
				EntryPoint:  "fii EntryPoint",
			},
			Auth: &types.Auth{
				Basic: &types.Basic{
					UsersFile: "fii Basic UsersFile",
					Users:     types.Users{"fii Basic Users 1", "fii Basic Users 2", "fii Basic Users 3"},
				},
				Digest: &types.Digest{
					UsersFile: "fii Digest UsersFile",
					Users:     types.Users{"fii Digest Users 1", "fii Digest Users 2", "fii Digest Users 3"},
				},
				Forward: &types.Forward{
					Address: "fii Address",
					TLS: &types.ClientTLS{
						CA:                 "fii CA",
						Cert:               "fii Cert",
						Key:                "fii Key",
						InsecureSkipVerify: true,
					},
					TrustForwardHeader: true,
				},
			},
			WhitelistSourceRange: []string{"fii WhitelistSourceRange 1", "fii WhitelistSourceRange 2", "fii WhitelistSourceRange 3"},
			Compress:             true,
			ProxyProtocol: &configuration.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
		},
	}
	config.Cluster = &types.Cluster{
		Node: "Cluster Node",
		Store: &types.Store{
			Prefix: "Cluster Store Prefix",
			// ...
		},
	}
	config.Constraints = types.Constraints{
		{
			Key:       "Constraints Key 1",
			Regex:     "Constraints Regex 2",
			MustMatch: true,
		},
		{
			Key:       "Constraints Key 1",
			Regex:     "Constraints Regex 2",
			MustMatch: true,
		},
	}
	config.ACME = &acme.ACME{
		Email: "acme Email",
		Domains: []types.Domain{
			{
				Main: "Domains Main",
				SANs: []string{"Domains acme SANs 1", "Domains acme SANs 2", "Domains acme SANs 3"},
			},
		},
		Storage:           "Storage",
		StorageFile:       "StorageFile",
		OnDemand:          true,
		OnHostRule:        true,
		CAServer:          "CAServer",
		EntryPoint:        "EntryPoint",
		DNSChallenge:      &acmeprovider.DNSChallenge{Provider: "DNSProvider"},
		DelayDontCheckDNS: 666,
		ACMELogging:       true,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			// ...
		},
	}
	config.DefaultEntryPoints = configuration.DefaultEntryPoints{"DefaultEntryPoints 1", "DefaultEntryPoints 2", "DefaultEntryPoints 3"}
	config.ProvidersThrottleDuration = flaeg.Duration(666 * time.Second)
	config.MaxIdleConnsPerHost = 666
	config.IdleTimeout = flaeg.Duration(666 * time.Second)
	config.InsecureSkipVerify = true
	config.RootCAs = traefiktls.FilesOrContents{"RootCAs 1", "RootCAs 2", "RootCAs 3"}
	config.Retry = &configuration.Retry{
		Attempts: 666,
	}
	config.HealthCheck = &configuration.HealthCheckConfig{
		Interval: flaeg.Duration(666 * time.Second),
	}
	config.API = &api.Handler{
		EntryPoint:            "traefik",
		Dashboard:             true,
		Debug:                 true,
		CurrentConfigurations: &safe.Safe{},
		Statistics: &types.Statistics{
			RecentErrors: 666,
		},
		Stats: &stats.Stats{
			Uptime:              time.Now(),
			Pid:                 666,
			ResponseCounts:      map[string]int{"foo": 1},
			TotalResponseCounts: map[string]int{"bar": 1},
			TotalResponseTime:   time.Now(),
		},
		StatsRecorder: &middlewares.StatsRecorder{},
		DashboardAssets: &assetfs.AssetFS{
			Asset: func(path string) ([]byte, error) {
				return nil, nil
			},
			AssetDir: func(path string) ([]string, error) {
				return nil, nil
			},
			AssetInfo: func(path string) (os.FileInfo, error) {
				return nil, nil
			},
			Prefix: "fii",
		},
	}
	config.RespondingTimeouts = &configuration.RespondingTimeouts{
		ReadTimeout:  flaeg.Duration(666 * time.Second),
		WriteTimeout: flaeg.Duration(666 * time.Second),
		IdleTimeout:  flaeg.Duration(666 * time.Second),
	}
	config.ForwardingTimeouts = &configuration.ForwardingTimeouts{
		DialTimeout:           flaeg.Duration(666 * time.Second),
		ResponseHeaderTimeout: flaeg.Duration(666 * time.Second),
	}
	config.Docker = &docker.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "docker Filename",
			Constraints: types.Constraints{
				{
					Key:       "docker Constraints Key 1",
					Regex:     "docker Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "docker Constraints Key 1",
					Regex:     "docker Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint: "docker Endpoint",
		Domain:   "docker Domain",
		TLS: &types.ClientTLS{
			CA:                 "docker CA",
			Cert:               "docker Cert",
			Key:                "docker Key",
			InsecureSkipVerify: true,
		},
		ExposedByDefault: true,
		UseBindPortIP:    true,
		SwarmMode:        true,
	}
	config.File = &file.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "file Filename",
			Constraints: types.Constraints{
				{
					Key:       "file Constraints Key 1",
					Regex:     "file Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "file Constraints Key 1",
					Regex:     "file Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Directory: "file Directory",
	}
	config.Web = &configuration.WebCompatibility{
		Address:  "web Address",
		CertFile: "web CertFile",
		KeyFile:  "web KeyFile",
		ReadOnly: true,
		Statistics: &types.Statistics{
			RecentErrors: 666,
		},
		Metrics: &types.Metrics{
			Prometheus: &types.Prometheus{
				Buckets: types.Buckets{6.5, 6.6, 6.7},
			},
			Datadog: &types.Datadog{
				Address:      "Datadog Address",
				PushInterval: "Datadog PushInterval",
			},
			StatsD: &types.Statsd{
				Address:      "StatsD Address",
				PushInterval: "StatsD PushInterval",
			},
		},
		Path: "web Path",
		Auth: &types.Auth{
			Basic: &types.Basic{
				UsersFile: "web Basic UsersFile",
				Users:     types.Users{"web Basic Users 1", "web Basic Users 2", "web Basic Users 3"},
			},
			Digest: &types.Digest{
				UsersFile: "web Digest UsersFile",
				Users:     types.Users{"web Digest Users 1", "web Digest Users 2", "web Digest Users 3"},
			},
			Forward: &types.Forward{
				Address: "web Address",
				TLS: &types.ClientTLS{
					CA:                 "web CA",
					Cert:               "web Cert",
					Key:                "web Key",
					InsecureSkipVerify: true,
				},
				TrustForwardHeader: true,
			},
		},
		Debug: true,
	}
	config.Marathon = &marathon.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "marathon Filename",
			Constraints: types.Constraints{
				{
					Key:       "marathon Constraints Key 1",
					Regex:     "marathon Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "marathon Constraints Key 1",
					Regex:     "marathon Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:                "",
		Domain:                  "",
		ExposedByDefault:        true,
		GroupsAsSubDomains:      true,
		DCOSToken:               "",
		MarathonLBCompatibility: true,
		TLS: &types.ClientTLS{
			CA:                 "marathon CA",
			Cert:               "marathon Cert",
			Key:                "marathon Key",
			InsecureSkipVerify: true,
		},
		DialerTimeout:     flaeg.Duration(666 * time.Second),
		KeepAlive:         flaeg.Duration(666 * time.Second),
		ForceTaskHostname: true,
		Basic: &marathon.Basic{
			HTTPBasicAuthUser: "marathon HTTPBasicAuthUser",
			HTTPBasicPassword: "marathon HTTPBasicPassword",
		},
		RespectReadinessChecks: true,
	}
	config.ConsulCatalog = &consulcatalog.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "ConsulCatalog Filename",
			Constraints: types.Constraints{
				{
					Key:       "ConsulCatalog Constraints Key 1",
					Regex:     "ConsulCatalog Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "ConsulCatalog Constraints Key 1",
					Regex:     "ConsulCatalog Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:         "ConsulCatalog Endpoint",
		Domain:           "ConsulCatalog Domain",
		ExposedByDefault: true,
		Prefix:           "ConsulCatalog Prefix",
		FrontEndRule:     "ConsulCatalog FrontEndRule",
	}
	config.Kubernetes = &kubernetes.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "k8s Filename",
			Constraints: types.Constraints{
				{
					Key:       "k8s Constraints Key 1",
					Regex:     "k8s Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "k8s Constraints Key 1",
					Regex:     "k8s Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:               "k8s Endpoint",
		Token:                  "k8s Token",
		CertAuthFilePath:       "k8s CertAuthFilePath",
		DisablePassHostHeaders: true,
		Namespaces:             kubernetes.Namespaces{"k8s Namespaces 1", "k8s Namespaces 2", "k8s Namespaces 3"},
		LabelSelector:          "k8s LabelSelector",
	}
	config.Mesos = &mesos.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "mesos Filename",
			Constraints: types.Constraints{
				{
					Key:       "mesos Constraints Key 1",
					Regex:     "mesos Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "mesos Constraints Key 1",
					Regex:     "mesos Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:           "mesos Endpoint",
		Domain:             "mesos Domain",
		ExposedByDefault:   true,
		GroupsAsSubDomains: true,
		ZkDetectionTimeout: 666,
		RefreshSeconds:     666,
		IPSources:          "mesos IPSources",
		StateTimeoutSecond: 666,
		Masters:            []string{"mesos Masters 1", "mesos Masters 2", "mesos Masters 3"},
	}
	config.Eureka = &eureka.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "eureka Filename",
			Constraints: types.Constraints{
				{
					Key:       "eureka Constraints Key 1",
					Regex:     "eureka Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "eureka Constraints Key 1",
					Regex:     "eureka Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:       "eureka Endpoint",
		Delay:          flaeg.Duration(30 * time.Second),
		RefreshSeconds: flaeg.Duration(30 * time.Second),
	}
	config.ECS = &ecs.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "ecs Filename",
			Constraints: types.Constraints{
				{
					Key:       "ecs Constraints Key 1",
					Regex:     "ecs Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "ecs Constraints Key 1",
					Regex:     "ecs Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Domain:               "ecs Domain",
		ExposedByDefault:     true,
		RefreshSeconds:       666,
		Clusters:             ecs.Clusters{"ecs Clusters 1", "ecs Clusters 2", "ecs Clusters 3"},
		Cluster:              "ecs Cluster",
		AutoDiscoverClusters: true,
		Region:               "ecs Region",
		AccessKeyID:          "ecs AccessKeyID",
		SecretAccessKey:      "ecs SecretAccessKey",
	}
	config.Rancher = &rancher.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "rancher Filename",
			Constraints: types.Constraints{
				{
					Key:       "rancher Constraints Key 1",
					Regex:     "rancher Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "rancher Constraints Key 1",
					Regex:     "rancher Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		APIConfiguration: rancher.APIConfiguration{
			Endpoint:  "rancher Endpoint",
			AccessKey: "rancher AccessKey",
			SecretKey: "rancher SecretKey",
		},
		API: &rancher.APIConfiguration{
			Endpoint:  "rancher Endpoint",
			AccessKey: "rancher AccessKey",
			SecretKey: "rancher SecretKey",
		},
		Metadata: &rancher.MetadataConfiguration{
			IntervalPoll: true,
			Prefix:       "rancher Metadata Prefix",
		},
		Domain:                    "rancher Domain",
		RefreshSeconds:            666,
		ExposedByDefault:          true,
		EnableServiceHealthFilter: true,
	}
	config.DynamoDB = &dynamodb.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "dynamodb Filename",
			Constraints: types.Constraints{
				{
					Key:       "dynamodb Constraints Key 1",
					Regex:     "dynamodb Constraints Regex 2",
					MustMatch: true,
				},
				{
					Key:       "dynamodb Constraints Key 1",
					Regex:     "dynamodb Constraints Regex 2",
					MustMatch: true,
				},
			},
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		AccessKeyID:     "dynamodb AccessKeyID",
		RefreshSeconds:  666,
		Region:          "dynamodb Region",
		SecretAccessKey: "dynamodb SecretAccessKey",
		TableName:       "dynamodb TableName",
		Endpoint:        "dynamodb Endpoint",
	}
	config.Etcd = &etcd.Provider{
		Provider: kv.Provider{
			BaseProvider: provider.BaseProvider{
				Watch:    true,
				Filename: "etcd Filename",
				Constraints: types.Constraints{
					{
						Key:       "etcd Constraints Key 1",
						Regex:     "etcd Constraints Regex 2",
						MustMatch: true,
					},
					{
						Key:       "etcd Constraints Key 1",
						Regex:     "etcd Constraints Regex 2",
						MustMatch: true,
					},
				},
				Trace:                     true,
				DebugLogGeneratedTemplate: true,
			},
			Endpoint: "etcd Endpoint",
			Prefix:   "etcd Prefix",
			TLS: &types.ClientTLS{
				CA:                 "etcd CA",
				Cert:               "etcd Cert",
				Key:                "etcd Key",
				InsecureSkipVerify: true,
			},
			Username: "etcd Username",
			Password: "etcd Password",
		},
	}
	config.Zookeeper = &zk.Provider{
		Provider: kv.Provider{
			BaseProvider: provider.BaseProvider{
				Watch:    true,
				Filename: "zk Filename",
				Constraints: types.Constraints{
					{
						Key:       "zk Constraints Key 1",
						Regex:     "zk Constraints Regex 2",
						MustMatch: true,
					},
					{
						Key:       "zk Constraints Key 1",
						Regex:     "zk Constraints Regex 2",
						MustMatch: true,
					},
				},
				Trace:                     true,
				DebugLogGeneratedTemplate: true,
			},
			Endpoint: "zk Endpoint",
			Prefix:   "zk Prefix",
			TLS: &types.ClientTLS{
				CA:                 "zk CA",
				Cert:               "zk Cert",
				Key:                "zk Key",
				InsecureSkipVerify: true,
			},
			Username: "zk Username",
			Password: "zk Password",
		},
	}
	config.Boltdb = &boltdb.Provider{
		Provider: kv.Provider{
			BaseProvider: provider.BaseProvider{
				Watch:    true,
				Filename: "boltdb Filename",
				Constraints: types.Constraints{
					{
						Key:       "boltdb Constraints Key 1",
						Regex:     "boltdb Constraints Regex 2",
						MustMatch: true,
					},
					{
						Key:       "boltdb Constraints Key 1",
						Regex:     "boltdb Constraints Regex 2",
						MustMatch: true,
					},
				},
				Trace:                     true,
				DebugLogGeneratedTemplate: true,
			},
			Endpoint: "boltdb Endpoint",
			Prefix:   "boltdb Prefix",
			TLS: &types.ClientTLS{
				CA:                 "boltdb CA",
				Cert:               "boltdb Cert",
				Key:                "boltdb Key",
				InsecureSkipVerify: true,
			},
			Username: "boltdb Username",
			Password: "boltdb Password",
		},
	}
	config.Consul = &consul.Provider{
		Provider: kv.Provider{
			BaseProvider: provider.BaseProvider{
				Watch:    true,
				Filename: "consul Filename",
				Constraints: types.Constraints{
					{
						Key:       "consul Constraints Key 1",
						Regex:     "consul Constraints Regex 2",
						MustMatch: true,
					},
					{
						Key:       "consul Constraints Key 1",
						Regex:     "consul Constraints Regex 2",
						MustMatch: true,
					},
				},
				Trace:                     true,
				DebugLogGeneratedTemplate: true,
			},
			Endpoint: "consul Endpoint",
			Prefix:   "consul Prefix",
			TLS: &types.ClientTLS{
				CA:                 "consul CA",
				Cert:               "consul Cert",
				Key:                "consul Key",
				InsecureSkipVerify: true,
			},
			Username: "consul Username",
			Password: "consul Password",
		},
	}

	cleanJSON, err := Do(config, true)
	if err != nil {
		t.Fatal(err, cleanJSON)
	}
}
