package anonymize

import (
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/provider"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/provider/file"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/elazarl/go-bindata-assetfs"
)

func TestDo_globalConfiguration(t *testing.T) {

	config := &static.Configuration{}

	config.Global = &static.Global{
		Debug:              true,
		CheckNewVersion:    true,
		SendAnonymousUsage: true,
	}
	config.AccessLog = &types.AccessLog{
		FilePath: "AccessLog FilePath",
		Format:   "AccessLog Format",
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
			TLS: &traefiktls.TLS{
				MinVersion:   "foo MinVersion",
				CipherSuites: []string{"foo CipherSuites 1", "foo CipherSuites 2", "foo CipherSuites 3"},
				ClientCA: traefiktls.ClientCA{
					Files:    traefiktls.FilesOrContents{"foo ClientCAFiles 1", "foo ClientCAFiles 2", "foo ClientCAFiles 3"},
					Optional: false,
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
			TLS: &traefiktls.TLS{
				MinVersion:   "fii MinVersion",
				CipherSuites: []string{"fii CipherSuites 1", "fii CipherSuites 2", "fii CipherSuites 3"},
				ClientCA: traefiktls.ClientCA{
					Files:    traefiktls.FilesOrContents{"fii ClientCAFiles 1", "fii ClientCAFiles 2", "fii ClientCAFiles 3"},
					Optional: false,
				},
			},
			ProxyProtocol: &static.ProxyProtocol{
				TrustedIPs: []string{"127.0.0.1/32", "192.168.0.1"},
			},
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
		Storage:      "Storage",
		StorageFile:  "StorageFile",
		OnDemand:     true,
		OnHostRule:   true,
		CAServer:     "CAServer",
		EntryPoint:   "EntryPoint",
		DNSChallenge: &acmeprovider.DNSChallenge{Provider: "DNSProvider"},
		ACMELogging:  true,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			// ...
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

	// FIXME Test the other providers once they are migrated

	cleanJSON, err := Do(config, true)
	if err != nil {
		t.Fatal(err, cleanJSON)
	}
}
