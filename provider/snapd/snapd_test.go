package snapd

import (
	"strconv"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		snaps             []snapData
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "when no snap",
			snaps:             []snapData{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when basic snap configuration",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort: "80",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend basic auth",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort:                          "80",
						label.TraefikFrontendAuthBasicUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile:    ".htpasswd",
						label.TraefikFrontendAuthBasicRemoveHeader: "true",
						label.TraefikFrontendAuthBasicRealm:        "myRealm",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Basic: &types.Basic{
							RemoveHeader: true,
							Realm:        "myRealm",
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when pass tls client certificate",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort:                                              "80",
						label.TraefikFrontendPassTLSClientCertPem:                      "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:           "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:            "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:   "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:      "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber: "true",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					PassTLSClientCert: &types.TLSClientHeaders{
						PEM: true,
						Infos: &types.TLSClientCertificateInfos{
							NotBefore: true,
							Sans:      true,
							NotAfter:  true,
							Subject: &types.TLSCLientCertificateSubjectInfos{
								CommonName:   true,
								Country:      true,
								Locality:     true,
								Organization: true,
								Province:     true,
								SerialNumber: true,
							},
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend basic auth backward compatibility",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort:              "80",
						label.TraefikFrontendAuthBasic: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Basic: &types.Basic{
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend digest auth",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort: "80",
						label.TraefikFrontendAuthDigestRemoveHeader: "true",
						label.TraefikFrontendAuthDigestUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:    ".htpasswd",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Digest: &types.Digest{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend forward auth",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikPort: "80",
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-snap-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Forward: &types.Forward{
							Address:            "auth.server",
							TrustForwardHeader: true,
							TLS: &types.ClientTLS{
								CA:                 "ca.crt",
								CAOptional:         true,
								InsecureSkipVerify: true,
								Cert:               "server.crt",
								Key:                "server.key",
							},
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-snap-localhost-0": {
							Rule: "Host:test.snap.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when snap has label 'enable' to false",
			snaps: []snapData{
				{
					SnapName: "test",
					Properties: map[string]string{
						label.TraefikEnable:   "false",
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",
						label.TraefikBackend:  "foobar",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			snaps: []snapData{
				{
					SnapName: "test1",
					Properties: map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendHealthCheckScheme:                "http",
						label.TraefikBackendHealthCheckPath:                  "/health",
						label.TraefikBackendHealthCheckPort:                  "880",
						label.TraefikBackendHealthCheckInterval:              "6",
						label.TraefikBackendHealthCheckTimeout:               "3",
						label.TraefikBackendHealthCheckHostname:              "foo.com",
						label.TraefikBackendHealthCheckHeaders:               "Foo:bar || Bar:foo",
						label.TraefikBackendLoadBalancerMethod:               "drr",
						label.TraefikBackendLoadBalancerStickiness:           "true",
						label.TraefikBackendLoadBalancerStickinessCookieName: "chocolate",
						label.TraefikBackendMaxConnAmount:                    "666",
						label.TraefikBackendMaxConnExtractorFunc:             "client.ip",
						label.TraefikBackendBufferingMaxResponseBodyBytes:    "10485760",
						label.TraefikBackendBufferingMemResponseBodyBytes:    "2097152",
						label.TraefikBackendBufferingMaxRequestBodyBytes:     "10485760",
						label.TraefikBackendBufferingMemRequestBodyBytes:     "2097152",
						label.TraefikBackendBufferingRetryExpression:         "IsNetworkError() && Attempts() <= 2",

						label.TraefikFrontendPassTLSClientCertPem:                      "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:           "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:            "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:   "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:      "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber: "true",

						label.TraefikFrontendAuthBasic:                        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicRealm:                   "myRealm",
						label.TraefikFrontendAuthBasicRemoveHeader:            "true",
						label.TraefikFrontendAuthBasicUsers:                   "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile:               ".htpasswd",
						label.TraefikFrontendAuthDigestRemoveHeader:           "true",
						label.TraefikFrontendAuthDigestUsers:                  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:              ".htpasswd",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.TraefikFrontendAuthHeaderField:                  "X-WebAuth-User",

						label.TraefikFrontendEntryPoints:                    "http,https",
						label.TraefikFrontendPassHostHeader:                 "true",
						label.TraefikFrontendPassTLSCert:                    "true",
						label.TraefikFrontendPriority:                       "666",
						label.TraefikFrontendRedirectEntryPoint:             "https",
						label.TraefikFrontendRedirectRegex:                  "nope",
						label.TraefikFrontendRedirectReplacement:            "nope",
						label.TraefikFrontendRedirectPermanent:              "true",
						label.TraefikFrontendRule:                           "Host:traefik.io",
						label.TraefikFrontendWhiteListSourceRange:           "10.10.10.10",
						label.TraefikFrontendWhiteListIPStrategyExcludedIPS: "10.10.10.10,10.10.10.11",
						label.TraefikFrontendWhiteListIPStrategyDepth:       "5",

						label.TraefikFrontendRequestHeaders:          "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendResponseHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendAllowedHosts:            "foo,bar,bor",
						label.TraefikFrontendHostsProxyHeaders:       "foo,bar,bor",
						label.TraefikFrontendSSLHost:                 "foo",
						label.TraefikFrontendCustomFrameOptionsValue: "foo",
						label.TraefikFrontendContentSecurityPolicy:   "foo",
						label.TraefikFrontendPublicKey:               "foo",
						label.TraefikFrontendReferrerPolicy:          "foo",
						label.TraefikFrontendCustomBrowserXSSValue:   "foo",
						label.TraefikFrontendSTSSeconds:              "666",
						label.TraefikFrontendSSLForceHost:            "true",
						label.TraefikFrontendSSLRedirect:             "true",
						label.TraefikFrontendSSLTemporaryRedirect:    "true",
						label.TraefikFrontendSTSIncludeSubdomains:    "true",
						label.TraefikFrontendSTSPreload:              "true",
						label.TraefikFrontendForceSTSHeader:          "true",
						label.TraefikFrontendFrameDeny:               "true",
						label.TraefikFrontendContentTypeNosniff:      "true",
						label.TraefikFrontendBrowserXSSFilter:        "true",
						label.TraefikFrontendIsDevelopment:           "true",

						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",

						label.TraefikFrontendRateLimitExtractorFunc:                                        "client.ip",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-traefik-io-0": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-traefik-io-0": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					PassTLSClientCert: &types.TLSClientHeaders{
						PEM: true,
						Infos: &types.TLSClientCertificateInfos{
							NotBefore: true,
							Sans:      true,
							NotAfter:  true,
							Subject: &types.TLSCLientCertificateSubjectInfos{
								CommonName:   true,
								Country:      true,
								Locality:     true,
								Organization: true,
								Province:     true,
								SerialNumber: true,
							},
						},
					},
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
						Basic: &types.Basic{
							Realm:        "myRealm",
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					WhiteList: &types.WhiteList{
						SourceRange: []string{"10.10.10.10"},
						IPStrategy: &types.IPStrategy{
							Depth:       5,
							ExcludedIPs: []string{"10.10.10.10", "10.10.10.11"},
						},
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
						SSLForceHost:         true,
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
							Backend: "backend-foobar",
						},
						"bar": {
							Status:  []string{"500", "600"},
							Query:   "bar_query",
							Backend: "backend-foobar",
						},
					},
					RateLimit: &types.RateLimit{
						ExtractorFunc: "client.ip",
						RateSet: map[string]*types.Rate{
							"foo": {
								Period:  parse.Duration(6 * time.Second),
								Average: 12,
								Burst:   18,
							},
							"bar": {
								Period:  parse.Duration(3 * time.Second),
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
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1-7f6444e0dff3330c8b0ad2bbbd383b0f": {
							URL:    "https://127.0.0.1:666",
							Weight: 12,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
						Stickiness: &types.Stickiness{
							CookieName: "chocolate",
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        666,
						ExtractorFunc: "client.ip",
					},
					HealthCheck: &types.HealthCheck{
						Scheme:   "http",
						Path:     "/health",
						Port:     880,
						Interval: "6",
						Timeout:  "3",
						Hostname: "foo.com",
						Headers: map[string]string{
							"Foo": "bar",
							"Bar": "foo",
						},
					},
					Buffering: &types.Buffering{
						MaxResponseBodyBytes: 10485760,
						MemResponseBodyBytes: 2097152,
						MaxRequestBodyBytes:  10485760,
						MemRequestBodyBytes:  2097152,
						RetryExpression:      "IsNetworkError() && Attempts() <= 2",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				Domain:           "snap.localhost",
				ExposedByDefault: true,
			}
			actualConfig := provider.buildConfiguration(test.snaps)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestSnapTraefikFilter(t *testing.T) {
	testCases := []struct {
		snap     snapData
		expected bool
		provider *Provider
	}{
		{
			snap: snapData{
				SnapName:   "snap",
				Properties: map[string]string{},
			},
			expected: false,
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "false",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: false,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:         "80",
					label.TraefikFrontendRule: "Host:foo.bar",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort: "80",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "anything",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:         "80",
					label.TraefikFrontendRule: "Host:foo.bar",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort: "80",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: true,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			snap: snapData{
				SnapName: "snap",
				Properties: map[string]string{
					label.TraefikPort:         "80",
					label.TraefikEnable:       "true",
					label.TraefikFrontendRule: "Host:i.love.this.host",
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: true,
		},
	}

	for snapID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(snapID), func(t *testing.T) {
			t.Parallel()

			actual := test.provider.containerFilter(test.snap)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestSnapGetFrontendName(t *testing.T) {
	testCases := []struct {
		snap     snapData
		expected string
	}{
		{
			snap:     snapData{SnapName: "foo"},
			expected: "Host-foo-snap-localhost-0",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
			}},
			expected: "Headers-User-Agent-bat-0-1-0-0",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			}},
			expected: "Host-foo-bar-0",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			}},
			expected: "Path-test-0",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "PathPrefix:/test2",
			}},
			expected: "PathPrefix-test2-0",
		},
	}

	for snapID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(snapID), func(t *testing.T) {
			t.Parallel()

			sData := test.snap
			segmentProperties := label.ExtractTraefikLabels(sData.Properties)
			sData.SegmentProperties = segmentProperties[""]

			provider := &Provider{
				Domain: "snap.localhost",
			}

			actual := provider.getFrontendName(sData, 0)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSnapGetFrontendRule(t *testing.T) {
	testCases := []struct {
		snap     snapData
		expected string
	}{
		{
			snap:     snapData{SnapName: "foo"},
			expected: "Host:foo.snap.localhost",
		},
		{
			snap: snapData{
				SnapName: "foo",
				Properties: map[string]string{
					label.TraefikDomain: "traefik.localhost",
				},
			},
			expected: "Host:foo.traefik.localhost",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			}},
			expected: "Host:foo.bar",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			}},
			expected: "Path:/test",
		},
	}

	for snapID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(snapID), func(t *testing.T) {
			t.Parallel()

			sData := test.snap
			segmentProperties := label.ExtractTraefikLabels(sData.Properties)

			provider := &Provider{
				Domain: "snap.localhost",
			}

			actual := provider.getFrontendRule(sData, segmentProperties[""])
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSnapGetBackendName(t *testing.T) {
	testCases := []struct {
		snap        snapData
		segmentName string
		expected    string
	}{
		{
			snap:     snapData{SnapName: "foo"},
			expected: "foo",
		},
		{
			snap:     snapData{SnapName: "bar"},
			expected: "bar",
		},
		{
			snap: snapData{Properties: map[string]string{
				label.TraefikBackend: "foobar",
			}},
			expected: "foobar",
		},
		{
			snap: snapData{
				SnapName: "bar",
				Properties: map[string]string{
					"traefik.sauternes.backend": "titi",
				},
			},
			segmentName: "sauternes",
			expected:    "bar-titi",
		},
	}

	for snapID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(snapID), func(t *testing.T) {
			t.Parallel()

			sData := test.snap
			segmentProperties := label.ExtractTraefikLabels(sData.Properties)
			sData.SegmentProperties = segmentProperties[test.segmentName]
			sData.SegmentName = test.segmentName

			actual := getBackendName(sData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSnapGetServers(t *testing.T) {
	p := &Provider{}

	testCases := []struct {
		desc     string
		snaps    []snapData
		expected map[string]types.Server
	}{
		{
			desc:     "no snap",
			expected: nil,
		},
		{
			desc: "with a simple snap",
			snaps: []snapData{
				{
					SnapName: "test1",
					Properties: map[string]string{
						label.TraefikPort: "80",
					},
				},
			},
			expected: map[string]types.Server{
				"server-test1-842895ca2aca17f6ee36ddb2f621194d": {
					URL:    "http://127.0.0.1:80",
					Weight: 1,
				},
			},
		},
		{
			desc: "with several snaps",
			snaps: []snapData{
				{
					SnapName: "test1",
					Properties: map[string]string{
						label.TraefikPort: "80",
					},
				},
				{
					SnapName: "test2",
					Properties: map[string]string{
						label.TraefikPort: "81",
					},
				},
				{
					SnapName: "test3",
					Properties: map[string]string{
						label.TraefikPort: "82",
					},
				},
			},
			expected: map[string]types.Server{
				"server-test1-842895ca2aca17f6ee36ddb2f621194d": {
					URL:    "http://127.0.0.1:80",
					Weight: 1,
				},
				"server-test2-789c09d92dae4d471033262d4d8b46ae": {
					URL:    "http://127.0.0.1:81",
					Weight: 1,
				},
				"server-test3-ea43c74c8e2b538ce33acc18f19382d6": {
					URL:    "http://127.0.0.1:82",
					Weight: 1,
				},
			},
		},
		{
			desc: "ignore one snap because no port",
			snaps: []snapData{
				{
					SnapName: "test1",
				},
				{
					SnapName: "test2",
					Properties: map[string]string{
						label.TraefikPort: "81",
					},
				},
				{
					SnapName: "test3",
					Properties: map[string]string{
						label.TraefikPort: "82",
					},
				},
			},
			expected: map[string]types.Server{
				"server-test2-789c09d92dae4d471033262d4d8b46ae": {
					URL:    "http://127.0.0.1:81",
					Weight: 1,
				},
				"server-test3-ea43c74c8e2b538ce33acc18f19382d6": {
					URL:    "http://127.0.0.1:82",
					Weight: 1,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			servers := p.getServers(test.snaps)

			assert.Equal(t, test.expected, servers)
		})
	}
}
