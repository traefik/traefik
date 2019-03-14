package anonymize

import (
	"os"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/ping"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/acme"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/provider/docker"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/kubernetes/crd"
	"github.com/containous/traefik/provider/kubernetes/ingress"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tracing/datadog"
	"github.com/containous/traefik/tracing/instana"
	"github.com/containous/traefik/tracing/jaeger"
	"github.com/containous/traefik/tracing/zipkin"
	"github.com/containous/traefik/types"
	assetfs "github.com/elazarl/go-bindata-assetfs"
)

func TestDo_globalConfiguration(t *testing.T) {

	config := &static.Configuration{}

	sendAnonymousUsage := true
	config.Global = &static.Global{
		Debug:              true,
		CheckNewVersion:    true,
		SendAnonymousUsage: &sendAnonymousUsage,
	}

	config.AccessLog = &types.AccessLog{
		FilePath: "AccessLog FilePath",
		Format:   "AccessLog Format",
		Filters: &types.AccessLogFilters{
			StatusCodes:   types.StatusCodes{"200", "500"},
			RetryAttempts: true,
			MinDuration:   10,
		},
		Fields: &types.AccessLogFields{
			DefaultMode: "drop",
			Names: types.FieldNames{
				"RequestHost": "keep",
			},
			Headers: &types.FieldHeaders{
				DefaultMode: "drop",
				Names: types.FieldHeaderNames{
					"Referer": "keep",
				},
			},
		},
		BufferingSize: 4,
	}

	config.Log = &types.TraefikLog{
		LogLevel: "LogLevel",
		FilePath: "/foo/path",
		Format:   "json",
	}

	config.EntryPoints = static.EntryPoints{
		"foo": {
			Address: "foo Address",
			Transport: &static.EntryPointsTransport{
				RespondingTimeouts: &static.RespondingTimeouts{
					ReadTimeout:  parse.Duration(111 * time.Second),
					WriteTimeout: parse.Duration(111 * time.Second),
					IdleTimeout:  parse.Duration(111 * time.Second),
				},
			},
			ProxyProtocol: &static.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
		},
		"fii": {
			Address: "fii Address",
			Transport: &static.EntryPointsTransport{
				RespondingTimeouts: &static.RespondingTimeouts{
					ReadTimeout:  parse.Duration(111 * time.Second),
					WriteTimeout: parse.Duration(111 * time.Second),
					IdleTimeout:  parse.Duration(111 * time.Second),
				},
			},
			ProxyProtocol: &static.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
		},
	}
	config.ACME = &acme.Configuration{
		Email:        "acme Email",
		ACMELogging:  true,
		CAServer:     "CAServer",
		Storage:      "Storage",
		EntryPoint:   "EntryPoint",
		KeyType:      "MyKeyType",
		OnHostRule:   true,
		DNSChallenge: &acmeprovider.DNSChallenge{Provider: "DNSProvider"},
		HTTPChallenge: &acmeprovider.HTTPChallenge{
			EntryPoint: "MyEntryPoint",
		},
		TLSChallenge: &acmeprovider.TLSChallenge{},
		Domains: []types.Domain{
			{
				Main: "Domains Main",
				SANs: []string{"Domains acme SANs 1", "Domains acme SANs 2", "Domains acme SANs 3"},
			},
		},
	}
	config.Providers = &static.Providers{
		ProvidersThrottleDuration: parse.Duration(111 * time.Second),
	}

	config.ServersTransport = &static.ServersTransport{
		InsecureSkipVerify:  true,
		RootCAs:             traefiktls.FilesOrContents{"RootCAs 1", "RootCAs 2", "RootCAs 3"},
		MaxIdleConnsPerHost: 111,
		ForwardingTimeouts: &static.ForwardingTimeouts{
			DialTimeout:           parse.Duration(111 * time.Second),
			ResponseHeaderTimeout: parse.Duration(111 * time.Second),
		},
	}

	config.API = &static.API{
		EntryPoint: "traefik",
		Dashboard:  true,
		Statistics: &types.Statistics{
			RecentErrors: 111,
		},
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
		Middlewares: []string{"first", "second"},
	}

	config.Providers.File = &file.Provider{
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

	config.Providers.Docker = &docker.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:                     true,
			Filename:                  "myfilename",
			Constraints:               nil,
			Trace:                     true,
			DebugLogGeneratedTemplate: true,
		},
		Endpoint:    "MyEndPoint",
		DefaultRule: "PathPrefix(`/`)",
		TLS: &types.ClientTLS{
			CA:                 "myCa",
			CAOptional:         true,
			Cert:               "mycert.pem",
			Key:                "mycert.key",
			InsecureSkipVerify: true,
		},
		ExposedByDefault:        true,
		UseBindPortIP:           true,
		SwarmMode:               true,
		Network:                 "MyNetwork",
		SwarmModeRefreshSeconds: 42,
	}

	config.Providers.Kubernetes = &ingress.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "myFileName",
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
		Endpoint:               "MyEndpoint",
		Token:                  "MyToken",
		CertAuthFilePath:       "MyCertAuthPath",
		DisablePassHostHeaders: true,
		EnablePassTLSCert:      true,
		Namespaces:             []string{"a", "b"},
		LabelSelector:          "myLabelSelector",
		IngressClass:           "MyIngressClass",
	}

	config.Providers.KubernetesCRD = &crd.Provider{
		BaseProvider: provider.BaseProvider{
			Watch:    true,
			Filename: "myFileName",
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
		Endpoint:               "MyEndpoint",
		Token:                  "MyToken",
		CertAuthFilePath:       "MyCertAuthPath",
		DisablePassHostHeaders: true,
		EnablePassTLSCert:      true,
		Namespaces:             []string{"a", "b"},
		LabelSelector:          "myLabelSelector",
		IngressClass:           "MyIngressClass",
	}

	// FIXME Test the other providers once they are migrated

	config.Metrics = &types.Metrics{
		Prometheus: &types.Prometheus{
			Buckets:     types.Buckets{0.1, 0.3, 1.2, 5},
			EntryPoint:  "MyEntryPoint",
			Middlewares: []string{"m1", "m2"},
		},
		Datadog: &types.Datadog{
			Address:      "localhost:8181",
			PushInterval: "12",
		},
		StatsD: &types.Statsd{
			Address:      "localhost:8182",
			PushInterval: "42",
		},
		InfluxDB: &types.InfluxDB{
			Address:         "localhost:8183",
			Protocol:        "http",
			PushInterval:    "22",
			Database:        "myDB",
			RetentionPolicy: "12",
			Username:        "a",
			Password:        "aaaa",
		},
	}

	config.Ping = &ping.Handler{
		EntryPoint:  "MyEntryPoint",
		Middlewares: []string{"m1", "m2", "m3"},
	}

	config.Tracing = &static.Tracing{
		Backend:       "myBackend",
		ServiceName:   "myServiceName",
		SpanNameLimit: 3,
		Jaeger: &jaeger.Config{
			SamplingServerURL:      "aaa",
			SamplingType:           "bbb",
			SamplingParam:          43,
			LocalAgentHostPort:     "ccc",
			Gen128Bit:              true,
			Propagation:            "ddd",
			TraceContextHeaderName: "eee",
		},
		Zipkin: &zipkin.Config{
			HTTPEndpoint: "fff",
			SameSpan:     true,
			ID128Bit:     true,
			Debug:        true,
			SampleRate:   53,
		},
		DataDog: &datadog.Config{
			LocalAgentHostPort: "ggg",
			GlobalTag:          "eee",
			Debug:              true,
			PrioritySampling:   true,
		},
		Instana: &instana.Config{
			LocalAgentHost: "fff",
			LocalAgentPort: 32,
			LogLevel:       "ggg",
		},
	}

	config.HostResolver = &types.HostResolverConfig{
		CnameFlattening: true,
		ResolvConfig:    "aaa",
		ResolvDepth:     3,
	}

	cleanJSON, err := Do(config, true)
	if err != nil {
		t.Fatal(err, cleanJSON)
	}
}
