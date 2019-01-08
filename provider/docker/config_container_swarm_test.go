package docker

import (
	"strconv"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwarmBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		services          []swarm.Service
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
		networks          map[string]*docker.NetworkResource
	}{
		{
			desc:              "when no container",
			services:          []swarm.Service{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
			networks:          map[string]*docker.NetworkResource{},
		},
		{
			desc: "when basic container configuration",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort: "80",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when container has label 'enable' to false",
			services: []swarm.Service{
				swarmService(
					serviceName("test1"),
					serviceLabels(map[string]string{
						label.TraefikEnable:   "false",
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",
						label.TraefikBackend:  "foobar",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when pass tls client cert configuration",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort:                                                 "80",
						label.TraefikFrontendPassTLSClientCertPem:                         "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                   "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
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
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when frontend basic auth configuration",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort:                          "80",
						label.TraefikFrontendAuthBasicUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile:    ".htpasswd",
						label.TraefikFrontendAuthBasicRemoveHeader: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Basic: &types.Basic{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when frontend basic auth configuration backward compatibility",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort:              "80",
						label.TraefikFrontendAuthBasic: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
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
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when frontend digest auth configuration",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort:                           "80",
						label.TraefikFrontendAuthDigestUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:    ".htpasswd",
						label.TraefikFrontendAuthDigestRemoveHeader: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
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
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when frontend forward auth configuration",
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort:                                     "80",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.TraefikFrontendAuthForwardAuthResponseHeaders:   "X-Auth-User,X-Auth-Token",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Auth: &types.Auth{
						Forward: &types.Forward{
							Address: "auth.server",
							TLS: &types.ClientTLS{
								CA:                 "ca.crt",
								CAOptional:         true,
								Cert:               "server.crt",
								Key:                "server.key",
								InsecureSkipVerify: true,
							},
							TrustForwardHeader:  true,
							AuthResponseHeaders: []string{"X-Auth-User", "X-Auth-Token"},
						},
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
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
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			desc: "when all labels are set",
			services: []swarm.Service{
				swarmService(
					serviceName("test1"),
					serviceLabels(map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendResponseForwardingFlushInterval:  "10ms",
						label.TraefikBackendHealthCheckScheme:                "http",
						label.TraefikBackendHealthCheckPath:                  "/health",
						label.TraefikBackendHealthCheckPort:                  "880",
						label.TraefikBackendHealthCheckInterval:              "6",
						label.TraefikBackendHealthCheckHostname:              "foo.com",
						label.TraefikBackendHealthCheckHeaders:               "Foo:bar || Bar:foo",
						label.TraefikBackendLoadBalancerMethod:               "drr",
						label.TraefikBackendLoadBalancerSticky:               "true",
						label.TraefikBackendLoadBalancerStickiness:           "true",
						label.TraefikBackendLoadBalancerStickinessCookieName: "chocolate",
						label.TraefikBackendMaxConnAmount:                    "666",
						label.TraefikBackendMaxConnExtractorFunc:             "client.ip",
						label.TraefikBackendBufferingMaxResponseBodyBytes:    "10485760",
						label.TraefikBackendBufferingMemResponseBodyBytes:    "2097152",
						label.TraefikBackendBufferingMaxRequestBodyBytes:     "10485760",
						label.TraefikBackendBufferingMemRequestBodyBytes:     "2097152",
						label.TraefikBackendBufferingRetryExpression:         "IsNetworkError() && Attempts() <= 2",

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

						label.TraefikFrontendAuthBasic:                 "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendEntryPoints:               "http,https",
						label.TraefikFrontendPassHostHeader:            "true",
						label.TraefikFrontendPassTLSCert:               "true",
						label.TraefikFrontendPriority:                  "666",
						label.TraefikFrontendRedirectEntryPoint:        "https",
						label.TraefikFrontendRedirectRegex:             "nope",
						label.TraefikFrontendRedirectReplacement:       "nope",
						label.TraefikFrontendRule:                      "Host:traefik.io",
						label.TraefikFrontendWhiteListSourceRange:      "10.10.10.10",
						label.TraefikFrontendWhiteListUseXForwardedFor: "true",

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
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
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
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
						Basic: &types.Basic{
							RemoveHeader: true,
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							UsersFile: ".htpasswd",
						},
					},
					WhiteList: &types.WhiteList{
						SourceRange:      []string{"10.10.10.10"},
						UseXForwardedFor: true,
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
						SSLForceHost:         true,
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
					ResponseForwarding: &types.ResponseForwarding{
						FlushInterval: "10ms",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
						Sticky: true,
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
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, service := range test.services {
				dData := parseService(service, test.networks)
				dockerDataList = append(dockerDataList, dData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				SwarmMode:        true,
			}

			actualConfig := provider.buildConfigurationV2(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestSwarmTraefikFilter(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected bool
		networks map[string]*docker.NetworkResource
		provider *Provider
	}{
		{
			service:  swarmService(),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "false",
				label.TraefikPort:   "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
				label.TraefikPort:         "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikPort: "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "true",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "anything",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
				label.TraefikPort:         "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikPort: "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "true",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)

			actual := test.provider.containerFilter(dData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestSwarmGetFrontendName(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host-foo-docker-localhost-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				serviceName("test"),
				serviceLabels(map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				}),
			),
			expected: "PathPrefix-test2-0",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}

			actual := provider.getFrontendName(dData, 0)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetFrontendRule(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host:foo.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceName("foo"),
				serviceLabels(map[string]string{
					label.TraefikDomain: "traefik.localhost",
				})),
			expected: "Host:foo.traefik.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)

			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}

			actual := provider.getFrontendRule(dData, segmentProperties[""])
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetBackendName(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "foo",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service:  swarmService(serviceName("bar")),
			expected: "bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikBackend: "foobar",
			})),
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			actual := getBackendName(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(withEndpointSpec(modeDNSSR)),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				withEndpointSpec(modeVIP),
				withEndpoint(virtualIP("1", "10.11.12.13/24")),
			),
			expected: "10.11.12.13",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceLabels(map[string]string{
					labelDockerNetwork: "barnet",
				}),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", "10.11.12.13/24"),
					virtualIP("2", "10.11.12.99/24"),
				),
			),
			expected: "10.11.12.99",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				SwarmMode: true,
			}

			dData := parseService(test.service, test.networks)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			actual := provider.getDeprecatedIPAddress(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSwarmGetPort(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(
				serviceLabels(map[string]string{
					label.TraefikPort: "8080",
				}),
				withEndpointSpec(modeDNSSR),
			),
			expected: "8080",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			actual := getPort(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}
