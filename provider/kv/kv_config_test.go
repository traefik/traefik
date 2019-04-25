package kv

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/abronan/valkeyrie/store"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func aKVPair(key string, value string) *store.KVPair {
	return &store.KVPair{Key: key, Value: []byte(value)}
}

func TestProviderBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  KvError
		expected *types.Configuration
	}{
		{
			desc: "name with dot",
			kvPairs: filler("traefik",
				frontend("frontend.with.dot",
					withPair("backend", "backend.with.dot.too"),
					withPair("routes/route.with.dot/rule", "Host:test.localhost")),
				backend("backend.with.dot.too",
					withPair("servers/server.with.dot/url", "http://172.17.0.2:80"),
					withPair("servers/server.with.dot/weight", strconv.Itoa(label.DefaultWeight)),
					withPair("servers/server.with.dot.without.url/weight", strconv.Itoa(label.DefaultWeight))),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend.with.dot.too": {
						LoadBalancer: &types.LoadBalancer{Method: label.DefaultBackendLoadBalancerMethod},
						Servers: map[string]types.Server{
							"server.with.dot": {
								URL:    "http://172.17.0.2:80",
								Weight: label.DefaultWeight,
							},
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend.with.dot": {
						Backend:        "backend.with.dot.too",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Routes: map[string]types.Route{
							"route.with.dot": {
								Rule: "Host:test.localhost",
							},
						},
					},
				},
			},
		},
		{
			desc: "basic auth Users",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
					withPair(pathFrontendAuthBasicRemoveHeader, "true"),
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				),
				backend("backend"),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend": {
						Backend:        "backend",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Auth: &types.Auth{
							HeaderField: "X-WebAuth-User",
							Basic: &types.Basic{
								RemoveHeader: true,
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
					},
				},
			},
		},
		{
			desc: "basic auth UsersFile",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
					withPair(pathFrontendAuthBasicUsersFile, ".htpasswd"),
				),
				backend("backend"),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend": {
						Backend:        "backend",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Auth: &types.Auth{
							HeaderField: "X-WebAuth-User",
							Basic: &types.Basic{
								UsersFile: ".htpasswd",
							},
						},
					},
				},
			},
		},
		{
			desc: "basic auth (backward compatibility)",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withList(pathFrontendBasicAuth, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				),
				backend("backend"),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend": {
						Backend:        "backend",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Auth: &types.Auth{
							Basic: &types.Basic{
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
					},
				},
			},
		},
		{
			desc: "digest auth",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
					withPair(pathFrontendAuthDigestRemoveHeader, "true"),
					withList(pathFrontendAuthDigestUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthDigestUsersFile, ".htpasswd"),
				),
				backend("backend"),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend": {
						Backend:        "backend",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Auth: &types.Auth{
							HeaderField: "X-WebAuth-User",
							Digest: &types.Digest{
								RemoveHeader: true,
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
								UsersFile: ".htpasswd",
							},
						},
					},
				},
			},
		},
		{
			desc: "forward auth",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
					withPair(pathFrontendAuthForwardAddress, "auth.server"),
					withPair(pathFrontendAuthForwardTrustForwardHeader, "true"),
					withPair(pathFrontendAuthForwardTLSCa, "ca.crt"),
					withPair(pathFrontendAuthForwardTLSCaOptional, "true"),
					withPair(pathFrontendAuthForwardTLSCert, "server.crt"),
					withPair(pathFrontendAuthForwardTLSKey, "server.key"),
					withPair(pathFrontendAuthForwardTLSInsecureSkipVerify, "true"),
					withPair(pathFrontendAuthForwardAuthResponseHeaders, "X-Auth-User,X-Auth-Token"),
				),
				backend("backend"),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend": {
						LoadBalancer: &types.LoadBalancer{
							Method: "wrr",
						},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend": {
						Backend:        "backend",
						PassHostHeader: true,
						EntryPoints:    []string{},
						Auth: &types.Auth{
							HeaderField: "X-WebAuth-User",
							Forward: &types.Forward{
								Address: "auth.server",
								TLS: &types.ClientTLS{
									CA:                 "ca.crt",
									CAOptional:         true,
									InsecureSkipVerify: true,
									Cert:               "server.crt",
									Key:                "server.key",
								},
								TrustForwardHeader:  true,
								AuthResponseHeaders: []string{"X-Auth-User", "X-Auth-Token"},
							},
						},
					},
				},
			},
		},
		{
			desc: "all parameters",
			kvPairs: filler("traefik",
				backend("backend1",
					withPair(pathBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
					withPair(pathBackendLoadBalancerMethod, "drr"),
					withPair(pathBackendLoadBalancerSticky, "true"),
					withPair(pathBackendLoadBalancerStickiness, "true"),
					withPair(pathBackendLoadBalancerStickinessCookieName, "tomate"),
					withPair(pathBackendHealthCheckScheme, "http"),
					withPair(pathBackendHealthCheckPath, "/health"),
					withPair(pathBackendHealthCheckPort, "80"),
					withPair(pathBackendHealthCheckInterval, "30s"),
					withPair(pathBackendHealthCheckHostname, "foo.com"),
					withPair(pathBackendHealthCheckHeaders+"Foo", "bar"),
					withPair(pathBackendHealthCheckHeaders+"Bar", "foo"),
					withPair(pathBackendMaxConnAmount, "5"),
					withPair(pathBackendMaxConnExtractorFunc, "client.ip"),
					withPair(pathBackendBufferingMaxResponseBodyBytes, "10485760"),
					withPair(pathBackendBufferingMemResponseBodyBytes, "2097152"),
					withPair(pathBackendBufferingMaxRequestBodyBytes, "10485760"),
					withPair(pathBackendBufferingMemRequestBodyBytes, "2097152"),
					withPair(pathBackendBufferingRetryExpression, "IsNetworkError() && Attempts() <= 2"),
					withPair("servers/server1/url", "http://172.17.0.2:80"),
					withPair("servers/server1/weight", strconv.Itoa(label.DefaultWeight)),
					withPair("servers/server2/weight", strconv.Itoa(label.DefaultWeight))),
				frontend("frontend1",
					withPair(pathFrontendBackend, "backend1"),
					withPair(pathFrontendPriority, "6"),
					withPair(pathFrontendPassHostHeader, "false"),

					withPair(pathFrontendPassTLSClientCertPem, "true"),
					withPair(pathFrontendPassTLSClientCertInfosNotBefore, "true"),
					withPair(pathFrontendPassTLSClientCertInfosNotAfter, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSans, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerCommonName, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerCountry, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerDomainComponent, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerLocality, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerOrganization, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerProvince, "true"),
					withPair(pathFrontendPassTLSClientCertInfosIssuerSerialNumber, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectCommonName, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectCountry, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectDomainComponent, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectLocality, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectOrganization, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectProvince, "true"),
					withPair(pathFrontendPassTLSClientCertInfosSubjectSerialNumber, "true"),

					withPair(pathFrontendPassTLSCert, "true"),
					withList(pathFrontendEntryPoints, "http", "https"),
					withList(pathFrontendWhiteListSourceRange, "1.1.1.1/24", "1234:abcd::42/32"),
					withPair(pathFrontendWhiteListUseXForwardedFor, "true"),

					withList(pathFrontendBasicAuth, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthBasicRemoveHeader, "true"),
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthBasicUsersFile, ".htpasswd"),
					withPair(pathFrontendAuthDigestRemoveHeader, "true"),
					withList(pathFrontendAuthDigestUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthDigestUsersFile, ".htpasswd"),
					withPair(pathFrontendAuthForwardAddress, "auth.server"),
					withPair(pathFrontendAuthForwardTrustForwardHeader, "true"),
					withPair(pathFrontendAuthForwardTLSCa, "ca.crt"),
					withPair(pathFrontendAuthForwardTLSCaOptional, "true"),
					withPair(pathFrontendAuthForwardTLSCert, "server.crt"),
					withPair(pathFrontendAuthForwardTLSKey, "server.key"),
					withPair(pathFrontendAuthForwardTLSInsecureSkipVerify, "true"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),

					withPair(pathFrontendRedirectEntryPoint, "https"),
					withPair(pathFrontendRedirectRegex, "nope"),
					withPair(pathFrontendRedirectReplacement, "nope"),
					withPair(pathFrontendRedirectPermanent, "true"),
					withErrorPage("foo", "error", "/test1", "500-501", "503-599"),
					withErrorPage("bar", "error", "/test2", "400-405"),
					withRateLimit("client.ip",
						withLimit("foo", "6", "12", "18"),
						withLimit("bar", "3", "6", "9")),

					withPair(pathFrontendCustomRequestHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendCustomRequestHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendCustomRequestHeaders+"X-Custom-Header", "test"),
					withPair(pathFrontendCustomResponseHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendCustomResponseHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendCustomResponseHeaders+"X-Custom-Header", "test"),
					withPair(pathFrontendSSLProxyHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendSSLProxyHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendSSLProxyHeaders+"X-Custom-Header", "test"),
					withPair(pathFrontendAllowedHosts, "example.com, ssl.example.com"),
					withList(pathFrontendHostsProxyHeaders, "foo", "bar", "goo", "hor"),
					withPair(pathFrontendSTSSeconds, "666"),
					withPair(pathFrontendSSLHost, "foo"),
					withPair(pathFrontendCustomFrameOptionsValue, "foo"),
					withPair(pathFrontendContentSecurityPolicy, "foo"),
					withPair(pathFrontendPublicKey, "foo"),
					withPair(pathFrontendReferrerPolicy, "foo"),
					withPair(pathFrontendCustomBrowserXSSValue, "foo"),
					withPair(pathFrontendSSLForceHost, "true"),
					withPair(pathFrontendSSLRedirect, "true"),
					withPair(pathFrontendSSLTemporaryRedirect, "true"),
					withPair(pathFrontendSTSIncludeSubdomains, "true"),
					withPair(pathFrontendSTSPreload, "true"),
					withPair(pathFrontendForceSTSHeader, "true"),
					withPair(pathFrontendFrameDeny, "true"),
					withPair(pathFrontendContentTypeNosniff, "true"),
					withPair(pathFrontendBrowserXSSFilter, "true"),
					withPair(pathFrontendIsDevelopment, "true"),

					withPair("routes/route1/rule", "Host:test.localhost"),
					withPair("routes/route2/rule", "Path:/foo")),
				entry("tls/foo",
					withList("entrypoints", "http", "https"),
					withPair("certificate/certfile", "certfile1"),
					withPair("certificate/keyfile", "keyfile1")),
				entry("tls/bar",
					withList("entrypoints", "http", "https"),
					withPair("certificate/certfile", "certfile2"),
					withPair("certificate/keyfile", "keyfile2")),
			),
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend1": {
						Servers: map[string]types.Server{
							"server1": {
								URL:    "http://172.17.0.2:80",
								Weight: label.DefaultWeight,
							},
						},
						CircuitBreaker: &types.CircuitBreaker{
							Expression: "NetworkErrorRatio() > 1",
						},
						LoadBalancer: &types.LoadBalancer{
							Method: "drr",
							Sticky: true,
							Stickiness: &types.Stickiness{
								CookieName: "tomate",
							},
						},
						MaxConn: &types.MaxConn{
							Amount:        5,
							ExtractorFunc: "client.ip",
						},
						HealthCheck: &types.HealthCheck{
							Scheme:   "http",
							Path:     "/health",
							Port:     80,
							Interval: "30s",
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
				Frontends: map[string]*types.Frontend{
					"frontend1": {
						Priority:    6,
						EntryPoints: []string{"http", "https"},
						Backend:     "backend1",
						PassTLSCert: true,
						WhiteList: &types.WhiteList{
							SourceRange:      []string{"1.1.1.1/24", "1234:abcd::42/32"},
							UseXForwardedFor: true,
						},
						PassTLSClientCert: &types.TLSClientHeaders{
							PEM: true,
							Infos: &types.TLSClientCertificateInfos{
								NotBefore: true,
								Sans:      true,
								NotAfter:  true,
								Subject: &types.TLSCLientCertificateDNInfos{
									CommonName:      true,
									Country:         true,
									DomainComponent: true,
									Locality:        true,
									Organization:    true,
									Province:        true,
									SerialNumber:    true,
								},
								Issuer: &types.TLSCLientCertificateDNInfos{
									CommonName:      true,
									Country:         true,
									DomainComponent: true,
									Locality:        true,
									Organization:    true,
									Province:        true,
									SerialNumber:    true,
								},
							},
						},
						Auth: &types.Auth{
							HeaderField: "X-WebAuth-User",
							Basic: &types.Basic{
								RemoveHeader: true,
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
								UsersFile: ".htpasswd",
							},
						},
						Redirect: &types.Redirect{
							EntryPoint: "https",
							Permanent:  true,
						},
						Errors: map[string]*types.ErrorPage{
							"foo": {
								Backend: "error",
								Query:   "/test1",
								Status:  []string{"500-501", "503-599"},
							},
							"bar": {
								Backend: "error",
								Query:   "/test2",
								Status:  []string{"400-405"},
							},
						},
						RateLimit: &types.RateLimit{
							ExtractorFunc: "client.ip",
							RateSet: map[string]*types.Rate{
								"foo": {
									Average: 6,
									Burst:   12,
									Period:  flaeg.Duration(18 * time.Second),
								},
								"bar": {
									Average: 3,
									Burst:   6,
									Period:  flaeg.Duration(9 * time.Second),
								},
							},
						},
						Routes: map[string]types.Route{
							"route1": {
								Rule: "Host:test.localhost",
							},
							"route2": {
								Rule: "Path:/foo",
							},
						},
						Headers: &types.Headers{
							CustomRequestHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
								"X-Custom-Header":              "test",
							},
							CustomResponseHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
								"X-Custom-Header":              "test",
							},
							SSLProxyHeaders: map[string]string{
								"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
								"Content-Type":                 "application/json; charset=utf-8",
								"X-Custom-Header":              "test",
							},
							AllowedHosts:            []string{"example.com", "ssl.example.com"},
							HostsProxyHeaders:       []string{"foo", "bar", "goo", "hor"},
							STSSeconds:              666,
							SSLHost:                 "foo",
							CustomFrameOptionsValue: "foo",
							ContentSecurityPolicy:   "foo",
							PublicKey:               "foo",
							ReferrerPolicy:          "foo",
							CustomBrowserXSSValue:   "foo",
							SSLForceHost:            true,
							SSLRedirect:             true,
							SSLTemporaryRedirect:    true,
							STSIncludeSubdomains:    true,
							STSPreload:              true,
							ForceSTSHeader:          true,
							FrameDeny:               true,
							ContentTypeNosniff:      true,
							BrowserXSSFilter:        true,
							IsDevelopment:           true,
						},
					},
				},
				TLS: []*tls.Configuration{
					{
						EntryPoints: []string{"http", "https"},
						Certificate: &tls.Certificate{
							CertFile: "certfile2",
							KeyFile:  "keyfile2",
						},
					},
					{
						EntryPoints: []string{"http", "https"},
						Certificate: &tls.Certificate{
							CertFile: "certfile1",
							KeyFile:  "keyfile1",
						},
					},
				},
			},
		},
		{
			desc: "Should recover on panic",
			kvPairs: filler("traefik",
				frontend("frontend",
					withPair(pathFrontendBackend, "backend"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
					withPair(pathFrontendAuthBasicRemoveHeader, "true"),
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				),
				backend("backend"),
			),
			kvError: KvError{
				List: store.ErrNotReachable,
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				Prefix: "traefik",
				kvClient: &Mock{
					KVPairs: test.kvPairs,
					Error:   test.kvError,
				},
			}

			actual, err := p.buildConfiguration()
			if test.kvError.Get != nil || test.kvError.List != nil {
				require.Error(t, err)
				require.Nil(t, actual)
			} else {
				require.NoError(t, err)
				require.NotNil(t, actual)
				assert.EqualValues(t, test.expected.Backends, actual.Backends)
				assert.EqualValues(t, test.expected.Frontends, actual.Frontends)
			}

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestProviderListShouldPanic(t *testing.T) {
	testCases := []struct {
		desc    string
		panic   bool
		kvError error
	}{
		{
			desc:    "Should panic on an unexpected error",
			kvError: store.ErrBackendNotSupported,
			panic:   true,
		},
		{
			desc:    "Should not panic on an ErrKeyNotFound error",
			kvError: store.ErrKeyNotFound,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			kvPairs := []*store.KVPair{
				aKVPair("foo", "bar"),
			}
			p := &Provider{
				kvClient: newKvClientMock(kvPairs, test.kvError),
			}

			keyParts := []string{"foo"}
			if test.panic {
				assert.Panics(t, func() { p.list(keyParts...) })
			} else {
				assert.NotPanics(t, func() { p.list(keyParts...) })
			}
		})
	}
}

func TestProviderList(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc:     "empty key parts and empty store",
			keyParts: []string{},
			expected: []string{},
		},
		{
			desc:     "when non existing key and empty store",
			keyParts: []string{"traefik"},
			expected: []string{},
		},
		{
			desc: "when non existing key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"bar"},
			expected: []string{},
		},
		{
			desc: "when one key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"foo"},
			expected: []string{"foo"},
		},
		{
			desc: "when multiple sub keys and nested sub key",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar"),
				aKVPair("foo/baz/2", "bar"),
				aKVPair("foo/baz/biz/1", "bar"),
			},
			keyParts: []string{"foo", "/baz/"},
			expected: []string{"foo/baz/1", "foo/baz/2"},
		},
		{
			desc:    "when KV error key not found",
			kvError: store.ErrKeyNotFound,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo/baz/1"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.list(test.keyParts...)

			sort.Strings(test.expected)
			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetShouldPanic(t *testing.T) {
	testCases := []struct {
		desc    string
		panic   bool
		kvError error
	}{
		{
			desc:    "Should panic on an unexpected error",
			kvError: store.ErrBackendNotSupported,
			panic:   true,
		},
		{
			desc:    "Should not panic on an ErrKeyNotFound error",
			kvError: store.ErrKeyNotFound,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			kvPairs := []*store.KVPair{
				aKVPair("foo", "bar"),
			}
			p := &Provider{
				kvClient: newKvClientMock(kvPairs, test.kvError),
			}

			keyParts := []string{"foo"}
			if test.panic {
				assert.Panics(t, func() { p.get("", keyParts...) })
			} else {
				assert.NotPanics(t, func() { p.get("", keyParts...) })
			}
		})
	}
}

