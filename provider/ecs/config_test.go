package ecs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc      string
		instances []ecsInstance
		expected  *types.Configuration
		err       error
	}{
		{
			desc: "config parsed successfully",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
							},
						},
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "config parsed successfully with health check labels",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{
						label.TraefikBackendHealthCheckPath:     aws.String("/health"),
						label.TraefikBackendHealthCheckInterval: aws.String("1s"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						HealthCheck: &types.HealthCheck{
							Path:     "/health",
							Interval: "1s",
						},
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
							},
						},
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "config parsed successfully with basic auth labels",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{
						label.TraefikFrontendAuthBasicUsers:        aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendAuthBasicUsersFile:    aws.String(".htpasswd"),
						label.TraefikFrontendAuthBasicRemoveHeader: aws.String("true"),
						label.TraefikFrontendAuthHeaderField:       aws.String("X-WebAuth-User"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
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
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "config parsed successfully with basic auth (backward compatibility) labels",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{
						label.TraefikFrontendAuthBasic: aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
							},
						},
						Auth: &types.Auth{
							Basic: &types.Basic{
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "config parsed successfully with digest auth labels",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{
						label.TraefikFrontendAuthDigestRemoveHeader: aws.String("true"),
						label.TraefikFrontendAuthDigestUsers:        aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendAuthDigestUsersFile:    aws.String(".htpasswd"),
						label.TraefikFrontendAuthHeaderField:        aws.String("X-WebAuth-User"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
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
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "config parsed successfully with forward auth labels",
			instances: []ecsInstance{
				instance(
					name("instance"),
					ID("1"),
					dockerLabels(map[string]*string{
						label.TraefikFrontendAuthForwardAddress:               aws.String("auth.server"),
						label.TraefikFrontendAuthForwardTrustForwardHeader:    aws.String("true"),
						label.TraefikFrontendAuthForwardTLSCa:                 aws.String("ca.crt"),
						label.TraefikFrontendAuthForwardTLSCaOptional:         aws.String("true"),
						label.TraefikFrontendAuthForwardTLSCert:               aws.String("server.crt"),
						label.TraefikFrontendAuthForwardTLSKey:                aws.String("server.key"),
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: aws.String("true"),
						label.TraefikFrontendAuthHeaderField:                  aws.String("X-WebAuth-User"),
						label.TraefikFrontendAuthForwardAuthResponseHeaders:   aws.String("X-Auth-User,X-Auth-Token"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-instance": {
						Servers: map[string]types.Server{
							"server-instance-1": {
								URL:    "http://10.0.0.1:1337",
								Weight: label.DefaultWeight,
							}},
					},
				},
				Frontends: map[string]*types.Frontend{
					"frontend-instance": {
						EntryPoints: []string{},
						Backend:     "backend-instance",
						Routes: map[string]types.Route{
							"route-frontend-instance": {
								Rule: "Host:instance",
							},
						},
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
						PassHostHeader: true,
					},
				},
			},
		},
		{
			desc: "when all labels are set",
			instances: []ecsInstance{
				instance(
					name("testing-instance"),
					ID("6"),
					dockerLabels(map[string]*string{
						label.TraefikPort:     aws.String("666"),
						label.TraefikProtocol: aws.String("https"),
						label.TraefikWeight:   aws.String("12"),

						label.TraefikBackend: aws.String("foobar"),

						label.TraefikBackendCircuitBreakerExpression:         aws.String("NetworkErrorRatio() > 0.5"),
						label.TraefikBackendResponseForwardingFlushInterval:  aws.String("10ms"),
						label.TraefikBackendHealthCheckScheme:                aws.String("http"),
						label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
						label.TraefikBackendHealthCheckPort:                  aws.String("880"),
						label.TraefikBackendHealthCheckInterval:              aws.String("6"),
						label.TraefikBackendHealthCheckHostname:              aws.String("foo.com"),
						label.TraefikBackendHealthCheckHeaders:               aws.String("Foo:bar || Bar:foo"),
						label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
						label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
						label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
						label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),
						label.TraefikBackendMaxConnAmount:                    aws.String("666"),
						label.TraefikBackendMaxConnExtractorFunc:             aws.String("client.ip"),
						label.TraefikBackendBufferingMaxResponseBodyBytes:    aws.String("10485760"),
						label.TraefikBackendBufferingMemResponseBodyBytes:    aws.String("2097152"),
						label.TraefikBackendBufferingMaxRequestBodyBytes:     aws.String("10485760"),
						label.TraefikBackendBufferingMemRequestBodyBytes:     aws.String("2097152"),
						label.TraefikBackendBufferingRetryExpression:         aws.String("IsNetworkError() && Attempts() <= 2"),

						label.TraefikFrontendPassTLSClientCertPem:                         aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosNotBefore:              aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosNotAfter:               aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSans:                   aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerCountry:          aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerLocality:         aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerProvince:         aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectCountry:         aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectLocality:        aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectProvince:        aws.String("true"),
						label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    aws.String("true"),

						label.TraefikFrontendAuthBasic:                        aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendAuthBasicRemoveHeader:            aws.String("true"),
						label.TraefikFrontendAuthBasicUsers:                   aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendAuthBasicUsersFile:               aws.String(".htpasswd"),
						label.TraefikFrontendAuthDigestRemoveHeader:           aws.String("true"),
						label.TraefikFrontendAuthDigestUsers:                  aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendAuthDigestUsersFile:              aws.String(".htpasswd"),
						label.TraefikFrontendAuthForwardAddress:               aws.String("auth.server"),
						label.TraefikFrontendAuthForwardTrustForwardHeader:    aws.String("true"),
						label.TraefikFrontendAuthForwardTLSCa:                 aws.String("ca.crt"),
						label.TraefikFrontendAuthForwardTLSCaOptional:         aws.String("true"),
						label.TraefikFrontendAuthForwardTLSCert:               aws.String("server.crt"),
						label.TraefikFrontendAuthForwardTLSKey:                aws.String("server.key"),
						label.TraefikFrontendAuthForwardTLSInsecureSkipVerify: aws.String("true"),
						label.TraefikFrontendAuthHeaderField:                  aws.String("X-WebAuth-User"),

						label.TraefikFrontendEntryPoints:               aws.String("http,https"),
						label.TraefikFrontendPassHostHeader:            aws.String("true"),
						label.TraefikFrontendPassTLSCert:               aws.String("true"),
						label.TraefikFrontendPriority:                  aws.String("666"),
						label.TraefikFrontendRedirectEntryPoint:        aws.String("https"),
						label.TraefikFrontendRedirectRegex:             aws.String("nope"),
						label.TraefikFrontendRedirectReplacement:       aws.String("nope"),
						label.TraefikFrontendRedirectPermanent:         aws.String("true"),
						label.TraefikFrontendRule:                      aws.String("Host:traefik.io"),
						label.TraefikFrontendWhiteListSourceRange:      aws.String("10.10.10.10"),
						label.TraefikFrontendWhiteListUseXForwardedFor: aws.String("true"),

						label.TraefikFrontendRequestHeaders:          aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendResponseHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendSSLProxyHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendAllowedHosts:            aws.String("foo,bar,bor"),
						label.TraefikFrontendHostsProxyHeaders:       aws.String("foo,bar,bor"),
						label.TraefikFrontendSSLHost:                 aws.String("foo"),
						label.TraefikFrontendCustomFrameOptionsValue: aws.String("foo"),
						label.TraefikFrontendContentSecurityPolicy:   aws.String("foo"),
						label.TraefikFrontendPublicKey:               aws.String("foo"),
						label.TraefikFrontendReferrerPolicy:          aws.String("foo"),
						label.TraefikFrontendCustomBrowserXSSValue:   aws.String("foo"),
						label.TraefikFrontendSTSSeconds:              aws.String("666"),
						label.TraefikFrontendSSLForceHost:            aws.String("true"),
						label.TraefikFrontendSSLRedirect:             aws.String("true"),
						label.TraefikFrontendSSLTemporaryRedirect:    aws.String("true"),
						label.TraefikFrontendSTSIncludeSubdomains:    aws.String("true"),
						label.TraefikFrontendSTSPreload:              aws.String("true"),
						label.TraefikFrontendForceSTSHeader:          aws.String("true"),
						label.TraefikFrontendFrameDeny:               aws.String("true"),
						label.TraefikFrontendContentTypeNosniff:      aws.String("true"),
						label.TraefikFrontendBrowserXSSFilter:        aws.String("true"),
						label.TraefikFrontendIsDevelopment:           aws.String("true"),

						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  aws.String("404"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   aws.String("foo_query"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  aws.String("500,600"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   aws.String("bar_query"),

						label.TraefikFrontendRateLimitExtractorFunc:                                        aws.String("client.ip"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: aws.String("12"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   aws.String("18"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  aws.String("3"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   aws.String("9"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-foobar": {
						Servers: map[string]types.Server{
							"server-testing-instance-6": {
								URL:    "https://10.0.0.1:666",
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
				Frontends: map[string]*types.Frontend{
					"frontend-foobar": {
						EntryPoints: []string{
							"http",
							"https",
						},
						Backend: "backend-foobar",
						Routes: map[string]types.Route{
							"route-frontend-foobar": {
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
							"bar": {
								Status: []string{
									"500",
									"600",
								},
								Backend: "backend-foobar",
								Query:   "bar_query",
							},
							"foo": {
								Status: []string{
									"404",
								},
								Backend: "backend-foobar",
								Query:   "foo_query",
							},
						},
						RateLimit: &types.RateLimit{
							RateSet: map[string]*types.Rate{
								"bar": {
									Period:  flaeg.Duration(3 * time.Second),
									Average: 6,
									Burst:   9,
								},
								"foo": {
									Period:  flaeg.Duration(6 * time.Second),
									Average: 12,
									Burst:   18,
								},
							},
							ExtractorFunc: "client.ip",
						},
						Redirect: &types.Redirect{
							EntryPoint:  "https",
							Regex:       "",
							Replacement: "",
							Permanent:   true,
						},
					},
				},
			},
		},
		{
			desc: "Containers with same backend name",
			instances: []ecsInstance{
				instance(
					name("testing-instance-v1"),
					ID("6"),
					dockerLabels(map[string]*string{
						label.TraefikPort:     aws.String("666"),
						label.TraefikProtocol: aws.String("https"),
						label.TraefikWeight:   aws.String("12"),

						label.TraefikBackend: aws.String("foobar"),

						label.TraefikBackendCircuitBreakerExpression:         aws.String("NetworkErrorRatio() > 0.5"),
						label.TraefikBackendHealthCheckScheme:                aws.String("http"),
						label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
						label.TraefikBackendHealthCheckPort:                  aws.String("880"),
						label.TraefikBackendHealthCheckInterval:              aws.String("6"),
						label.TraefikBackendHealthCheckHostname:              aws.String("foo.com"),
						label.TraefikBackendHealthCheckHeaders:               aws.String("Foo:bar || Bar:foo"),
						label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
						label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
						label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
						label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),
						label.TraefikBackendMaxConnAmount:                    aws.String("666"),
						label.TraefikBackendMaxConnExtractorFunc:             aws.String("client.ip"),
						label.TraefikBackendBufferingMaxResponseBodyBytes:    aws.String("10485760"),
						label.TraefikBackendBufferingMemResponseBodyBytes:    aws.String("2097152"),
						label.TraefikBackendBufferingMaxRequestBodyBytes:     aws.String("10485760"),
						label.TraefikBackendBufferingMemRequestBodyBytes:     aws.String("2097152"),
						label.TraefikBackendBufferingRetryExpression:         aws.String("IsNetworkError() && Attempts() <= 2"),

						label.TraefikFrontendAuthBasicUsers:            aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendEntryPoints:               aws.String("http,https"),
						label.TraefikFrontendPassHostHeader:            aws.String("true"),
						label.TraefikFrontendPassTLSCert:               aws.String("true"),
						label.TraefikFrontendPriority:                  aws.String("666"),
						label.TraefikFrontendRedirectEntryPoint:        aws.String("https"),
						label.TraefikFrontendRedirectRegex:             aws.String("nope"),
						label.TraefikFrontendRedirectReplacement:       aws.String("nope"),
						label.TraefikFrontendRedirectPermanent:         aws.String("true"),
						label.TraefikFrontendRule:                      aws.String("Host:traefik.io"),
						label.TraefikFrontendWhiteListSourceRange:      aws.String("10.10.10.10"),
						label.TraefikFrontendWhiteListUseXForwardedFor: aws.String("true"),

						label.TraefikFrontendRequestHeaders:          aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendResponseHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendSSLProxyHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendAllowedHosts:            aws.String("foo,bar,bor"),
						label.TraefikFrontendHostsProxyHeaders:       aws.String("foo,bar,bor"),
						label.TraefikFrontendSSLHost:                 aws.String("foo"),
						label.TraefikFrontendCustomFrameOptionsValue: aws.String("foo"),
						label.TraefikFrontendContentSecurityPolicy:   aws.String("foo"),
						label.TraefikFrontendPublicKey:               aws.String("foo"),
						label.TraefikFrontendReferrerPolicy:          aws.String("foo"),
						label.TraefikFrontendCustomBrowserXSSValue:   aws.String("foo"),
						label.TraefikFrontendSTSSeconds:              aws.String("666"),
						label.TraefikFrontendSSLForceHost:            aws.String("true"),
						label.TraefikFrontendSSLRedirect:             aws.String("true"),
						label.TraefikFrontendSSLTemporaryRedirect:    aws.String("true"),
						label.TraefikFrontendSTSIncludeSubdomains:    aws.String("true"),
						label.TraefikFrontendSTSPreload:              aws.String("true"),
						label.TraefikFrontendForceSTSHeader:          aws.String("true"),
						label.TraefikFrontendFrameDeny:               aws.String("true"),
						label.TraefikFrontendContentTypeNosniff:      aws.String("true"),
						label.TraefikFrontendBrowserXSSFilter:        aws.String("true"),
						label.TraefikFrontendIsDevelopment:           aws.String("true"),

						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  aws.String("404"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   aws.String("foo_query"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  aws.String("500,600"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   aws.String("bar_query"),

						label.TraefikFrontendRateLimitExtractorFunc:                                        aws.String("client.ip"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: aws.String("12"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   aws.String("18"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  aws.String("3"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   aws.String("9"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
				instance(
					name("testing-instance-v2"),
					ID("6"),
					dockerLabels(map[string]*string{
						label.TraefikPort:     aws.String("555"),
						label.TraefikProtocol: aws.String("https"),
						label.TraefikWeight:   aws.String("15"),

						label.TraefikBackend: aws.String("foobar"),

						label.TraefikBackendCircuitBreakerExpression:         aws.String("NetworkErrorRatio() > 0.5"),
						label.TraefikBackendHealthCheckScheme:                aws.String("http"),
						label.TraefikBackendHealthCheckPath:                  aws.String("/health"),
						label.TraefikBackendHealthCheckPort:                  aws.String("880"),
						label.TraefikBackendHealthCheckInterval:              aws.String("6"),
						label.TraefikBackendHealthCheckHostname:              aws.String("bar.com"),
						label.TraefikBackendHealthCheckHeaders:               aws.String("Foo:bar || Bar:foo"),
						label.TraefikBackendLoadBalancerMethod:               aws.String("drr"),
						label.TraefikBackendLoadBalancerSticky:               aws.String("true"),
						label.TraefikBackendLoadBalancerStickiness:           aws.String("true"),
						label.TraefikBackendLoadBalancerStickinessCookieName: aws.String("chocolate"),
						label.TraefikBackendMaxConnAmount:                    aws.String("666"),
						label.TraefikBackendMaxConnExtractorFunc:             aws.String("client.ip"),
						label.TraefikBackendBufferingMaxResponseBodyBytes:    aws.String("10485760"),
						label.TraefikBackendBufferingMemResponseBodyBytes:    aws.String("2097152"),
						label.TraefikBackendBufferingMaxRequestBodyBytes:     aws.String("10485760"),
						label.TraefikBackendBufferingMemRequestBodyBytes:     aws.String("2097152"),
						label.TraefikBackendBufferingRetryExpression:         aws.String("IsNetworkError() && Attempts() <= 2"),

						label.TraefikFrontendAuthBasic:                 aws.String("test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
						label.TraefikFrontendEntryPoints:               aws.String("http,https"),
						label.TraefikFrontendPassHostHeader:            aws.String("true"),
						label.TraefikFrontendPassTLSCert:               aws.String("true"),
						label.TraefikFrontendPriority:                  aws.String("666"),
						label.TraefikFrontendRedirectEntryPoint:        aws.String("https"),
						label.TraefikFrontendRedirectRegex:             aws.String("nope"),
						label.TraefikFrontendRedirectReplacement:       aws.String("nope"),
						label.TraefikFrontendRedirectPermanent:         aws.String("true"),
						label.TraefikFrontendRule:                      aws.String("Host:traefik.io"),
						label.TraefikFrontendWhiteListSourceRange:      aws.String("10.10.10.10"),
						label.TraefikFrontendWhiteListUseXForwardedFor: aws.String("true"),

						label.TraefikFrontendRequestHeaders:          aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendResponseHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendSSLProxyHeaders:         aws.String("Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
						label.TraefikFrontendAllowedHosts:            aws.String("foo,bar,bor"),
						label.TraefikFrontendHostsProxyHeaders:       aws.String("foo,bar,bor"),
						label.TraefikFrontendSSLHost:                 aws.String("foo"),
						label.TraefikFrontendCustomFrameOptionsValue: aws.String("foo"),
						label.TraefikFrontendContentSecurityPolicy:   aws.String("foo"),
						label.TraefikFrontendPublicKey:               aws.String("foo"),
						label.TraefikFrontendReferrerPolicy:          aws.String("foo"),
						label.TraefikFrontendCustomBrowserXSSValue:   aws.String("foo"),
						label.TraefikFrontendSTSSeconds:              aws.String("666"),
						label.TraefikFrontendSSLForceHost:            aws.String("true"),
						label.TraefikFrontendSSLRedirect:             aws.String("true"),
						label.TraefikFrontendSSLTemporaryRedirect:    aws.String("true"),
						label.TraefikFrontendSTSIncludeSubdomains:    aws.String("true"),
						label.TraefikFrontendSTSPreload:              aws.String("true"),
						label.TraefikFrontendForceSTSHeader:          aws.String("true"),
						label.TraefikFrontendFrameDeny:               aws.String("true"),
						label.TraefikFrontendContentTypeNosniff:      aws.String("true"),
						label.TraefikFrontendBrowserXSSFilter:        aws.String("true"),
						label.TraefikFrontendIsDevelopment:           aws.String("true"),

						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  aws.String("404"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   aws.String("foo_query"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  aws.String("500,600"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: aws.String("foobar"),
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   aws.String("bar_query"),

						label.TraefikFrontendRateLimitExtractorFunc:                                        aws.String("client.ip"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: aws.String("12"),
						label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   aws.String("18"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  aws.String("3"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: aws.String("6"),
						label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   aws.String("9"),
					}),
					iMachine(
						mState(ec2.InstanceStateNameRunning),
						mPrivateIP("10.2.2.1"),
						mPorts(
							mPort(0, 1337),
						),
					),
				),
			},
			expected: &types.Configuration{
				Backends: map[string]*types.Backend{
					"backend-foobar": {
						Servers: map[string]types.Server{
							"server-testing-instance-v1-6": {
								URL:    "https://10.0.0.1:666",
								Weight: 12,
							},
							"server-testing-instance-v2-6": {
								URL:    "https://10.2.2.1:555",
								Weight: 15,
							},
						},
						CircuitBreaker: &types.CircuitBreaker{
							Expression: "NetworkErrorRatio() > 0.5",
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
				Frontends: map[string]*types.Frontend{
					"frontend-foobar": {
						EntryPoints: []string{
							"http",
							"https",
						},
						Backend: "backend-foobar",
						Routes: map[string]types.Route{
							"route-frontend-foobar": {
								Rule: "Host:traefik.io",
							},
						},
						PassHostHeader: true,
						PassTLSCert:    true,
						Priority:       666,
						Auth: &types.Auth{
							Basic: &types.Basic{
								Users: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
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
							"bar": {
								Status: []string{
									"500",
									"600",
								},
								Backend: "backend-foobar",
								Query:   "bar_query",
							},
							"foo": {
								Status: []string{
									"404",
								},
								Backend: "backend-foobar",
								Query:   "foo_query",
							},
						},
						RateLimit: &types.RateLimit{
							RateSet: map[string]*types.Rate{
								"bar": {
									Period:  flaeg.Duration(3 * time.Second),
									Average: 6,
									Burst:   9,
								},
								"foo": {
									Period:  flaeg.Duration(6 * time.Second),
									Average: 12,
									Burst:   18,
								},
							},
							ExtractorFunc: "client.ip",
						},
						Redirect: &types.Redirect{
							EntryPoint:  "https",
							Regex:       "",
							Replacement: "",
							Permanent:   true,
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{ExposedByDefault: true}

			instances := fakeLoadTraefikLabels(test.instances)

			got, err := p.buildConfiguration(instances)
			assert.Equal(t, test.err, err) // , err.Error()
			assert.Equal(t, test.expected, got, test.desc)
		})
	}
}

func TestFilterInstance(t *testing.T) {
	testCases := []struct {
		desc             string
		instanceInfo     ecsInstance
		exposedByDefault bool
		expected         bool
		constrain        bool
	}{
		{
			desc:             "Instance without enable label and exposed by default enabled should be not filtered",
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
			exposedByDefault: true,
			expected:         true,
		},
		{
			desc:             "Instance without enable label and exposed by default disabled should be filtered",
			instanceInfo:     simpleEcsInstance(map[string]*string{}),
			exposedByDefault: false,
			expected:         false,
		},
		{
			desc: "Instance with enable label set to false and exposed by default enabled should be filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikEnable: aws.String("false"),
			}),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with enable label set to true and exposed by default disabled should be not filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikEnable: aws.String("true"),
			}),
			exposedByDefault: false,
			expected:         true,
		},
		{
			desc: "Instance with empty private ip and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilPrivateIP := simpleEcsInstance(map[string]*string{})
				nilPrivateIP.machine.privateIP = ""
				return nilPrivateIP
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with nil machine and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilMachine := simpleEcsInstance(map[string]*string{})
				nilMachine.machine = nil
				return nilMachine
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with empty machine state and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				nilMachineState := simpleEcsInstance(map[string]*string{})
				nilMachineState.machine.state = ""
				return nilMachineState
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with invalid machine state and exposed by default enabled should be filtered",
			instanceInfo: func() ecsInstance {
				invalidMachineState := simpleEcsInstance(map[string]*string{})
				invalidMachineState.machine.state = ec2.InstanceStateNameStopped
				return invalidMachineState
			}(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc:             "Instance with no port mappings should be filtered",
			instanceInfo:     simpleEcsInstanceNoNetwork(map[string]*string{}),
			exposedByDefault: true,
			expected:         false,
		},
		{
			desc: "Instance with no port mapping and with label should not be filtered",
			instanceInfo: simpleEcsInstanceNoNetwork(map[string]*string{
				label.TraefikPort: aws.String("80"),
			}),
			exposedByDefault: true,
			expected:         true,
		},
		{
			desc: "Instance with failing constraint should be filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikTags: aws.String("private"),
			}),
			exposedByDefault: true,
			expected:         false,
			constrain:        true,
		},
		{
			desc: "Instance with passing constraint should not be filtered",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikTags: aws.String("public"),
			}),
			exposedByDefault: true,
			expected:         true,
			constrain:        true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			prov := &Provider{
				ExposedByDefault: test.exposedByDefault,
			}
			if test.constrain {
				constraints := types.Constraints{}
				assert.NoError(t, constraints.Set("tag==public"))
				prov.Constraints = constraints
			}

			actual := prov.filterInstance(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetHost(t *testing.T) {
	testCases := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Default host should be 10.0.0.0",
			expected:     "10.0.0.0",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getHost(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPort(t *testing.T) {
	testCases := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Default port should be 80",
			expected:     "80",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Label should override network port",
			expected: "4242",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikPort: aws.String("4242"),
			}),
		},
		{
			desc:     "Label should provide exposed port",
			expected: "80",
			instanceInfo: simpleEcsInstanceNoNetwork(map[string]*string{
				label.TraefikPort: aws.String("80"),
			}),
		},
		{
			desc:     "Container label should provide exposed port",
			expected: "6536",
			instanceInfo: simpleEcsInstanceDynamicPorts(map[string]*string{
				label.TraefikPort: aws.String("8080"),
			}),
		},
		{
			desc:     "Wrong port container label should provide default exposed port",
			expected: "9000",
			instanceInfo: simpleEcsInstanceDynamicPorts(map[string]*string{
				label.TraefikPort: aws.String("9000"),
			}),
		},
		{
			desc:     "Invalid port container label should provide default exposed port",
			expected: "6535",
			instanceInfo: simpleEcsInstanceDynamicPorts(map[string]*string{
				label.TraefikPort: aws.String("foo"),
			}),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getPort(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncStringValue(t *testing.T) {
	testCases := []struct {
		desc         string
		expected     string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Protocol label is not set should return a string equals to http",
			expected:     "http",
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Protocol label is set to http should return a string equals to http",
			expected: "http",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikProtocol: aws.String("http"),
			}),
		},
		{
			desc:     "Protocol label is set to https should return a string equals to https",
			expected: "https",
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikProtocol: aws.String("https"),
			}),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFuncStringValueV1(label.TraefikProtocol, label.DefaultProtocol)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFuncSliceString(t *testing.T) {
	testCases := []struct {
		desc         string
		expected     []string
		instanceInfo ecsInstance
	}{
		{
			desc:         "Frontend entrypoints label not set should return empty array",
			expected:     nil,
			instanceInfo: simpleEcsInstance(map[string]*string{}),
		},
		{
			desc:     "Frontend entrypoints label set to http should return a string array of 1 element",
			expected: []string{"http"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikFrontendEntryPoints: aws.String("http"),
			}),
		},
		{
			desc:     "Frontend entrypoints label set to http,https should return a string array of 2 elements",
			expected: []string{"http", "https"},
			instanceInfo: simpleEcsInstance(map[string]*string{
				label.TraefikFrontendEntryPoints: aws.String("http,https"),
			}),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getFuncSliceStringV1(label.TraefikFrontendEntryPoints)(test.instanceInfo)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func makeEcsInstance(containerDef *ecs.ContainerDefinition) ecsInstance {
	container := &ecs.Container{
		Name:            containerDef.Name,
		NetworkBindings: make([]*ecs.NetworkBinding, len(containerDef.PortMappings)),
	}

	for i, pm := range containerDef.PortMappings {
		container.NetworkBindings[i] = &ecs.NetworkBinding{
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
			Protocol:      pm.Protocol,
			BindIP:        aws.String("0.0.0.0"),
		}
	}

	instance := ecsInstance{
		Name:                "foo-http",
		ID:                  "123456789abc",
		containerDefinition: containerDef,
		machine: &machine{
			state:     ec2.InstanceStateNameRunning,
			privateIP: "10.0.0.0",
			ports:     []portMapping{{hostPort: 1337}},
		},
	}

	if containerDef != nil {
		instance.TraefikLabels = aws.StringValueMap(containerDef.DockerLabels)
	}

	return instance
}

func simpleEcsInstance(labels map[string]*string) ecsInstance {
	instance := makeEcsInstance(&ecs.ContainerDefinition{
		Name:         aws.String("http"),
		DockerLabels: labels,
	})
	instance.machine.ports = []portMapping{{hostPort: 80}}
	return instance
}

func simpleEcsInstanceNoNetwork(labels map[string]*string) ecsInstance {
	instance := makeEcsInstance(&ecs.ContainerDefinition{
		Name:         aws.String("http"),
		DockerLabels: labels,
	})
	instance.machine.ports = []portMapping{}
	return instance
}

func simpleEcsInstanceDynamicPorts(labels map[string]*string) ecsInstance {
	instance := makeEcsInstance(&ecs.ContainerDefinition{
		Name:         aws.String("http"),
		DockerLabels: labels,
	})
	instance.machine.ports = []portMapping{
		{
			containerPort: 80,
			hostPort:      6535,
		},
		{
			containerPort: 8080,
			hostPort:      6536,
		},
	}
	return instance
}

func fakeLoadTraefikLabels(instances []ecsInstance) []ecsInstance {
	var result []ecsInstance
	for _, instance := range instances {
		instance.TraefikLabels = aws.StringValueMap(instance.containerDefinition.DockerLabels)
		result = append(result, instance)
	}
	return result
}
