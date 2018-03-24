package marathon

import (
	"fmt"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigurationAPIErrors(t *testing.T) {
	fakeClient := newFakeClient(true, marathon.Applications{})

	p := &Provider{
		marathonClient: fakeClient,
	}

	actualConfig := p.getConfiguration()
	fakeClient.AssertExpectations(t)

	if actualConfig != nil {
		t.Errorf("configuration should have been nil, got %v", actualConfig)
	}
}

func TestBuildConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		application       marathon.Application
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "simple application",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "filtered task",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80), state(taskStateStaging))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {},
			},
		},
		{
			desc: "max connection extractor function label only",
			application: application(
				appPorts(80),
				withTasks(localhostTask(taskPorts(80))),

				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: nil,
				},
			},
		},
		{
			desc: "multiple ports",
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					Backend: "backend-app",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
				},
			},
		},
		{
			desc: "with all labels",
			application: application(
				appPorts(80),
				withTasks(task(host("127.0.0.1"), taskPorts(80))),

				withLabel(label.TraefikPort, "666"),
				withLabel(label.TraefikProtocol, "https"),
				withLabel(label.TraefikWeight, "12"),

				withLabel(label.TraefikBackend, "foobar"),

				withLabel(label.TraefikBackendCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
				withLabel(label.TraefikBackendHealthCheckPath, "/health"),
				withLabel(label.TraefikBackendHealthCheckPort, "880"),
				withLabel(label.TraefikBackendHealthCheckInterval, "6"),
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

				withLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"),
				withLabel(label.TraefikFrontendEntryPoints, "http,https"),
				withLabel(label.TraefikFrontendPassHostHeader, "true"),
				withLabel(label.TraefikFrontendPassTLSCert, "true"),
				withLabel(label.TraefikFrontendPriority, "666"),
				withLabel(label.TraefikFrontendRedirectEntryPoint, "https"),
				withLabel(label.TraefikFrontendRedirectRegex, "nope"),
				withLabel(label.TraefikFrontendRedirectReplacement, "nope"),
				withLabel(label.TraefikFrontendRedirectPermanent, "true"),
				withLabel(label.TraefikFrontendRule, "Host:traefik.io"),
				withLabel(label.TraefikFrontendWhiteListSourceRange, "10.10.10.10"),
				withLabel(label.TraefikFrontendWhiteListUseXForwardedFor, "true"),

				withLabel(label.TraefikFrontendRequestHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
				withLabel(label.TraefikFrontendResponseHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
				withLabel(label.TraefikFrontendSSLProxyHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
				withLabel(label.TraefikFrontendAllowedHosts, "foo,bar,bor"),
				withLabel(label.TraefikFrontendHostsProxyHeaders, "foo,bar,bor"),
				withLabel(label.TraefikFrontendSSLHost, "foo"),
				withLabel(label.TraefikFrontendCustomFrameOptionsValue, "foo"),
				withLabel(label.TraefikFrontendContentSecurityPolicy, "foo"),
				withLabel(label.TraefikFrontendPublicKey, "foo"),
				withLabel(label.TraefikFrontendReferrerPolicy, "foo"),
				withLabel(label.TraefikFrontendCustomBrowserXSSValue, "foo"),
				withLabel(label.TraefikFrontendSTSSeconds, "666"),
				withLabel(label.TraefikFrontendSSLRedirect, "true"),
				withLabel(label.TraefikFrontendSSLTemporaryRedirect, "true"),
				withLabel(label.TraefikFrontendSTSIncludeSubdomains, "true"),
				withLabel(label.TraefikFrontendSTSPreload, "true"),
				withLabel(label.TraefikFrontendForceSTSHeader, "true"),
				withLabel(label.TraefikFrontendFrameDeny, "true"),
				withLabel(label.TraefikFrontendContentTypeNosniff, "true"),
				withLabel(label.TraefikFrontendBrowserXSSFilter, "true"),
				withLabel(label.TraefikFrontendIsDevelopment, "true"),

				withLabel(label.Prefix+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageStatus, "404"),
				withLabel(label.Prefix+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageBackend, "foobar"),
				withLabel(label.Prefix+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageQuery, "foo_query"),
				withLabel(label.Prefix+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageStatus, "500,600"),
				withLabel(label.Prefix+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageBackend, "foobar"),
				withLabel(label.Prefix+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageQuery, "bar_query"),

				withLabel(label.TraefikFrontendRateLimitExtractorFunc, "client.ip"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitPeriod, "6"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitAverage, "12"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitBurst, "18"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitPeriod, "3"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitAverage, "6"),
				withLabel(label.Prefix+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitBurst, "9"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backendfoobar",
					Routes: map[string]types.Route{
						"route-host-app": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
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
							Backend: "foobar",
							Query:   "bar_query",
						},
						"foo": {
							Status: []string{
								"404",
							},
							Backend: "foobar",
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
				"backendfoobar": {
					Servers: map[string]types.Server{
						"server-task": {
							URL:    "https://127.0.0.1:666",
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
						Path:     "/health",
						Port:     880,
						Interval: "6",
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

			test.application.ID = "/app"

			for _, task := range test.application.Tasks {
				task.ID = "task"
				if task.State == "" {
					task.State = "TASK_RUNNING"
				}
			}

			p := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
			}

			actualConfig := p.buildConfigurationV2(withApplications(test.application))

			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestBuildConfigurationServices(t *testing.T) {
	testCases := []struct {
		desc              string
		application       marathon.Application
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc: "multiple ports with services",
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),

				withLabel(label.TraefikBackendMaxConnAmount, "1000"),
				withLabel(label.TraefikBackendMaxConnExtractorFunc, "client.ip"),
				withServiceLabel(label.TraefikPort, "80", "web"),
				withServiceLabel(label.TraefikPort, "81", "admin"),
				withLabel("traefik..port", "82"), // This should be ignored, as it fails to match the segmentPropertiesRegexp regex.
				withServiceLabel(label.TraefikFrontendRule, "Host:web.app.docker.localhost", "web"),
				withServiceLabel(label.TraefikFrontendRule, "Host:admin.app.docker.localhost", "admin"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-service-web": {
					Backend: "backend-app-service-web",
					Routes: map[string]types.Route{
						`route-host-app-service-web`: {
							Rule: "Host:web.app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
				"frontend-app-service-admin": {
					Backend: "backend-app-service-admin",
					Routes: map[string]types.Route{
						`route-host-app-service-admin`: {
							Rule: "Host:admin.app.docker.localhost",
						},
					},
					PassHostHeader: true,
					BasicAuth:      []string{},
					EntryPoints:    []string{},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-app-service-web": {
					Servers: map[string]types.Server{
						"server-task-service-web": {
							URL:    "http://localhost:80",
							Weight: 0,
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
				"backend-app-service-admin": {
					Servers: map[string]types.Server{
						"server-task-service-admin": {
							URL:    "http://localhost:81",
							Weight: 0,
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
			application: application(
				appPorts(80, 81),
				withTasks(localhostTask(taskPorts(80, 81))),

				// withLabel(label.TraefikBackend, "foobar"),

				withLabel(label.TraefikBackendCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
				withLabel(label.TraefikBackendHealthCheckPath, "/health"),
				withLabel(label.TraefikBackendHealthCheckPort, "880"),
				withLabel(label.TraefikBackendHealthCheckInterval, "6"),
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

				withServiceLabel(label.TraefikPort, "80", "containous"),
				withServiceLabel(label.TraefikProtocol, "https", "containous"),
				withServiceLabel(label.TraefikWeight, "12", "containous"),

				withServiceLabel(label.TraefikFrontendAuthBasic, "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", "containous"),
				withServiceLabel(label.TraefikFrontendEntryPoints, "http,https", "containous"),
				withServiceLabel(label.TraefikFrontendPassHostHeader, "true", "containous"),
				withServiceLabel(label.TraefikFrontendPassTLSCert, "true", "containous"),
				withServiceLabel(label.TraefikFrontendPriority, "666", "containous"),
				withServiceLabel(label.TraefikFrontendRedirectEntryPoint, "https", "containous"),
				withServiceLabel(label.TraefikFrontendRedirectRegex, "nope", "containous"),
				withServiceLabel(label.TraefikFrontendRedirectReplacement, "nope", "containous"),
				withServiceLabel(label.TraefikFrontendRedirectPermanent, "true", "containous"),
				withServiceLabel(label.TraefikFrontendRule, "Host:traefik.io", "containous"),
				withServiceLabel(label.TraefikFrontendWhiteListSourceRange, "10.10.10.10", "containous"),
				withServiceLabel(label.TraefikFrontendWhiteListUseXForwardedFor, "true", "containous"),

				withServiceLabel(label.TraefikFrontendRequestHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
				withServiceLabel(label.TraefikFrontendResponseHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
				withServiceLabel(label.TraefikFrontendSSLProxyHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8", "containous"),
				withServiceLabel(label.TraefikFrontendAllowedHosts, "foo,bar,bor", "containous"),
				withServiceLabel(label.TraefikFrontendHostsProxyHeaders, "foo,bar,bor", "containous"),
				withServiceLabel(label.TraefikFrontendSSLHost, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendCustomFrameOptionsValue, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendContentSecurityPolicy, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendPublicKey, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendReferrerPolicy, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendCustomBrowserXSSValue, "foo", "containous"),
				withServiceLabel(label.TraefikFrontendSTSSeconds, "666", "containous"),
				withServiceLabel(label.TraefikFrontendSSLRedirect, "true", "containous"),
				withServiceLabel(label.TraefikFrontendSSLTemporaryRedirect, "true", "containous"),
				withServiceLabel(label.TraefikFrontendSTSIncludeSubdomains, "true", "containous"),
				withServiceLabel(label.TraefikFrontendSTSPreload, "true", "containous"),
				withServiceLabel(label.TraefikFrontendForceSTSHeader, "true", "containous"),
				withServiceLabel(label.TraefikFrontendFrameDeny, "true", "containous"),
				withServiceLabel(label.TraefikFrontendContentTypeNosniff, "true", "containous"),
				withServiceLabel(label.TraefikFrontendBrowserXSSFilter, "true", "containous"),
				withServiceLabel(label.TraefikFrontendIsDevelopment, "true", "containous"),

				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageStatus, "404"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageBackend, "foobar"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"foo."+label.SuffixErrorPageQuery, "foo_query"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageStatus, "500,600"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageBackend, "foobar"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendErrorPage+"bar."+label.SuffixErrorPageQuery, "bar_query"),

				withServiceLabel(label.TraefikFrontendRateLimitExtractorFunc, "client.ip", "containous"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitPeriod, "6"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitAverage, "12"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"foo."+label.SuffixRateLimitBurst, "18"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitPeriod, "3"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitAverage, "6"),
				withLabel(label.Prefix+"containous."+label.BaseFrontendRateLimit+"bar."+label.SuffixRateLimitBurst, "9"),
			),
			expectedFrontends: map[string]*types.Frontend{
				"frontend-app-service-containous": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-app-service-containous",
					Routes: map[string]types.Route{
						"route-host-app-service-containous": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
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
							Backend: "foobar",
							Query:   "bar_query",
						},
						"foo": {
							Status: []string{
								"404",
							},
							Backend: "foobar",
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
						"server-task-service-containous": {
							URL:    "https://localhost:80",
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
						Path:     "/health",
						Port:     880,
						Interval: "6",
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

			test.application.ID = "/app"

			for _, task := range test.application.Tasks {
				task.ID = "task"
				if task.State == "" {
					task.State = "TASK_RUNNING"
				}
			}

			p := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
			}

			actualConfig := p.buildConfigurationV2(withApplications(test.application))

			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestApplicationFilterConstraints(t *testing.T) {
	testCases := []struct {
		desc                      string
		application               marathon.Application
		marathonLBCompatibility   bool
		filterMarathonConstraints bool
		expected                  bool
	}{
		{
			desc:                    "tags missing",
			application:             application(),
			marathonLBCompatibility: false,
			expected:                false,
		},
		{
			desc:                    "tag matching",
			application:             application(withLabel(label.TraefikTags, "valid")),
			marathonLBCompatibility: false,
			expected:                true,
		},
		{
			desc:                      "constraint missing",
			application:               application(),
			marathonLBCompatibility:   false,
			filterMarathonConstraints: true,
			expected:                  false,
		},
		{
			desc:                      "constraint invalid",
			application:               application(constraint("service_cluster:CLUSTER:test")),
			marathonLBCompatibility:   false,
			filterMarathonConstraints: true,
			expected:                  false,
		},
		{
			desc:                      "constraint valid",
			application:               application(constraint("valid")),
			marathonLBCompatibility:   false,
			filterMarathonConstraints: true,
			expected:                  true,
		},
		{
			desc: "LB compatibility tag matching",
			application: application(
				withLabel("HAPROXY_GROUP", "valid"),
				withLabel(label.TraefikTags, "notvalid"),
			),
			marathonLBCompatibility: true,
			expected:                true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				ExposedByDefault:          true,
				MarathonLBCompatibility:   test.marathonLBCompatibility,
				FilterMarathonConstraints: test.filterMarathonConstraints,
			}

			constraint, err := types.NewConstraint("tag==valid")
			if err != nil {
				t.Fatalf("failed to create constraint 'tag==valid': %v", err)
			}
			p.Constraints = types.Constraints{constraint}

			actual := p.applicationFilter(test.application)

			if actual != test.expected {
				t.Errorf("got %v, expected %v", actual, test.expected)
			}
		})
	}
}

func TestApplicationFilterEnabled(t *testing.T) {
	testCases := []struct {
		desc             string
		exposedByDefault bool
		enabledLabel     string
		expected         bool
	}{
		{
			desc:             "exposed",
			exposedByDefault: true,
			enabledLabel:     "",
			expected:         true,
		},
		{
			desc:             "exposed and tolerated by valid label value",
			exposedByDefault: true,
			enabledLabel:     "true",
			expected:         true,
		},
		{
			desc:             "exposed and tolerated by invalid label value",
			exposedByDefault: true,
			enabledLabel:     "invalid",
			expected:         true,
		},
		{
			desc:             "exposed but overridden by label",
			exposedByDefault: true,
			enabledLabel:     "false",
			expected:         false,
		},
		{
			desc:             "non-exposed",
			exposedByDefault: false,
			enabledLabel:     "",
			expected:         false,
		},
		{
			desc:             "non-exposed but overridden by label",
			exposedByDefault: false,
			enabledLabel:     "true",
			expected:         true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{ExposedByDefault: test.exposedByDefault}

			app := application(withLabel(label.TraefikEnable, test.enabledLabel))

			if provider.applicationFilter(app) != test.expected {
				t.Errorf("got unexpected filtering = %t", !test.expected)
			}
		})
	}
}

func TestTaskFilter(t *testing.T) {
	testCases := []struct {
		desc         string
		task         marathon.Task
		application  marathon.Application
		readyChecker *readinessChecker
		expected     bool
	}{
		{
			desc:        "missing port",
			task:        task(),
			application: application(),
			expected:    true,
		},
		{
			desc: "task not running",
			task: task(
				taskPorts(80),
				state(taskStateStaging),
			),
			application: application(appPorts(80)),
			expected:    false,
		},
		{
			desc:        "existing port",
			task:        task(taskPorts(80)),
			application: application(appPorts(80)),
			expected:    true,
		},
		{
			desc: "ambiguous port specification",
			task: task(taskPorts(80, 443)),
			application: application(
				appPorts(80, 443),
				withLabel(label.TraefikPort, "443"),
				withLabel(label.TraefikPortIndex, "1"),
			),
			expected: true,
		},
		{
			desc: "single service without port",
			task: task(taskPorts(80, 81)),
			application: application(
				appPorts(80, 81),
				withServiceLabel(label.TraefikPort, "80", "web"),
				withServiceLabel(label.TraefikPort, "illegal", "admin"),
			),
			expected: true,
		},
		{
			desc: "single service missing port",
			task: task(taskPorts(80, 81)),
			application: application(
				appPorts(80, 81),
				withServiceLabel(label.TraefikPort, "81", "admin"),
			),
			expected: true,
		},
		{
			desc: "readiness check false",
			task: task(taskPorts(80)),
			application: application(
				appPorts(80),
				deployments("deploymentId"),
				readinessCheck(0),
				readinessCheckResult(testTaskName, false),
			),
			readyChecker: testReadinessChecker(),
			expected:     false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{readyChecker: test.readyChecker}

			actual := p.taskFilter(test.task, test.application)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetSubDomain(t *testing.T) {
	testCases := []struct {
		path             string
		expected         string
		groupAsSubDomain bool
	}{
		{"/test", "test", false},
		{"/test", "test", true},
		{"/a/b/c/d", "d.c.b.a", true},
		{"/b/a/d/c", "c.d.a.b", true},
		{"/d/c/b/a", "a.b.c.d", true},
		{"/c/d/a/b", "b.a.d.c", true},
		{"/a/b/c/d", "a-b-c-d", false},
		{"/b/a/d/c", "b-a-d-c", false},
		{"/d/c/b/a", "d-c-b-a", false},
		{"/c/d/a/b", "c-d-a-b", false},
	}

	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("path=%s,group=%t", test.path, test.groupAsSubDomain), func(t *testing.T) {
			t.Parallel()

			p := &Provider{GroupsAsSubDomains: test.groupAsSubDomain}

			actual := p.getSubDomain(test.path)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetPort(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		task        marathon.Task
		serviceName string
		expected    string
	}{
		{
			desc:        "port missing",
			application: application(),
			task:        task(),
			expected:    "",
		},
		{
			desc:        "numeric port",
			application: application(withLabel(label.TraefikPort, "80")),
			task:        task(),
			expected:    "80",
		},
		{
			desc:        "string port",
			application: application(withLabel(label.TraefikPort, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc:        "negative port",
			application: application(withLabel(label.TraefikPort, "-1")),
			task:        task(taskPorts(80)),
			expected:    "",
		},
		{
			desc:        "task port available",
			application: application(),
			task:        task(taskPorts(80)),
			expected:    "80",
		},
		{
			desc: "port definition available",
			application: application(
				portDefinition(443),
			),
			task:     task(),
			expected: "443",
		},
		{
			desc:        "IP-per-task port available",
			application: application(ipAddrPerTask(8000)),
			task:        task(),
			expected:    "8000",
		},
		{
			desc:        "multiple task ports available",
			application: application(),
			task:        task(taskPorts(80, 443)),
			expected:    "80",
		},
		{
			desc:        "numeric port index specified",
			application: application(withLabel(label.TraefikPortIndex, "1")),
			task:        task(taskPorts(80, 443)),
			expected:    "443",
		},
		{
			desc:        "string port index specified",
			application: application(withLabel(label.TraefikPortIndex, "foobar")),
			task:        task(taskPorts(80)),
			expected:    "80",
		},
		{
			desc: "port and port index specified",
			application: application(
				withLabel(label.TraefikPort, "80"),
				withLabel(label.TraefikPortIndex, "1"),
			),
			task:     task(taskPorts(80, 443)),
			expected: "80",
		},
		{
			desc:        "task and application ports specified",
			application: application(appPorts(9999)),
			task:        task(taskPorts(7777)),
			expected:    "7777",
		},
		{
			desc:        "multiple task ports with service index available",
			application: application(withLabel(label.Prefix+"http.portIndex", "0")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
		{
			desc:        "multiple task ports with service port available",
			application: application(withLabel(label.Prefix+"https.port", "443")),
			task:        task(taskPorts(80, 443)),
			serviceName: "https",
			expected:    "443",
		},
		{
			desc:        "multiple task ports with services but default port available",
			application: application(withLabel(label.Prefix+"http.weight", "100")),
			task:        task(taskPorts(80, 443)),
			serviceName: "http",
			expected:    "80",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getPortV1(test.task, test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetFrontendRule(t *testing.T) {
	testCases := []struct {
		desc                    string
		application             marathon.Application
		serviceName             string
		expected                string
		marathonLBCompatibility bool
	}{
		{
			desc:                    "label missing",
			application:             application(appID("test")),
			marathonLBCompatibility: true,
			expected:                "Host:test.docker.localhost",
		},
		{
			desc: "HAProxy vhost available and LB compat disabled",
			application: application(
				appID("test"),
				withLabel("HAPROXY_0_VHOST", "foo.bar"),
			),
			marathonLBCompatibility: false,
			expected:                "Host:test.docker.localhost",
		},
		{
			desc:                    "HAProxy vhost available and LB compat enabled",
			application:             application(withLabel("HAPROXY_0_VHOST", "foo.bar")),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc: "frontend rule available",

			application: application(
				withLabel(label.TraefikFrontendRule, "Host:foo.bar"),
				withLabel("HAPROXY_0_VHOST", "unused"),
			),
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
		{
			desc:                    "service label existing",
			application:             application(withServiceLabel(label.TraefikFrontendRule, "Host:foo.bar", "app")),
			serviceName:             "app",
			marathonLBCompatibility: true,
			expected:                "Host:foo.bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p := &Provider{
				Domain:                  "docker.localhost",
				MarathonLBCompatibility: test.marathonLBCompatibility,
			}

			actual := p.getFrontendRuleV1(test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendName(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    string
	}{
		{
			desc:        "label missing",
			application: application(appID("/group/app")),
			expected:    "backend-group-app",
		},
		{
			desc:        "label existing",
			application: application(withLabel(label.TraefikBackend, "bar")),
			expected:    "backendbar",
		},
		{
			desc:        "service label existing",
			application: application(withServiceLabel(label.TraefikBackend, "bar", "app")),
			serviceName: "app",
			expected:    "backendbar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{}

			actual := p.getBackendNameV1(test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetServers(t *testing.T) {
	testCases := []struct {
		desc        string
		application marathon.Application
		serviceName string
		expected    map[string]types.Server
	}{
		{
			desc:        "should return nil when no task",
			application: application(ipAddrPerTask(80)),
			expected:    nil,
		},
		{
			desc: "should return nil when all hosts are empty",
			application: application(
				withTasks(
					task(ipAddresses("1.1.1.1"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), withTaskID("B"), taskPorts(80)),
					task(ipAddresses("1.1.1.3"), withTaskID("C"), taskPorts(80))),
			),
			expected: nil,
		},
		{
			desc: "with 3 tasks",
			application: application(
				ipAddrPerTask(80),
				withTasks(
					task(ipAddresses("1.1.1.1"), withTaskID("A"), taskPorts(80)),
					task(ipAddresses("1.1.1.2"), withTaskID("B"), taskPorts(80)),
					task(ipAddresses("1.1.1.3"), withTaskID("C"), taskPorts(80))),
			),
			expected: map[string]types.Server{
				"server-A": {
					URL:    "http://1.1.1.1:80",
					Weight: 0,
				},
				"server-B": {
					URL:    "http://1.1.1.2:80",
					Weight: 0,
				},
				"server-C": {
					URL:    "http://1.1.1.3:80",
					Weight: 0,
				},
			},
		},
	}

	p := &Provider{}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := p.getServersV1(test.application, test.serviceName)

			assert.Equal(t, test.expected, actual)
		})
	}
}