func TestProviderGet(t *testing.T) {
	testCases := []struct {
		desc         string
		kvPairs      []*store.KVPair
		storeType    store.Backend
		keyParts     []string
		defaultValue string
		kvError      error
		expected     string
	}{
		{
			desc:         "when empty key parts, empty store",
			defaultValue: "circle",
			keyParts:     []string{},
			expected:     "circle",
		},
		{
			desc:         "when non existing key",
			defaultValue: "circle",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"bar"},
			expected: "circle",
		},
		{
			desc: "when one part key",
			kvPairs: []*store.KVPair{
				aKVPair("foo", "bar"),
			},
			keyParts: []string{"foo"},
			expected: "bar",
		},
		{
			desc: "when several parts key",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo", "/baz/", "2"},
			expected: "bar2",
		},
		{
			desc:         "when several parts key, starts with /",
			defaultValue: "circle",
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"/foo", "/baz/", "2"},
			expected: "circle",
		},
		{
			desc:      "when several parts key starts with /, ETCD v2",
			storeType: store.ETCD,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"/foo", "/baz/", "2"},
			expected: "bar2",
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: []*store.KVPair{
				aKVPair("foo/baz/1", "bar1"),
				aKVPair("foo/baz/2", "bar2"),
				aKVPair("foo/baz/biz/1", "bar3"),
			},
			keyParts: []string{"foo/baz/1"},
			expected: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient:  newKvClientMock(test.kvPairs, test.kvError),
				storeType: test.storeType,
			}

			actual := p.get(test.defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key %v", test.keyParts)
		})
	}
}

