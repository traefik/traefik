package mesos

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/mesosphere/mesos-dns/records/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					withLabel(label.TraefikBackendMaxConnAmount, "666"),
					withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
					withLabel(label.TraefikBackendBufferingMaxResponseBodyBytes, "10485760"),
					withLabel(label.TraefikBackendBufferingMemResponseBodyBytes, "2097152"),
					withLabel(label.TraefikBackendBufferingMaxRequestBodyBytes, "10485760"),
					withLabel(label.TraefikBackendBufferingMemRequestBodyBytes, "2097152"),
					withLabel(label.TraefikBackendBufferingRetryExpression, "IsNetworkError() && Attempts() <= 2"),

					withSegmentLabel(label.TraefikPort, "80", "containous"),
					withSegmentLabel(label.TraefikPortName, "web", "containous"),
					withSegmentLabel(label.TraefikProtocol, "https", "containous"),
					withSegmentLabel(label.TraefikWeight, "12", "containous"),

					withSegmentLabel(label.TraefikFrontendPassTLSClientCertPem, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosNotBefore, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosNotAfter, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSans, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerCommonName, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerCountry, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerLocality, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerOrganization, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerProvince, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectCommonName, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectCountry, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectLocality, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectOrganization, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectProvince, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber, "true", "containous"),

					withSegmentLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthBasicRemoveHeader, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthBasicUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthBasicUsersFile, ".htpasswd", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthDigestRemoveHeader, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthDigestUsers, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthDigestUsersFile, ".htpasswd", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardAddress, "auth.server", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTrustForwardHeader, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCa, "ca.crt", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCaOptional, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSCert, "server.crt", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSKey, "server.key", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthForwardTLSInsecureSkipVerify, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendAuthHeaderField, "X-WebAuth-User", "containous"),

					withSegmentLabel(label.TraefikFrontendEntryPoints, "http,https", "containous"),
					withSegmentLabel(label.TraefikFrontendPassHostHeader, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPassTLSCert, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendPriority, "666", "containous"),
					withSegmentLabel(label.TraefikFrontendRedirectEntryPoint, "https", "containous"),
					withSegmentLabel(label.TraefikFrontendRedirectRegex, "nope", "containous"),
					withSegmentLabel(label.TraefikFrontendRedirectReplacement, "nope", "containous"),
					withSegmentLabel(label.TraefikFrontendRedirectPermanent, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendRule, "Host:traefik.io", "containous"),
					withSegmentLabel(label.TraefikFrontendWhiteListSourceRange, "10.10.10.10", "containous"),
					withSegmentLabel(label.TraefikFrontendWhiteListUseXForwardedFor, "true", "containous"),

					withSegmentLabel(label.TraefikFrontendRequestHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
					withSegmentLabel(label.TraefikFrontendResponseHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
					withSegmentLabel(label.TraefikFrontendSSLProxyHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
					withSegmentLabel(label.TraefikFrontendAllowedHosts, "foo,bar,bor", "containous"),
					withSegmentLabel(label.TraefikFrontendHostsProxyHeaders, "foo,bar,bor", "containous"),
					withSegmentLabel(label.TraefikFrontendSSLForceHost, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendSSLHost, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendCustomFrameOptionsValue, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendContentSecurityPolicy, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendPublicKey, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendReferrerPolicy, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendCustomBrowserXSSValue, "foo", "containous"),
					withSegmentLabel(label.TraefikFrontendSTSSeconds, "666", "containous"),
					withSegmentLabel(label.TraefikFrontendSSLRedirect, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendSSLTemporaryRedirect, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendSTSIncludeSubdomains, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendSTSPreload, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendForceSTSHeader, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendFrameDeny, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendContentTypeNosniff, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendBrowserXSSFilter, "true", "containous"),
					withSegmentLabel(label.TraefikFrontendIsDevelopment, "true", "containous"),

					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageStatus, "404"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageBackend, "foobar"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageQuery, "foo_query"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageStatus, "500,600"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageBackend, "foobar"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageQuery, "bar_query"),

					withSegmentLabel(label.TraefikFrontendRateLimitExtractorFunc, "client.ip", "containous"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitPeriod, "6"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitAverage, "12"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitBurst, "18"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitPeriod, "3"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitAverage, "6"),
					withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitBurst, "9"),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-taskID-service-containous": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-app-service-containous",
					Routes: map[string]types.Route{
						"route-host-app-taskID-service-containous": {
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
				"backend-app-service-containous": {
					Servers: map[string]types.Server{
						"server-app-taskID-service-containous": {
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
