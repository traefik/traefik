package docker

import (
	"strconv"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerBuildConfiguration(t *testing.T) {
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
			desc: "when basic container configuration",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when pass tls client certificate",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikFrontendPassTLSClientCertPem:                         "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                   "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
					CircuitBreaker: nil,
				},
			},
		}, {
			desc: "when frontend basic auth",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikFrontendAuthBasicUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthBasicUsersFile:    ".htpasswd",
						label.TraefikFrontendAuthBasicRemoveHeader: "true",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend basic auth backward compatibility",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikFrontendAuthBasic: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend digest auth",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikFrontendAuthDigestRemoveHeader: "true",
						label.TraefikFrontendAuthDigestUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendAuthDigestUsersFile:    ".htpasswd",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when frontend forward auth",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikFrontendAuthForwardTrustForwardHeader:    "true",
						label.TraefikFrontendAuthForwardAddress:               "auth.server",
						label.TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
						label.TraefikFrontendAuthForwardTLSCaOptional:         "true",
						label.TraefikFrontendAuthForwardTLSCert:               "server.crt",
						label.TraefikFrontendAuthForwardTLSKey:                "server.key",
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.TraefikFrontendAuthForwardAuthResponseHeaders:   "X-Auth-User,X-Auth-Token",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
								InsecureSkipVerify: true,
								Cert:               "server.crt",
								Key:                "server.key",
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
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when basic container configuration with multiple network",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
					withNetwork("webnet", ipv4("127.0.0.2")),
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
						"server-test-48093b9fc43454203aacd2bc4057a08c": {
							URL:    "http://127.0.0.2:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when basic container configuration with specific network",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						"traefik.docker.network": "mywebnet",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
					withNetwork("webnet", ipv4("127.0.0.2")),
					withNetwork("mywebnet", ipv4("127.0.0.3")),
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
						"server-test-405767e9733427148cd8dae6c4d331b0": {
							URL:    "http://127.0.0.3:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when container has label 'enable' to false",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikEnable:   "false",
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",
						label.TraefikBackend:  "foobar",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
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

						label.TraefikFrontendPassTLSClientCertPem:                         "true",
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.TraefikFrontendPassTLSClientCertInfosSans:                   "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",

						label.TraefikFrontendAuthBasic:                        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
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

						label.TraefikFrontendEntryPoints:               "http,https",
						label.TraefikFrontendPassHostHeader:            "true",
						label.TraefikFrontendPassTLSCert:               "true",
						label.TraefikFrontendPriority:                  "666",
						label.TraefikFrontendRedirectEntryPoint:        "https",
						label.TraefikFrontendRedirectRegex:             "nope",
						label.TraefikFrontendRedirectReplacement:       "nope",
						label.TraefikFrontendRedirectPermanent:         "true",
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
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
		},
		{
			desc: "when docker compose scale with different compose service names",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test_0"),
					labels(map[string]string{
						labelDockerComposeProject: "myProject",
						labelDockerComposeService: "myService",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test_1"),
					labels(map[string]string{
						labelDockerComposeProject: "myProject",
						labelDockerComposeService: "myService",
					}),

					ports(nat.PortMap{
						"80/tcp": {},
					}),

					withNetwork("bridge", ipv4("127.0.0.2")),
				),
				containerJSON(
					name("test_2"),
					labels(map[string]string{
						labelDockerComposeProject: "myProject",
						labelDockerComposeService: "myService2",
					}),

					ports(nat.PortMap{
						"80/tcp": {},
					}),

					withNetwork("bridge", ipv4("127.0.0.3")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-myService-myProject-docker-localhost-0": {
					Backend:        "backend-myService-myProject",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-myService-myProject-docker-localhost-0": {
							Rule: "Host:myService.myProject.docker.localhost",
						},
					},
				},
				"frontend-Host-myService2-myProject-docker-localhost-2": {
					Backend:        "backend-myService2-myProject",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-myService2-myProject-docker-localhost-2": {
							Rule: "Host:myService2.myProject.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-myService-myProject": {
					Servers: map[string]types.Server{
						"server-test-0-842895ca2aca17f6ee36ddb2f621194d": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
						"server-test-1-48093b9fc43454203aacd2bc4057a08c": {
							URL:    "http://127.0.0.2:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-myService2-myProject": {
					Servers: map[string]types.Server{
						"server-test-2-405767e9733427148cd8dae6c4d331b0": {
							URL:    "http://127.0.0.3:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dockerDataList []dockerData
			for _, cont := range test.containers {
				dData := parseContainer(cont)
				dockerDataList = append(dockerDataList, dData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				Network:          "webnet",
			}
			actualConfig := provider.buildConfigurationV2(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestDockerTraefikFilter(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  bool
		provider  *Provider
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: false,
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable: "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikFrontendRule: "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container-multi-ports",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikPort: "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable: "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikFrontendRule: "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						label.TraefikEnable:       "true",
						label.TraefikFrontendRule: "Host:i.love.this.host",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: true,
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			actual := test.provider.containerFilter(dData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestDockerGetFrontendName(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host-foo-docker-localhost-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "Host-bar-foo-docker-localhost-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "PathPrefix:/test2",
			})),
			expected: "PathPrefix-test2-0",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			provider := &Provider{
				Domain: "docker.localhost",
			}

			actual := provider.getFrontendName(dData, 0)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host:foo.docker.localhost",
		},
		{
			container: containerJSON(name("foo"),
				labels(map[string]string{
					label.TraefikDomain: "traefik.localhost",
				})),
			expected: "Host:foo.traefik.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "Host:bar.foo.docker.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)

			provider := &Provider{
				Domain: "docker.localhost",
			}

			actual := provider.getFrontendRule(dData, segmentProperties[""])
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetBackendName(t *testing.T) {
	testCases := []struct {
		container   docker.ContainerJSON
		segmentName string
		expected    string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "foo",
		},
		{
			container: containerJSON(name("bar")),
			expected:  "bar",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikBackend: "foobar",
			})),
			expected: "foobar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "bar-foo",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
				"traefik.sauternes.backend":  "titi",
			})),
			segmentName: "sauternes",
			expected:    "bar-foo-titi",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[test.segmentName]
			dData.SegmentName = test.segmentName

			actual := getBackendName(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetIPAddress(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			expected:  "10.11.12.13",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					labelDockerNetwork: "testnet",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
			),
			expected: "10.11.12.13",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					labelDockerNetwork: "testnet2",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.14",
		},
		{
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "127.0.0.1",
		},
		{
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.14",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					labelDockerNetwork: "testnet",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.13",
		},
		{
			container: containerJSON(
				networkMode("host"),
			),
			expected: "127.0.0.1",
		},
		{
			container: containerJSON(
				networkMode("host"),
				nodeIP("10.0.0.5"),
			),
			expected: "10.0.0.5",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			provider := &Provider{
				Network: "webnet",
			}

			actual := provider.getDeprecatedIPAddress(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetIPPort(t *testing.T) {
	testCases := []struct {
		desc         string
		container    docker.ContainerJSON
		ip, port     string
		expectsError bool
	}{
		{
			desc: "label traefik.port not set, no binding, falling back on the container's IP/Port",
			container: containerJSON(
				ports(nat.PortMap{
					"8080/tcp": {},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "10.11.12.13",
			port: "8080",
		},
		{
			desc: "label traefik.port not set, single binding with port only, falling back on the container's IP/Port",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostPort: "8082",
						},
					},
				}),
			),
			ip:   "10.11.12.13",
			port: "80",
		},
		{
			desc: "label traefik.port not set, binding with ip:port should create a route to the bound ip:port",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "1.2.3.4",
			port: "8081",
		},
		{
			desc: "label traefik.port set, no binding, falling back on the container's IP/traefik.port",
			container: containerJSON(
				labels(map[string]string{
					label.TraefikPort: "80",
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "10.11.12.13",
			port: "80",
		},
		{
			desc: "label traefik.port set, single binding with ip:port for the label, creates the route",
			container: containerJSON(
				labels(map[string]string{
					label.TraefikPort: "443",
				}),
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "5.6.7.8",
			port: "8082",
		},
		{
			desc: "label traefik.port set, no binding on the corresponding port, falling back on the container's IP/label.port",
			container: containerJSON(
				labels(map[string]string{
					label.TraefikPort: "80",
				}),
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "10.11.12.13",
			port: "80",
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (first) binding",
			container: containerJSON(
				labels(map[string]string{
					label.TraefikPort: "80",
				}),
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "1.2.3.4",
			port: "8081",
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (second) binding",
			container: containerJSON(
				labels(map[string]string{
					label.TraefikPort: "443",
				}),
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			ip:   "5.6.7.8",
			port: "8082",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			provider := &Provider{
				Network:       "testnet",
				UseBindPortIP: true,
			}

			actualIP, actualPort, actualError := provider.getIPPort(dData)
			if test.expectsError {
				require.Error(t, actualError)
			} else {
				require.NoError(t, actualError)
			}
			assert.Equal(t, test.ip, actualIP)
			assert.Equal(t, test.port, actualPort)
		})
	}
}

func TestDockerGetPort(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			container: containerJSON(ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "80",
		},
		{
			container: containerJSON(ports(nat.PortMap{
				"80/tcp":  {},
				"443/tcp": {},
			})),
			expected: "80",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "8080",
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "8080",
			}), ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "8080",
			}), ports(nat.PortMap{
				"8080/tcp": {},
				"80/tcp":   {},
			})),
			expected: "8080",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			segmentProperties := label.ExtractTraefikLabels(dData.Labels)
			dData.SegmentLabels = segmentProperties[""]

			actual := getPort(dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServers(t *testing.T) {
	p := &Provider{}

	testCases := []struct {
		desc       string
		containers []docker.ContainerJSON
		expected   map[string]types.Server
	}{
		{
			desc:     "no container",
			expected: nil,
		},
		{
			desc: "with a simple container",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					withNetwork("testnet", ipv4("10.10.10.10")),
					ports(nat.PortMap{
						"80/tcp": {},
					})),
			},
			expected: map[string]types.Server{
				"server-test1-fb00f762970935200c76ccdaf91458f6": {
					URL:    "http://10.10.10.10:80",
					Weight: 1,
				},
			},
		},
		{
			desc: "with several containers",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					withNetwork("testnet", ipv4("10.10.10.11")),
					ports(nat.PortMap{
						"80/tcp": {},
					})),
				containerJSON(
					name("test2"),
					withNetwork("testnet", ipv4("10.10.10.12")),
					ports(nat.PortMap{
						"81/tcp": {},
					})),
				containerJSON(
					name("test3"),
					withNetwork("testnet", ipv4("10.10.10.13")),
					ports(nat.PortMap{
						"82/tcp": {},
					})),
			},
			expected: map[string]types.Server{
				"server-test1-743440b6f4a8ffd8737626215f2c5a33": {
					URL:    "http://10.10.10.11:80",
					Weight: 1,
				},
				"server-test2-547f74bbb5da02b6c8141ce9aa96c13b": {
					URL:    "http://10.10.10.12:81",
					Weight: 1,
				},
				"server-test3-c57fd8b848c814a3f2a4a4c12e13c179": {
					URL:    "http://10.10.10.13:82",
					Weight: 1,
				},
			},
		},
		{
			desc: "ignore one container because no ip address",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					withNetwork("testnet", ipv4("")),
					ports(nat.PortMap{
						"80/tcp": {},
					})),
				containerJSON(
					name("test2"),
					withNetwork("testnet", ipv4("10.10.10.12")),
					ports(nat.PortMap{
						"81/tcp": {},
					})),
				containerJSON(
					name("test3"),
					withNetwork("testnet", ipv4("10.10.10.13")),
					ports(nat.PortMap{
						"82/tcp": {},
					})),
			},
			expected: map[string]types.Server{
				"server-test2-547f74bbb5da02b6c8141ce9aa96c13b": {
					URL:    "http://10.10.10.12:81",
					Weight: 1,
				},
				"server-test3-c57fd8b848c814a3f2a4a4c12e13c179": {
					URL:    "http://10.10.10.13:82",
					Weight: 1,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var dockerDataList []dockerData
			for _, cont := range test.containers {
				dData := parseContainer(cont)
				dockerDataList = append(dockerDataList, dData)
			}

			servers := p.getServers(dockerDataList)

			assert.Equal(t, test.expected, servers)
		})
	}
}