func TestProviderLast(t *testing.T) {
	p := &Provider{}

	testCases := []struct {
		key      string
		expected string
	}{
		{
			key:      "",
			expected: "",
		},
		{
			key:      "foo",
			expected: "foo",
		},
		{
			key:      "foo/bar",
			expected: "bar",
		},
		{
			key:      "foo/bar/baz",
			expected: "baz",
		},
		// FIXME is this wanted ?
		{
			key:      "foo/bar/",
			expected: "",
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			actual := p.last(test.key)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderSplitGet(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "courgette, carotte, tomate, aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: []string{"courgette", "carotte", "tomate", "aubergine"},
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			values := p.splitGet(test.keyParts...)

			assert.Equal(t, test.expected, values, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetList(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc: "comma separated",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("entrypoints", "courgette, carotte, tomate, aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: []string{"courgette", "carotte", "tomate", "aubergine"},
		},
		{
			desc: "multiple entries",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("entrypoints/0", "courgette"),
					withPair("entrypoints/1", "carotte"),
					withPair("entrypoints/2", "tomate"),
					withPair("entrypoints/3", "aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: []string{"courgette", "carotte", "tomate", "aubergine"},
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			values := p.getList(test.keyParts...)

			assert.Equal(t, test.expected, values, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetSlice(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected []string
	}{
		{
			desc: "multiple entries",
			kvPairs: filler("traefik",
				frontend("foo",
					withList("entrypoints", "courgette", "carotte", "tomate", "aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: []string{"courgette", "carotte", "tomate", "aubergine"},
		},
		{
			desc: "comma separated",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("entrypoints", "courgette, carotte, tomate, aubergine"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/entrypoints"},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			values := p.getSlice(test.keyParts...)

			assert.Equal(t, test.expected, values, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetBool(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected bool
	}{
		{
			desc: "when value is 'true",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: true,
		},
		{
			desc: "when value is 'false",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "false"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: false,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getBool(false, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetInt(t *testing.T) {
	defaultValue := 666

	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected int
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "6"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: 6,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getInt(defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetInt64(t *testing.T) {
	var defaultValue int64 = 666

	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		kvError  error
		keyParts []string
		expected int64
	}{
		{
			desc: "when has value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "6"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: 6,
		},
		{
			desc: "when empty value",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", ""),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:     "when not existing key",
			kvPairs:  nil,
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
		{
			desc:    "when KV error",
			kvError: store.ErrKeyNotFound,
			kvPairs: filler("traefik",
				frontend("foo",
					withPair("bar", "true"),
				),
			),
			keyParts: []string{"traefik/frontends/foo/bar"},
			expected: defaultValue,
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: newKvClientMock(test.kvPairs, test.kvError),
			}

			actual := p.getInt64(defaultValue, test.keyParts...)

			assert.Equal(t, test.expected, actual, "key: %v", test.keyParts)
		})
	}
}

func TestProviderGetMap(t *testing.T) {
	testCases := []struct {
		desc     string
		keyParts []string
		kvPairs  []*store.KVPair
		expected map[string]string
	}{
		{
			desc:     "when several keys",
			keyParts: []string{"traefik/frontends/foo", pathFrontendCustomRequestHeaders},
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendCustomRequestHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendCustomRequestHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendCustomRequestHeaders+"X-Custom-Header", "test"),
				),
			),
			expected: map[string]string{
				"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
				"Content-Type":                 "application/json; charset=utf-8",
				"X-Custom-Header":              "test",
			},
		},
		{
			desc:     "when no keys",
			keyParts: []string{"traefik/frontends/foo", pathFrontendCustomRequestHeaders},
			kvPairs:  filler("traefik", frontend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getMap(test.keyParts...)

			assert.EqualValues(t, test.expected, result)
		})
	}
}

func TestProviderHasStickinessLabel(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		rootPath string
		expected bool
	}{
		{
			desc:     "without option",
			expected: false,
		},
		{
			desc:     "with cookie name without stickiness=true",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickinessCookieName, "aubergine"),
				),
			),
			expected: false,
		},
		{
			desc:     "stickiness=true",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickiness, "true"),
				),
			),
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				kvClient: &Mock{
					KVPairs: test.kvPairs,
				},
			}

			actual := p.hasStickinessLabel(test.rootPath)

			if actual != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func TestWhiteList(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.WhiteList
	}{
		{
			desc:     "should return nil when no white list labels",
			rootPath: "traefik/frontends/foo",
			expected: nil,
		},
		{
			desc:     "should return a struct when only range",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendWhiteListSourceRange, "10.10.10.10"))),
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: false,
			},
		},
		{
			desc:     "should return a struct when range and UseXForwardedFor",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendWhiteListSourceRange, "10.10.10.10"),
					withPair(pathFrontendWhiteListUseXForwardedFor, "true"))),
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: true,
			},
		},
		{
			desc:     "should return nil when only UseXForwardedFor",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendWhiteListUseXForwardedFor, "true"))),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			actual := p.getWhiteList(test.rootPath)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.Redirect
	}{
		{
			desc:     "should use entry point when entry point key is valued in the store",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRedirectEntryPoint, "https"))),
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc:     "should use entry point when entry point key is valued in the store (permanent)",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRedirectEntryPoint, "https"),
					withPair(pathFrontendRedirectPermanent, "true"))),
			expected: &types.Redirect{
				EntryPoint: "https",
				Permanent:  true,
			},
		},
		{
			desc:     "should use regex when regex keys are valued in the store",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRedirectRegex, "(.*)"),
					withPair(pathFrontendRedirectReplacement, "$1"))),
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
		{
			desc:     "should use regex when regex keys are valued in the store (permanent)",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRedirectRegex, "(.*)"),
					withPair(pathFrontendRedirectReplacement, "$1"),
					withPair(pathFrontendRedirectPermanent, "true"))),
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
				Permanent:   true,
			},
		},
		{
			desc:     "should only use entry point when entry point and regex base are valued in the store",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRedirectEntryPoint, "https"),
					withPair(pathFrontendRedirectRegex, "nope"),
					withPair(pathFrontendRedirectReplacement, "nope"))),
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc:     "should return when redirect keys are not valued in the store",
			rootPath: "traefik/frontends/foo",
			kvPairs:  filler("traefik", frontend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			actual := p.getRedirect(test.rootPath)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetErrorPages(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected map[string]*types.ErrorPage
	}{
		{
			desc:     "2 errors pages",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withErrorPage("foo", "error", "/test1", "500-501", "503-599"),
					withErrorPage("bar", "error", "/test2", "400-405"))),
			expected: map[string]*types.ErrorPage{
				"foo": {
					Backend: "error",
					Query:   "/test1",
					Status:  []string{"500-501", "503-599"},
				},
				"bar": {
					Backend: "error",
					Query:   "/test2",
					Status:  []string{"400-405"},
				},
			},
		},
		{
			desc:     "return nil when no errors pages",
			rootPath: "traefik/frontends/foo",
			kvPairs:  filler("traefik", frontend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			actual := p.getErrorPages(test.rootPath)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetRateLimit(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.RateLimit
	}{
		{
			desc:     "with several limits",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withRateLimit("client.ip",
						withLimit("foo", "6", "12", "18"),
						withLimit("bar", "3", "6", "9")))),
			expected: &types.RateLimit{
				ExtractorFunc: "client.ip",
				RateSet: map[string]*types.Rate{
					"foo": {
						Average: 6,
						Burst:   12,
						Period:  flaeg.Duration(18 * time.Second),
					},
					"bar": {
						Average: 3,
						Burst:   6,
						Period:  flaeg.Duration(9 * time.Second),
					},
				},
			},
		},
		{
			desc:     "return nil when no extractor func",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withRateLimit("",
						withLimit("foo", "6", "12", "18"),
						withLimit("bar", "3", "6", "9")))),
			expected: nil,
		},
		{
			desc:     "return nil when no rate limit keys",
			rootPath: "traefik/frontends/foo",
			kvPairs:  filler("traefik", frontend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			actual := p.getRateLimit(test.rootPath)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetHeaders(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.Headers
	}{
		{
			desc:     "Custom Request Headers",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendCustomRequestHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendCustomRequestHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendCustomRequestHeaders+"X-Custom-Header", "test"))),
			expected: &types.Headers{
				CustomRequestHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
					"X-Custom-Header":              "test",
				},
			},
		},
		{
			desc:     "Custom esponse Headers",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendCustomResponseHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendCustomResponseHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendCustomResponseHeaders+"X-Custom-Header", "test"))),
			expected: &types.Headers{
				CustomResponseHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
					"X-Custom-Header":              "test",
				},
			},
		},
		{
			desc:     "SSL Proxy Headers",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSSLProxyHeaders+"Access-Control-Allow-Methods", "POST,GET,OPTIONS"),
					withPair(pathFrontendSSLProxyHeaders+"Content-Type", "application/json; charset=utf-8"),
					withPair(pathFrontendSSLProxyHeaders+"X-Custom-Header", "test"))),
			expected: &types.Headers{
				SSLProxyHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
					"X-Custom-Header":              "test",
				},
			},
		},
		{
			desc:     "Allowed Hosts",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendAllowedHosts, "foo, bar, goo, hor"))),
			expected: &types.Headers{
				AllowedHosts: []string{"foo", "bar", "goo", "hor"},
			},
		},
		{
			desc:     "Hosts Proxy Headers",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendHostsProxyHeaders, "foo, bar, goo, hor"))),
			expected: &types.Headers{
				HostsProxyHeaders: []string{"foo", "bar", "goo", "hor"},
			},
		},
		{
			desc:     "SSL Redirect",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSSLRedirect, "true"))),
			expected: &types.Headers{
				SSLRedirect: true,
			},
		},
		{
			desc:     "SSL Temporary Redirect",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSSLTemporaryRedirect, "true"))),
			expected: &types.Headers{
				SSLTemporaryRedirect: true,
			},
		},
		{
			desc:     "SSL Host",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSSLHost, "foo"))),
			expected: &types.Headers{
				SSLHost: "foo",
			},
		},
		{
			desc:     "STS Seconds",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSTSSeconds, "666"))),
			expected: &types.Headers{
				STSSeconds: 666,
			},
		},
		{
			desc:     "STS Include Subdomains",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSTSIncludeSubdomains, "true"))),
			expected: &types.Headers{
				STSIncludeSubdomains: true,
			},
		},
		{
			desc:     "STS Preload",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendSTSPreload, "true"))),
			expected: &types.Headers{
				STSPreload: true,
			},
		},
		{
			desc:     "Force STS Header",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendForceSTSHeader, "true"))),
			expected: &types.Headers{
				ForceSTSHeader: true,
			},
		},
		{
			desc:     "Frame Deny",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendFrameDeny, "true"))),
			expected: &types.Headers{
				FrameDeny: true,
			},
		},
		{
			desc:     "Custom Frame Options Value",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendCustomFrameOptionsValue, "foo"))),
			expected: &types.Headers{
				CustomFrameOptionsValue: "foo",
			},
		},
		{
			desc:     "Content Type Nosniff",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendContentTypeNosniff, "true"))),
			expected: &types.Headers{
				ContentTypeNosniff: true,
			},
		},
		{
			desc:     "Browser XSS Filter",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendBrowserXSSFilter, "true"))),
			expected: &types.Headers{
				BrowserXSSFilter: true,
			},
		},
		{
			desc:     "Custom Browser XSS Value",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendCustomBrowserXSSValue, "foo"))),
			expected: &types.Headers{
				CustomBrowserXSSValue: "foo",
			},
		},
		{
			desc:     "Content Security Policy",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendContentSecurityPolicy, "foo"))),
			expected: &types.Headers{
				ContentSecurityPolicy: "foo",
			},
		},
		{
			desc:     "Public Key",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendPublicKey, "foo"))),
			expected: &types.Headers{
				PublicKey: "foo",
			},
		},
		{
			desc:     "Referrer Policy",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendReferrerPolicy, "foo"))),
			expected: &types.Headers{
				ReferrerPolicy: "foo",
			},
		},
		{
			desc:     "Is Development",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendIsDevelopment, "true"))),
			expected: &types.Headers{
				IsDevelopment: true,
			},
		},
		{
			desc:     "should return nil when not significant configuration",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendIsDevelopment, "false"))),
			expected: nil,
		},
		{
			desc:     "should return nil when no headers configuration",
			rootPath: "traefik/frontends/foo",
			kvPairs:  filler("traefik", frontend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			headers := p.getHeaders(test.rootPath)

			assert.Equal(t, test.expected, headers)
		})
	}
}

