package ecs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSegmentBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		instanceInfo      []ecsInstance
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "when no container",
			instanceInfo:      []ecsInstance{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "simple configuration",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                 "2503",
						"traefik.sauternes.frontend.entryPoints": "http,https",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend:        "backend-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-863563a2e23c95502862016417ee95ea": {
							URL:    "http://127.0.0.1:2503",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "auth basic",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                                                "2503",
						"traefik.sauternes.frontend.entryPoints":                                "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthHeaderField:       "X-WebAuth-User",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsersFile:    ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicRemoveHeader: "true",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend:        "backend-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
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
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-863563a2e23c95502862016417ee95ea": {
							URL:    "http://127.0.0.1:2503",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "auth basic backward compatibility",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                                    "2503",
						"traefik.sauternes.frontend.entryPoints":                    "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasic: "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend:        "backend-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
						},
					},
					Auth: &types.Auth{
						Basic: &types.Basic{
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-863563a2e23c95502862016417ee95ea": {
							URL:    "http://127.0.0.1:2503",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "auth digest",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                                                 "2503",
						"traefik.sauternes.frontend.entryPoints":                                 "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthHeaderField:        "X-WebAuth-User",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsers:        "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsersFile:    ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestRemoveHeader: "true",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend:        "backend-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
						},
					},
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
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-863563a2e23c95502862016417ee95ea": {
							URL:    "http://127.0.0.1:2503",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "auth forward",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						"traefik.sauternes.port":                                                           "2503",
						"traefik.sauternes.frontend.entryPoints":                                           "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthHeaderField:                  "X-WebAuth-User",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardAddress:               "auth.server",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTrustForwardHeader:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCa:                 "ca.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCaOptional:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCert:               "server.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSKey:                "server.key",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardAuthResponseHeaders:   "X-Auth-User,X-Auth-Token",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend:        "backend-foo-sauternes",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
						},
					},
					Auth: &types.Auth{
						HeaderField: "X-WebAuth-User",
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
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-863563a2e23c95502862016417ee95ea": {
							URL:    "http://127.0.0.1:2503",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when all labels are set",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("foo"),
					labels(map[string]string{
						label.Prefix + "sauternes." + label.SuffixPort:     "666",
						label.Prefix + "sauternes." + label.SuffixProtocol: "https",
						label.Prefix + "sauternes." + label.SuffixWeight:   "12",

						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicRemoveHeader:            "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsers:                   "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasicUsersFile:               ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestRemoveHeader:           "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsers:                  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthDigestUsersFile:              ".htpasswd",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardAddress:               "auth.server",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTrustForwardHeader:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCa:                 "ca.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCaOptional:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSCert:               "server.crt",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSKey:                "server.key",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthForwardTLSInsecureSkipVerify: "true",
						label.Prefix + "sauternes." + label.SuffixFrontendAuthHeaderField:                  "X-WebAuth-User",

						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertPem:                         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosNotBefore:              "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosNotAfter:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSans:                   "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerCountry:          "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerLocality:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerProvince:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectCountry:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectLocality:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectProvince:        "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",

						label.Prefix + "sauternes." + label.SuffixFrontendAuthBasic:                 "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "sauternes." + label.SuffixFrontendEntryPoints:               "http,https",
						label.Prefix + "sauternes." + label.SuffixFrontendPassHostHeader:            "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPassTLSCert:               "true",
						label.Prefix + "sauternes." + label.SuffixFrontendPriority:                  "666",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectEntryPoint:        "https",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectRegex:             "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectReplacement:       "nope",
						label.Prefix + "sauternes." + label.SuffixFrontendRedirectPermanent:         "true",
						label.Prefix + "sauternes." + label.SuffixFrontendWhiteListSourceRange:      "10.10.10.10",
						label.Prefix + "sauternes." + label.SuffixFrontendWhiteListUseXForwardedFor: "true",

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
						label.Prefix + "sauternes." + label.SuffixFrontendHeadersSSLForceHost:            "true",
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
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 666),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-foo-sauternes": {
					Backend: "backend-foo-sauternes",
					EntryPoints: []string{
						"http",
						"https",
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

					Routes: map[string]types.Route{
						"route-frontend-sauternes-foo-sauternes": {
							Rule: "Host:foo.ecs.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-sauternes": {
					Servers: map[string]types.Server{
						"server-foo-123456789abc-7f6444e0dff3330c8b0ad2bbbd383b0f": {
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
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("test1"),
					labels(map[string]string{
						"traefik.sauternes.port":                         "2503",
						"traefik.sauternes.protocol":                     "https",
						"traefik.sauternes.weight":                       "80",
						"traefik.sauternes.backend":                      "foobar",
						"traefik.sauternes.frontend.passHostHeader":      "false",
						"traefik.sauternes.frontend.rule":                "Path:/mypath",
						"traefik.sauternes.frontend.priority":            "5000",
						"traefik.sauternes.frontend.entryPoints":         "http,https,ws",
						"traefik.sauternes.frontend.auth.basic":          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.sauternes.frontend.redirect.entryPoint": "https",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
				instance(
					ID("abc987654321"),
					name("test2"),
					labels(map[string]string{
						"traefik.anothersauternes.port":          "8079",
						"traefik.anothersauternes.weight":        "33",
						"traefik.anothersauternes.frontend.rule": "Path:/anotherpath",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(80, 8079),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: false,
					Priority:       5000,
					EntryPoints:    []string{"http", "https", "ws"},
					Auth: &types.Auth{
						Basic: &types.Basic{
							Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
						},
					},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-test1-foobar": {
							Rule: "Path:/mypath",
						},
					},
				},
				"frontend-anothersauternes-test2-anothersauternes": {
					Backend:        "backend-test2-anothersauternes",
					PassHostHeader: true,
					EntryPoints:    []string{},
					Routes: map[string]types.Route{
						"route-frontend-anothersauternes-test2-anothersauternes": {
							Rule: "Path:/anotherpath",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"server-test1-123456789abc-79533a101142718f0fdf84c42593c41e": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-test2-anothersauternes": {
					Servers: map[string]types.Server{
						"server-test2-abc987654321-045e3e4aa5a744a325c099b803700a93": {
							URL:    "http://127.0.0.2:8079",
							Weight: 33,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "several segments with the same backend name and same port",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("test1"),
					labels(map[string]string{
						"traefik.port":                         "2503",
						"traefik.protocol":                     "https",
						"traefik.weight":                       "80",
						"traefik.frontend.entryPoints":         "http,https",
						"traefik.frontend.redirect.entryPoint": "https",

						"traefik.sauternes.backend":           "foobar",
						"traefik.sauternes.frontend.rule":     "Path:/sauternes",
						"traefik.sauternes.frontend.priority": "5000",

						"traefik.arbois.backend":           "foobar",
						"traefik.arbois.frontend.rule":     "Path:/arbois",
						"traefik.arbois.frontend.priority": "3000",
					}),

					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       5000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-test1-foobar": {
							Rule: "Path:/sauternes",
						},
					},
				},
				"frontend-arbois-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       3000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-arbois-test1-foobar": {
							Rule: "Path:/arbois",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"server-test1-123456789abc-79533a101142718f0fdf84c42593c41e": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "several segments with the same backend name and different port (wrong behavior)",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("test1"),
					labels(map[string]string{
						"traefik.protocol":                     "https",
						"traefik.frontend.entryPoints":         "http,https",
						"traefik.frontend.redirect.entryPoint": "https",

						"traefik.sauternes.port":              "2503",
						"traefik.sauternes.weight":            "80",
						"traefik.sauternes.backend":           "foobar",
						"traefik.sauternes.frontend.rule":     "Path:/sauternes",
						"traefik.sauternes.frontend.priority": "5000",

						"traefik.arbois.port":              "2504",
						"traefik.arbois.weight":            "90",
						"traefik.arbois.backend":           "foobar",
						"traefik.arbois.frontend.rule":     "Path:/arbois",
						"traefik.arbois.frontend.priority": "3000",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
							mPort(80, 2504),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       5000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-test1-foobar": {
							Rule: "Path:/sauternes",
						},
					},
				},
				"frontend-arbois-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       3000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-arbois-test1-foobar": {
							Rule: "Path:/arbois",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"server-test1-123456789abc-79533a101142718f0fdf84c42593c41e": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
						"server-test1-123456789abc-315a41140f1bd825b066e39686c18482": {
							URL:    "https://127.0.0.1:2504",
							Weight: 90,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "several segments with the same backend name and different port binding",
			instanceInfo: []ecsInstance{
				instance(
					ID("123456789abc"),
					name("test1"),
					labels(map[string]string{
						"traefik.protocol":                     "https",
						"traefik.frontend.entryPoints":         "http,https",
						"traefik.frontend.redirect.entryPoint": "https",

						"traefik.sauternes.port":              "2503",
						"traefik.sauternes.weight":            "80",
						"traefik.sauternes.backend":           "foobar",
						"traefik.sauternes.frontend.rule":     "Path:/sauternes",
						"traefik.sauternes.frontend.priority": "5000",

						"traefik.arbois.port":              "8080",
						"traefik.arbois.weight":            "90",
						"traefik.arbois.backend":           "foobar",
						"traefik.arbois.frontend.rule":     "Path:/arbois",
						"traefik.arbois.frontend.priority": "3000",
					}),
					iMachine(
						mName("machine1"),
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 2503),
							mPort(8080, 2504),
						),
					),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-sauternes-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       5000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-sauternes-test1-foobar": {
							Rule: "Path:/sauternes",
						},
					},
				},
				"frontend-arbois-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: true,
					Priority:       3000,
					EntryPoints:    []string{"http", "https"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-arbois-test1-foobar": {
							Rule: "Path:/arbois",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"server-test1-123456789abc-79533a101142718f0fdf84c42593c41e": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
						"server-test1-123456789abc-315a41140f1bd825b066e39686c18482": {
							URL:    "https://127.0.0.1:2504",
							Weight: 90,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	provider := &Provider{
		Domain:           "ecs.localhost",
		ExposedByDefault: true,
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualConfig, err := provider.buildConfiguration(test.instanceInfo)

			assert.NoError(t, err)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}
