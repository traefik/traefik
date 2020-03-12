package traefik

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/ping"
	"github.com/containous/traefik/v2/pkg/provider/rest"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in fixtures")

func Test_createConfiguration(t *testing.T) {
	testCases := []struct {
		desc      string
		staticCfg static.Configuration
	}{
		{
			desc: "full_configuration.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  true,
					Dashboard: true,
					Debug:     true,
				},
				Ping: &ping.Handler{
					EntryPoint:    "test",
					ManualRouting: false,
				},
				Providers: &static.Providers{
					Rest: &rest.Provider{
						Insecure: true,
					},
				},
				Metrics: &types.Metrics{
					Prometheus: &types.Prometheus{
						EntryPoint:    "test",
						ManualRouting: false,
					},
				},
			},
		},
		{
			desc: "full_configuration_secure.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  false,
					Dashboard: true,
				},
				Ping: &ping.Handler{
					EntryPoint:    "test",
					ManualRouting: true,
				},
				Providers: &static.Providers{
					Rest: &rest.Provider{
						Insecure: false,
					},
				},
				Metrics: &types.Metrics{
					Prometheus: &types.Prometheus{
						EntryPoint:    "test",
						ManualRouting: true,
					},
				},
			},
		},
		{
			desc: "api_insecure_with_dashboard.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  true,
					Dashboard: true,
				},
			},
		},
		{
			desc: "api_insecure_without_dashboard.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  true,
					Dashboard: false,
				},
			},
		},
		{
			desc: "api_secure_with_dashboard.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  false,
					Dashboard: true,
				},
			},
		},
		{
			desc: "api_secure_without_dashboard.json",
			staticCfg: static.Configuration{
				API: &static.API{
					Insecure:  false,
					Dashboard: false,
				},
			},
		},
		{
			desc: "ping_simple.json",
			staticCfg: static.Configuration{
				Ping: &ping.Handler{
					EntryPoint:    "test",
					ManualRouting: false,
				},
			},
		},
		{
			desc: "ping_custom.json",
			staticCfg: static.Configuration{
				Ping: &ping.Handler{
					EntryPoint:    "test",
					ManualRouting: true,
				},
			},
		},
		{
			desc: "rest_insecure.json",
			staticCfg: static.Configuration{
				Providers: &static.Providers{
					Rest: &rest.Provider{
						Insecure: true,
					},
				},
			},
		},
		{
			desc: "rest_secure.json",
			staticCfg: static.Configuration{
				Providers: &static.Providers{
					Rest: &rest.Provider{
						Insecure: false,
					},
				},
			},
		},
		{
			desc: "prometheus_simple.json",
			staticCfg: static.Configuration{
				Metrics: &types.Metrics{
					Prometheus: &types.Prometheus{
						EntryPoint:    "test",
						ManualRouting: false,
					},
				},
			},
		},
		{
			desc: "prometheus_custom.json",
			staticCfg: static.Configuration{
				Metrics: &types.Metrics{
					Prometheus: &types.Prometheus{
						EntryPoint:    "test",
						ManualRouting: true,
					},
				},
			},
		},
		{
			desc: "models.json",
			staticCfg: static.Configuration{
				EntryPoints: map[string]*static.EntryPoint{
					"websecure": {
						HTTP: static.HTTPConfig{
							Middlewares: []string{"test"},
							TLS: &static.TLSConfig{
								Options:      "opt",
								CertResolver: "le",
								Domains: []types.Domain{
									{Main: "mainA", SANs: []string{"sanA1", "sanA2"}},
									{Main: "mainB", SANs: []string{"sanB1", "sanB2"}},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "redirection.json",
			staticCfg: static.Configuration{
				EntryPoints: map[string]*static.EntryPoint{
					"web": {
						Address: ":80",
						HTTP: static.HTTPConfig{
							Redirections: &static.Redirections{
								EntryPoint: &static.RedirectEntryPoint{
									To:     "websecure",
									Scheme: "https",
								},
							},
						},
					},
					"websecure": {
						Address: ":443",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := Provider{staticCfg: test.staticCfg}

			cfg := provider.createConfiguration()

			filename := filepath.Join("fixtures", test.desc)

			if *updateExpected {
				newJSON, err := json.MarshalIndent(cfg, "", "  ")
				require.NoError(t, err)

				err = ioutil.WriteFile(filename, newJSON, 0644)
				require.NoError(t, err)
			}

			expectedJSON, err := ioutil.ReadFile(filename)
			require.NoError(t, err)

			actualJSON, err := json.MarshalIndent(cfg, "", "  ")
			require.NoError(t, err)

			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
		})
	}
}