func TestProviderGetLoadBalancer(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.LoadBalancer
	}{
		{
			desc:     "when all keys",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerMethod, "drr"),
					withPair(pathBackendLoadBalancerSticky, "true"),
					withPair(pathBackendLoadBalancerStickiness, "true"),
					withPair(pathBackendLoadBalancerStickinessCookieName, "aubergine"))),
			expected: &types.LoadBalancer{
				Method: "drr",
				Sticky: true,
				Stickiness: &types.Stickiness{
					CookieName: "aubergine",
				},
			},
		},
		{
			desc:     "when no specific configuration",
			rootPath: "traefik/backends/foo",
			kvPairs:  filler("traefik", backend("foo")),
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
		{
			desc:     "when method is set",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerMethod, "drr"))),
			expected: &types.LoadBalancer{
				Method: "drr",
			},
		},
		{
			desc:     "when sticky is set",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerSticky, "true"))),
			expected: &types.LoadBalancer{
				Method: "wrr",
				Sticky: true,
			},
		},
		{
			desc:     "when stickiness is set",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickiness, "true"))),
			expected: &types.LoadBalancer{
				Method:     "wrr",
				Stickiness: &types.Stickiness{},
			},
		},
		{
			desc:     "when stickiness cookie name is set",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickiness, "true"),
					withPair(pathBackendLoadBalancerStickinessCookieName, "aubergine"))),
			expected: &types.LoadBalancer{
				Method: "wrr",
				Stickiness: &types.Stickiness{
					CookieName: "aubergine",
				},
			},
		},
		{
			desc:     "when stickiness cookie name is set but not stickiness",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendLoadBalancerStickinessCookieName, "aubergine"))),
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getLoadBalancer(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetCircuitBreaker(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.CircuitBreaker
	}{
		{
			desc:     "when cb expression defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression))),
			expected: &types.CircuitBreaker{
				Expression: label.DefaultCircuitBreakerExpression,
			},
		},
		{
			desc:     "when no cb expression",
			rootPath: "traefik/backends/foo",
			kvPairs:  filler("traefik", backend("foo")),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getCircuitBreaker(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetMaxConn(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.MaxConn
	}{
		{
			desc:     "when max conn keys are defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendMaxConnAmount, "5"),
					withPair(pathBackendMaxConnExtractorFunc, "client.ip"))),
			expected: &types.MaxConn{
				Amount:        5,
				ExtractorFunc: "client.ip",
			},
		},
		{
			desc:     "should return nil when only extractor func is defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendMaxConnExtractorFunc, "client.ip"))),
			expected: nil,
		},
		{
			desc:     "when only amount is defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendMaxConnAmount, "5"))),
			expected: &types.MaxConn{
				Amount:        5,
				ExtractorFunc: "request.host",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getMaxConn(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetHealthCheck(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.HealthCheck
	}{
		{
			desc:     "when all configuration keys defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendHealthCheckPath, "/health"),
					withPair(pathBackendHealthCheckPort, "80"),
					withPair(pathBackendHealthCheckInterval, "10s"))),
			expected: &types.HealthCheck{
				Interval: "10s",
				Path:     "/health",
				Port:     80,
			},
		},
		{
			desc:     "when only path defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendHealthCheckPath, "/health"))),
			expected: &types.HealthCheck{
				Interval: "30s",
				Path:     "/health",
				Port:     0,
			},
		},
		{
			desc:     "should return nil when no path",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendHealthCheckPort, "80"),
					withPair(pathBackendHealthCheckInterval, "30s"))),
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getHealthCheck(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetBufferingReal(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.Buffering
	}{
		{
			desc:     "when all configuration keys defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendBufferingMaxResponseBodyBytes, "10485760"),
					withPair(pathBackendBufferingMemResponseBodyBytes, "2097152"),
					withPair(pathBackendBufferingMaxRequestBodyBytes, "10485760"),
					withPair(pathBackendBufferingMemRequestBodyBytes, "2097152"),
					withPair(pathBackendBufferingRetryExpression, "IsNetworkError() && Attempts() <= 2"))),
			expected: &types.Buffering{
				MaxResponseBodyBytes: 10485760,
				MemResponseBodyBytes: 2097152,
				MaxRequestBodyBytes:  10485760,
				MemRequestBodyBytes:  2097152,
				RetryExpression:      "IsNetworkError() && Attempts() <= 2",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getBuffering(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetTLSes(t *testing.T) {
	testCases := []struct {
		desc     string
		kvPairs  []*store.KVPair
		expected []*tls.Configuration
	}{
		{
			desc: "when several TLS configuration defined",
			kvPairs: filler("traefik",
				entry("tls/foo",
					withPair("entrypoints", "http,https"),
					withPair("certificate/certfile", "certfile1"),
					withPair("certificate/keyfile", "keyfile1")),
				entry("tls/bar",
					withPair("entrypoints", "http,https"),
					withPair("certificate/certfile", "certfile2"),
					withPair("certificate/keyfile", "keyfile2"))),
			expected: []*tls.Configuration{
				{
					EntryPoints: []string{"http", "https"},
					Certificate: &tls.Certificate{
						CertFile: "certfile2",
						KeyFile:  "keyfile2",
					},
				},
				{
					EntryPoints: []string{"http", "https"},
					Certificate: &tls.Certificate{
						CertFile: "certfile1",
						KeyFile:  "keyfile1",
					},
				},
			},
		},
		{
			desc:     "should return nil when no TLS configuration",
			kvPairs:  filler("traefik", entry("tls/foo")),
			expected: nil,
		},
		{
			desc: "should return nil when no entry points",
			kvPairs: filler("traefik",
				entry("tls/foo",
					withPair("certificate/certfile", "certfile2"),
					withPair("certificate/keyfile", "keyfile2"))),
			expected: nil,
		},
		{
			desc: "should return nil when no cert file and no key file",
			kvPairs: filler("traefik",
				entry("tls/foo",
					withPair("entrypoints", "http,https"))),
			expected: nil,
		},
	}
	prefix := "traefik"

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getTLSSection(prefix)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetAuth(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected *types.Auth
	}{
		{
			desc:     "should return nil when no data",
			expected: nil,
		},
		{
			desc:     "should return a valid basic auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendAuthBasicRemoveHeader, "true"),
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthBasicUsersFile, ".htpasswd"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"))),
			expected: &types.Auth{
				HeaderField: "X-WebAuth-User",
				Basic: &types.Basic{
					RemoveHeader: true,
					Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					UsersFile: ".htpasswd",
				},
			},
		},
		{
			desc:     "should return a valid basic auth (backward compatibility)",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendBasicAuth, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				)),
			expected: &types.Auth{
				Basic: &types.Basic{
					Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
				},
			},
		},
		{
			desc:     "should return a valid digest auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withList(pathFrontendAuthDigestUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withPair(pathFrontendAuthDigestUsersFile, ".htpasswd"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
				)),
			expected: &types.Auth{
				HeaderField: "X-WebAuth-User",
				Digest: &types.Digest{
					Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					UsersFile: ".htpasswd",
				},
			},
		},
		{
			desc:     "should return a valid forward auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendAuthForwardAddress, "auth.server"),
					withPair(pathFrontendAuthForwardTrustForwardHeader, "true"),
					withPair(pathFrontendAuthForwardTLSCa, "ca.crt"),
					withPair(pathFrontendAuthForwardTLSCaOptional, "true"),
					withPair(pathFrontendAuthForwardTLSCert, "server.crt"),
					withPair(pathFrontendAuthForwardTLSKey, "server.key"),
					withPair(pathFrontendAuthForwardTLSInsecureSkipVerify, "true"),
					withPair(pathFrontendAuthHeaderField, "X-WebAuth-User"),
				)),
			expected: &types.Auth{
				HeaderField: "X-WebAuth-User",
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
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getAuth(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderHasDeprecatedBasicAuth(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected bool
	}{
		{
			desc:     "should return nil when no data",
			expected: false,
		},
		{
			desc:     "should return a valid basic auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				)),
			expected: false,
		},
		{
			desc:     "should return a valid basic auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withList(pathFrontendBasicAuth, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				)),
			expected: true,
		},
		{
			desc:     "should return a valid basic auth",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withList(pathFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					withList(pathFrontendBasicAuth, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				)),
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.hasDeprecatedBasicAuth(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetRoutes(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected map[string]types.Route
	}{
		{
			desc:     "should return nil when no data",
			expected: nil,
		},
		{
			desc:     "should return nil when route key exists but without rule key",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRoutes+"bar", "test1"),
					withPair(pathFrontendRoutes+"bir", "test2"))),
			expected: nil,
		},
		{
			desc:     "should return a map when configuration keys are defined",
			rootPath: "traefik/frontends/foo",
			kvPairs: filler("traefik",
				frontend("foo",
					withPair(pathFrontendRoutes+"bar"+pathFrontendRule, "test1"),
					withPair(pathFrontendRoutes+"bir"+pathFrontendRule, "test2"))),
			expected: map[string]types.Route{
				"bar": {
					Rule: "test1",
				},
				"bir": {
					Rule: "test2",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getRoutes(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetServers(t *testing.T) {
	testCases := []struct {
		desc     string
		rootPath string
		kvPairs  []*store.KVPair
		expected map[string]types.Server
	}{
		{
			desc:     "should return nil when no data",
			expected: nil,
		},
		{
			desc:     "should return nil when server has no URL",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendServers+"server1/weight", "7"),
					withPair(pathBackendServers+"server2/weight", "6"))),
			expected: nil,
		},
		{
			desc:     "should use default weight when invalid weight value",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendServers+"server1/url", "http://172.17.0.2:80"),
					withPair(pathBackendServers+"server1/weight", "kls"))),
			expected: map[string]types.Server{
				"server1": {
					URL:    "http://172.17.0.2:80",
					Weight: label.DefaultWeight,
				},
			},
		},
		{
			desc:     "should return a map when configuration keys are defined",
			rootPath: "traefik/backends/foo",
			kvPairs: filler("traefik",
				backend("foo",
					withPair(pathBackendServers+"server1/url", "http://172.17.0.2:80"),
					withPair(pathBackendServers+"server2/url", "http://172.17.0.3:80"),
					withPair(pathBackendServers+"server2/weight", "6"))),
			expected: map[string]types.Server{
				"server1": {
					URL:    "http://172.17.0.2:80",
					Weight: label.DefaultWeight,
				},
				"server2": {
					URL:    "http://172.17.0.3:80",
					Weight: 6,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := newProviderMock(test.kvPairs)

			result := p.getServers(test.rootPath)

			assert.Equal(t, test.expected, result)
		})
	}
}
