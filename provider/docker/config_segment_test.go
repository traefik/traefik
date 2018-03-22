package docker

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "when no container",
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "simple configuration",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                 "2503",
						"traefik.sauternes.frontend.entryPoints": "http,https",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-foo-sauternes": {
					Backend:        "backend-foo-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-foo-foo-sauternes": {
							Rule: "Host:foo.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-sauternes-foo-0": {
							URL:    "http://127.0.0.1:2503",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when all labels are set",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("foo"),
					labels(map[string]string{
						label.Prefix + "sauternes." + label.SuffixPort:     "666",
						label.Prefix + "sauternes." + label.SuffixProtocol: "https",
						label.Prefix + "sauternes." + label.SuffixWeight:   "12",

						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasic:            "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendEntryPoints:          "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendPassHostHeader:       "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSCert:          "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPriority:             "666",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectEntryPoint:   "https",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectRegex:        "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectReplacement:  "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectPermanent:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendWhitelistSourceRange: "10.10.10.10",

						label.Prefix + "sauternes." + label.SuffixFrontendRequestHeaders:                 "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendResponseHeaders:                "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersAllowedHosts:            "foo,bar,bor",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersHostsProxyHeaders:       "foo,bar,bor",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLHost:                 "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersCustomFrameOptionsValue: "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersContentSecurityPolicy:   "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersPublicKey:               "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersReferrerPolicy:          "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersCustomBrowserXSSValue:   "foo",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSSeconds:              "666",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLRedirect:             "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLTemporaryRedirect:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSIncludeSubdomains:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSTSPreload:              "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersForceSTSHeader:          "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersFrameDeny:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersContentTypeNosniff:      "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersBrowserXSSFilter:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersIsDevelopment:           "true",

						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "sauternes." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",

						label.Prefix + "sauternes." + label.SuffixFrontendRateLimitExtractorFunc:                          "client.ip",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
						label.Prefix + "sauternes." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-foo-sauternes": {
					Backend: "backend-foo-foo-sauternes",
					EntryPoints: []string{
						"http",
						"https",
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					WhitelistSourceRange: []string{
						"10.10.10.10",
					},
					Headers: &types.Headers{
						CustomRequestHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						CustomResponseHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						AllowedHosts: []string{
							"foo",
							"bar",
							"bor",
						},
						HostsProxyHeaders: []string{
							"foo",
							"bar",
							"bor",
						},
						SSLRedirect:          true,
						SSLTemporaryRedirect: true,
						SSLHost:              "foo",
						SSLProxyHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						STSSeconds:              666,
						STSIncludeSubdomains:    true,
						STSPreload:              true,
						ForceSTSHeader:          true,
						FrameDeny:               true,
						CustomFrameOptionsValue: "foo",
						ContentTypeNosniff:      true,
						BrowserXSSFilter:        true,
						CustomBrowserXSSValue:   "foo",
						ContentSecurityPolicy:   "foo",
						PublicKey:               "foo",
						ReferrerPolicy:          "foo",
						IsDevelopment:           true,
					},
					Errors: map[string]*types.ErrorPage{
						"foo": {
							Status:  []string{"404"},
							Query:   "foo_query",
							Backend: "foobar",
						},
						"bar": {
							Status:  []string{"500", "600"},
							Query:   "bar_query",
							Backend: "foobar",
						},
					},
					RateLimit: &types.RateLimit{
						ExtractorFunc: "client.ip",
						RateSet: map[string]*types.Rate{
							"foo": {
								Period:  flaeg.Duration(6 * time.Second),
								Average: 12,
								Burst:   18,
							},
							"bar": {
								Period:  flaeg.Duration(3 * time.Second),
								Average: 6,
								Burst:   9,
							},
						},
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "",
						Replacement: "",
						Permanent:   true,
					},

					Routes: map[string]types.Route{
						"route-frontend-foo-foo-sauternes": {
							Rule: "Host:foo.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-sauternes-foo-0": {
							URL:    "https://127.0.0.1:666",
							Weight: 12,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "several containers",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						"traefik.sauternes.port":                         "2503",
						"traefik.sauternes.protocol":                     "https",
						"traefik.sauternes.weight":                       "80",
						"traefik.sauternes.frontend.backend":             "foobar",
						"traefik.sauternes.frontend.passHostHeader":      "false",
						"traefik.sauternes.frontend.rule":                "Path:/mypath",
						"traefik.sauternes.frontend.priority":            "5000",
						"traefik.sauternes.frontend.entryPoints":         "http,https,ws",
						"traefik.sauternes.frontend.auth.basic":          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.sauternes.frontend.redirect.entryPoint": "https",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test2"),
					labels(map[string]string{
						"traefik.anothersauternes.port":          "8079",
						"traefik.anothersauternes.weight":        "33",
						"traefik.anothersauternes.frontend.rule": "Path:/anotherpath",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: false,
					Priority:       5000,
					EntryPoints:    []string{"http", "https", "ws"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-test1-foobar": {
							Rule: "Path:/mypath",
						},
					},
				},
				"frontend-test2-test2-anothersauternes": {
					Backend:        "backend-test2-test2-anothersauternes",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-test2-test2-anothersauternes": {
							Rule: "Path:/anotherpath",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"server-sauternes-test1-0": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-test2-test2-anothersauternes": {
					Servers: map[string]types.Server{
						"server-anothersauternes-test2-0": {
							URL:    "http://127.0.0.1:8079",
							Weight: 33,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	provider := &Provider{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, container := range test.containers {
				dData := parseContainer(container)
				dockerDataList = append(dockerDataList, dData)
			}

			actualConfig := provider.buildConfigurationV2(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}
