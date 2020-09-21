package mesos

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/provider/label"
	"github.com/traefik/traefik/types"
)

func TestBuildConfigurationSegments(t *testing.T) {
	p := &Provider{
		Domain:           "mesos.localhost",
		ExposedByDefault: true,
		IPSources:        "host",
	}

	testCases := []struct {
		desc              string
		tasks             []state.Task
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "multiple ports with segments",
			tasks: []state.Task{
				aTask("app-taskID",
					withIP("127.0.0.1"),
					withInfo("/app",
						withPorts(
							withPort("TCP", 80, "web"),
							withPort("TCP", 81, "admin"),
						),
					),
					withStatus(withHealthy(true), withState("TASK_RUNNING")),
					withLabel(label.TraefikBackendMaxConnAmount, "1000"),
					withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
					withSegmentLabel(label.TraefikPort, "80", "web"),
					withSegmentLabel(label.TraefikPort, "81", "admin"),
					withLabel("traefik..port", "82"), // This should be ignored, as it fails to match the segmentPropertiesRegexp regex.
					withSegmentLabel(label.TraefikFrontendRule, "Host:web.app.mesos.localhost", "web"),
					withSegmentLabel(label.TraefikFrontendRule, "Host:admin.app.mesos.localhost", "admin"),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-taskID-service-web": {
					Backend: "backend-app-service-web",
					Routes: map[string]types.Route{
						`route-host-app-taskID-service-web`: {
							Rule: "Host:web.app.mesos.localhost",
						},
					},
					PassHostHeader: true,
					EntryPoints:    []string{},
				},
				"frontend-app-taskID-service-admin": {
					Backend: "backend-app-service-admin",
					Routes: map[string]types.Route{
						`route-host-app-taskID-service-admin`: {
							Rule: "Host:admin.app.mesos.localhost",
						},
					},
					PassHostHeader: true,
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-web": {
					Servers: map[string]types.Server{
						"server-app-taskID-service-web": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
				"backend-app-service-admin": {
					Servers: map[string]types.Server{
						"server-app-taskID-service-admin": {
							URL:    "http://127.0.0.1:81",
							Weight: label.DefaultWeight,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
		{
			desc: "when all labels are set",
			tasks: []state.Task{
				aTask("app-taskID",
					withIP("127.0.0.1"),
					withInfo("/app",
						withPorts(
							withPort("TCP", 80, "web"),
							withPort("TCP", 81, "admin"),
						),
					),
					withStatus(withHealthy(true), withState("TASK_RUNNING")),

					withLabel(label.TraefikBackendCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
					withLabel(label.TraefikBackendHealthCheckScheme, "http"),
					withLabel(label.TraefikBackendHealthCheckPath, "/health"),
					withLabel(label.TraefikBackendHealthCheckPort, "880"),
					withLabel(label.TraefikBackendHealthCheckInterval, "6"),
					withLabel(label.TraefikBackendHealthCheckHostname, "foo.com"),
					withLabel(label.TraefikBackendHealthCheckHeaders, "Foo:bar || Bar:foo"),
					withLabel(label.TraefikBackendLoadBalancerMethod, "drr"),
					withLabel(label.TraefikBackendLoadBalancerSticky, "true"),
					withLabel(label.TraefikBackendLoadBalancerStickiness, "true"),
					withLabel(label.TraefikBackendLoadBalancerStickinessCookieName, "chocolate"),
					withLabel(label.TraefikBackendLoadBalancerStickinessSecure, "true"),
					withLabel(label.TraefikBackendLoadBalancerStickinessHTTPOnly, "true"),
					withLabel(label.TraefikBackendLoadBalancerStickinessSameSite, "none"),
					withLabel(label.TraefikBackendMaxConnAmount, "666"),
					withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
					withLabel(label.TraefikBackendBufferingMaxResponseBodyBytes, "10485760"),
					withLabel(label.TraefikBackendBufferingMemResponseBodyBytes, "2097152"),
					withLabel(label.TraefikBackendBufferingMaxRequestBodyBytes, "10485760"),
					withLabel(label.TraefikBackendBufferingMemRequestBodyBytes, "2097152"),
					withLabel(label.TraefikBackendBufferingRetryExpression, "IsNetworkError() && Attempts() <= 2"),

					withSegmentLabel(label.TraefikPort, "80", "traefiklabs"),
					withSegmentLabel(label.TraefikPortName, "web", "traefiklabs"),
					withSegmentLabel(label.TraefikProtocol, "https", "traefiklabs"),
					withSegmentLabel(label.TraefikWeight, "12", "traefiklabs"),

					withSegmentLabel(label.TraefikFrontendPassTLSClientCertPem, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosNotBefore, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosNotAfter, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSans, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerCountry, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerLocality, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerProvince, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectCountry, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectLocality, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectProvince, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber, "true", "traefiklabs"),

					withSegmentLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthBasicRemoveHeader, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthBasicUsersFile, ".htpasswd", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthDigestRemoveHeader, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthDigestUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthDigestUsersFile, ".htpasswd", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardAddress, "auth.server", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTrustForwardHeader, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCa, "ca.crt", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCaOptional, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCert, "server.crt", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSKey, "server.key", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSInsecureSkipVerify, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAuthHeaderField, "X-WebAuth-User", "traefiklabs"),

					withSegmentLabel(label.TraefikFrontendEntryPoints, "http,https", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassHostHeader, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPassTLSCert, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPriority, "666", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendRedirectEntryPoint, "https", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendRedirectRegex, "nope", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendRedirectReplacement, "nope", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendRedirectPermanent, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendRule, "Host:traefik.io", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendWhiteListSourceRange, "10.10.10.10", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendWhiteListUseXForwardedFor, "true", "traefiklabs"),

					withSegmentLabel(label.TraefikFrontendRequestHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendResponseHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSSLProxyHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendAllowedHosts, "foo,bar,bor", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendHostsProxyHeaders, "foo,bar,bor", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSSLForceHost, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSSLHost, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendCustomFrameOptionsValue, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendContentSecurityPolicy, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendPublicKey, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendReferrerPolicy, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendCustomBrowserXSSValue, "foo", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSTSSeconds, "666", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSSLRedirect, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSSLTemporaryRedirect, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSTSIncludeSubdomains, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendSTSPreload, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendForceSTSHeader, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendFrameDeny, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendContentTypeNosniff, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendBrowserXSSFilter, "true", "traefiklabs"),
					withSegmentLabel(label.TraefikFrontendIsDevelopment, "true", "traefiklabs"),

					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageStatus, "404"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageBackend, "foobar"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageQuery, "foo_query"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageStatus, "500,600"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageBackend, "foobar"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageQuery, "bar_query"),

					withSegmentLabel(label.TraefikFrontendRateLimitExtractorFunc, "client.ip", "traefiklabs"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitPeriod, "6"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitAverage, "12"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitBurst, "18"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitPeriod, "3"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitAverage, "6"),
					withLabel(label.Prefix+"traefiklabs."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitBurst, "9"),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-taskID-service-traefiklabs": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-app-service-traefiklabs",
					Routes: map[string]types.Route{
						"route-host-app-taskID-service-traefiklabs": {
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
						EntryPoint: "https",
						Permanent:  true,
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-traefiklabs": {
					Servers: map[string]types.Server{
						"server-app-taskID-service-traefiklabs": {
							URL:    "https://127.0.0.1:80",
							Weight: 12,
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
							Secure:     true,
							HTTPOnly:   true,
							SameSite:   "none",
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
							"Bar": "foo",
							"Foo": "bar",
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

			actualConfig := p.buildConfigurationV2(test.tasks)

			require.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}
