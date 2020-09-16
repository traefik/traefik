package anonymize

import (
	"os"
	"testing"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/static"
	"github.com/traefik/traefik/v2/pkg/ping"
	"github.com/traefik/traefik/v2/pkg/provider/acme"
	"github.com/traefik/traefik/v2/pkg/provider/docker"
	"github.com/traefik/traefik/v2/pkg/provider/file"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/ingress"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"github.com/traefik/traefik/v2/pkg/tracing/datadog"
	"github.com/traefik/traefik/v2/pkg/tracing/haystack"
	"github.com/traefik/traefik/v2/pkg/tracing/instana"
	"github.com/traefik/traefik/v2/pkg/tracing/jaeger"
	"github.com/traefik/traefik/v2/pkg/tracing/zipkin"
	"github.com/traefik/traefik/v2/pkg/types"
)

func TestDo_globalConfiguration(t *testing.T) {
	config := &static.Configuration{}

	config.Global = &static.Global{
		CheckNewVersion:    true,
		SendAnonymousUsage: true,
	}

	config.AccessLog = &types.AccessLog{
		FilePath: "AccessLog FilePath",
		Format:   "AccessLog Format",
		Filters: &types.AccessLogFilters{
			StatusCodes:   []string{"200", "500"},
			RetryAttempts: true,
			MinDuration:   10,
		},
		Fields: &types.AccessLogFields{
			DefaultMode: "drop",
			Names: map[string]string{
				"RequestHost": "keep",
			},
			Headers: &types.FieldHeaders{
				DefaultMode: "drop",
				Names: map[string]string{
					"Referer": "keep",
				},
			},
		},
		BufferingSize: 4,
	}

	config.Log = &types.TraefikLog{
		Level:    "Level",
		FilePath: "/foo/path",
		Format:   "json",
	}

	config.EntryPoints = static.EntryPoints{
		"foo": {
			Address: "foo Address",
			Transport: &static.EntryPointsTransport{
				RespondingTimeouts: &static.RespondingTimeouts{
					ReadTimeout:  ptypes.Duration(111 * time.Second),
					WriteTimeout: ptypes.Duration(111 * time.Second),
					IdleTimeout:  ptypes.Duration(111 * time.Second),
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
					ReadTimeout:  ptypes.Duration(111 * time.Second),
					WriteTimeout: ptypes.Duration(111 * time.Second),
					IdleTimeout:  ptypes.Duration(111 * time.Second),
				},
			},
			ProxyProtocol: &static.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
		},
	}
	config.CertificatesResolvers = map[string]static.CertificateResolver{
		"default": {
			ACME: &acme.Configuration{
				Email:        "acme Email",
				CAServer:     "CAServer",
				Storage:      "Storage",
				KeyType:      "MyKeyType",
				DNSChallenge: &acme.DNSChallenge{Provider: "DNSProvider"},
				HTTPChallenge: &acme.HTTPChallenge{
					EntryPoint: "MyEntryPoint",
				},
				TLSChallenge: &acme.TLSChallenge{},
			},
		},
	}
	config.Providers = &static.Providers{
		ProvidersThrottleDuration: ptypes.Duration(111 * time.Second),
	}

	config.ServersTransport = &static.ServersTransport{
		InsecureSkipVerify:  true,
		RootCAs:             []traefiktls.FileOrContent{"RootCAs 1", "RootCAs 2", "RootCAs 3"},
		MaxIdleConnsPerHost: 111,
		ForwardingTimeouts: &static.ForwardingTimeouts{
			DialTimeout:           ptypes.Duration(111 * time.Second),
			ResponseHeaderTimeout: ptypes.Duration(111 * time.Second),
		},
	}

	config.API = &static.API{
		Dashboard: true,
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

	config.Providers.File = &file.Provider{
		Directory:                 "file Directory",
		Watch:                     true,
		Filename:                  "file Filename",
		DebugLogGeneratedTemplate: true,
	}

	config.Providers.Docker = &docker.Provider{
		Constraints: `Label("foo", "bar")`,
		Watch:       true,
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

	config.Providers.KubernetesIngress = &ingress.Provider{
		Endpoint:               "MyEndpoint",
		Token:                  "MyToken",
		CertAuthFilePath:       "MyCertAuthPath",
		DisablePassHostHeaders: true,
		Namespaces:             []string{"a", "b"},
		LabelSelector:          "myLabelSelector",
		IngressClass:           "MyIngressClass",
	}

	config.Providers.KubernetesCRD = &crd.Provider{
		Endpoint:               "MyEndpoint",
		Token:                  "MyToken",
		CertAuthFilePath:       "MyCertAuthPath",
		DisablePassHostHeaders: true,
		Namespaces:             []string{"a", "b"},
		LabelSelector:          "myLabelSelector",
		IngressClass:           "MyIngressClass",
	}

	// FIXME Test the other providers once they are migrated

	config.Metrics = &types.Metrics{
		Prometheus: &types.Prometheus{
			Buckets: []float64{0.1, 0.3, 1.2, 5},
		},
		Datadog: &types.Datadog{
			Address:      "localhost:8181",
			PushInterval: 12,
		},
		StatsD: &types.Statsd{
			Address:      "localhost:8182",
			PushInterval: 42,
		},
		InfluxDB: &types.InfluxDB{
			Address:         "localhost:8183",
			Protocol:        "http",
			PushInterval:    22,
			Database:        "myDB",
			RetentionPolicy: "12",
			Username:        "a",
			Password:        "aaaa",
		},
	}

	config.Ping = &ping.Handler{}

	config.Tracing = &static.Tracing{
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
			SampleRate:   53,
		},
		Datadog: &datadog.Config{
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
		Haystack: &haystack.Config{
			LocalAgentHost:          "fff",
			LocalAgentPort:          32,
			GlobalTag:               "eee",
			TraceIDHeaderName:       "fff",
			ParentIDHeaderName:      "ggg",
			SpanIDHeaderName:        "hhh",
			BaggagePrefixHeaderName: "iii",
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
