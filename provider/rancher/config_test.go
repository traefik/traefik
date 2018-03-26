package rancher

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderBuildConfiguration(t *testing.T) {
	provider := &Provider{
		Domain:           "rancher.localhost",
		ExposedByDefault: true,
	}

	testCases := []struct {
		desc              string
		services          []rancherData
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "without services",
			services:          []rancherData{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			services: []rancherData{
				{
					Labels: map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendHealthCheckPath:                  "/health",
						label.TraefikBackendHealthCheckPort:                  "880",
						label.TraefikBackendHealthCheckInterval:              "6",
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

						label.TraefikFrontendAuthBasic:                 "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
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
					Health:     "healthy",
					Containers: []string{"10.0.0.1", "10.0.0.2"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-traefik-io": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-traefik-io": {
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
						SourceRange: []string{
							"10.10.10.10",
						},
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
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "https://10.0.0.1:666",
							Weight: 12,
						},
						"server-1": {
							URL:    "https://10.0.0.2:666",
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
		{
			desc: "with services",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:                       "80",
						label.TraefikFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendRedirectEntryPoint: "https",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Priority:       0,
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
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

			actualConfig := provider.buildConfiguration(test.services)
			require.NotNil(t, actualConfig)

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderServiceFilter(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
		EnableServiceHealthFilter: true,
	}

	constraint, _ := types.NewConstraint("tag==ch*se")
	provider.Constraints = types.Constraints{constraint}

	testCases := []struct {
		desc     string
		service  rancherData
		expected bool
	}{
		{
			desc: "missing Port labels, don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "false",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "unhealthy",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "unhealthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "inactive",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "not-cheesy",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "inactive",
			},
			expected: false,
		},
		{
			desc: "healthy & active, tag: cheese",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & active, tag: chose",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "chose",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & upgraded",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheeeeese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "upgraded",
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.serviceFilter(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestContainerFilter(t *testing.T) {
	testCases := []struct {
		name        string
		healthState string
		state       string
		expected    bool
	}{
		{
			healthState: "unhealthy",
			state:       "running",
			expected:    false,
		},
		{
			healthState: "healthy",
			state:       "stopped",
			expected:    false,
		},
		{
			state:    "stopped",
			expected: false,
		},
		{
			healthState: "healthy",
			state:       "running",
			expected:    true,
		},
		{
			healthState: "updating-healthy",
			state:       "updating-running",
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.healthState+" "+test.state, func(t *testing.T) {
			t.Parallel()

			actual := containerFilter(test.name, test.healthState, test.state)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendName(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "default",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host-foo-rancher-localhost",
		},
		{
			desc: "with Headers label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar",
				},
			},
			expected: "Host-foo-bar",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Path:/test",
				},
			},
			expected: "Path-test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				},
			},
			expected: "PathPrefix-test2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getFrontendName(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendRule(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "host",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host:foo.rancher.localhost",
		},
		{
			desc: "host with /",
			service: rancherData{
				Name: "foo/bar",
			},
			expected: "Host:foo.bar.rancher.localhost",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar.com",
				},
			},
			expected: "Host:foo.bar.com",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Path:/test",
				},
			},
			expected: "Path:/test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				},
			},
			expected: "PathPrefix:/test2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getFrontendRule(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendName(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "without label",
			service: rancherData{
				Name: "test-service",
			},
			expected: "test-service",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikBackend: "foobar",
				},
			},

			expected: "foobar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getBackendName(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetCircuitBreaker(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.CircuitBreaker
	}{
		{
			desc: "should return nil when no CB label",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when CB label is set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendCircuitBreakerExpression: "NetworkErrorRatio() > 0.5",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.CircuitBreaker{
				Expression: "NetworkErrorRatio() > 0.5",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getCircuitBreaker(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetLoadBalancer(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.LoadBalancer
	}{
		{
			desc: "should return nil when no LB labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when labels are set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendLoadBalancerMethod:               "drr",
					label.TraefikBackendLoadBalancerSticky:               "true",
					label.TraefikBackendLoadBalancerStickiness:           "true",
					label.TraefikBackendLoadBalancerStickinessCookieName: "foo",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.LoadBalancer{
				Method: "drr",
				Sticky: true,
				Stickiness: &types.Stickiness{
					CookieName: "foo",
				},
			},
		},
		{
			desc: "should return a nil Stickiness when Stickiness is not set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendLoadBalancerMethod:               "drr",
					label.TraefikBackendLoadBalancerSticky:               "true",
					label.TraefikBackendLoadBalancerStickinessCookieName: "foo",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.LoadBalancer{
				Method:     "drr",
				Sticky:     true,
				Stickiness: nil,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getLoadBalancer(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetMaxConn(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.MaxConn
	}{
		{
			desc: "should return nil when no max conn labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return nil when no amount label",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendMaxConnExtractorFunc: "client.ip",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return default when no empty extractorFunc label",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendMaxConnExtractorFunc: "",
					label.TraefikBackendMaxConnAmount:        "666",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.MaxConn{
				ExtractorFunc: "request.host",
				Amount:        666,
			},
		},
		{
			desc: "should return a struct when max conn labels are set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendMaxConnExtractorFunc: "client.ip",
					label.TraefikBackendMaxConnAmount:        "666",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.MaxConn{
				ExtractorFunc: "client.ip",
				Amount:        666,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getMaxConn(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetHealthCheck(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.HealthCheck
	}{
		{
			desc: "should return nil when no health check labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return nil when no health check Path label",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendHealthCheckPort:     "80",
					label.TraefikBackendHealthCheckInterval: "6",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when health check labels are set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendHealthCheckPath:     "/health",
					label.TraefikBackendHealthCheckPort:     "80",
					label.TraefikBackendHealthCheckInterval: "6",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.HealthCheck{
				Path:     "/health",
				Port:     80,
				Interval: "6",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getHealthCheck(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBuffering(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.Buffering
	}{
		{
			desc: "should return nil when no buffering labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when buffering labels are set",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikBackendBufferingMaxResponseBodyBytes: "10485760",
					label.TraefikBackendBufferingMemResponseBodyBytes: "2097152",
					label.TraefikBackendBufferingMaxRequestBodyBytes:  "10485760",
					label.TraefikBackendBufferingMemRequestBodyBytes:  "2097152",
					label.TraefikBackendBufferingRetryExpression:      "IsNetworkError() && Attempts() <= 2",
				},
				Health: "healthy",
				State:  "active",
			},
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

			actual := getBuffering(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetServers(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected map[string]types.Server
	}{
		{
			desc: "should return nil when no server labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return nil when no server IPs",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "7",
				},
				Containers: []string{},
				Health:     "healthy",
				State:      "active",
			},
			expected: nil,
		},
		{
			desc: "should use default weight when invalid weight value",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "kls",
				},
				Containers: []string{"10.10.10.0"},
				Health:     "healthy",
				State:      "active",
			},
			expected: map[string]types.Server{
				"server-0": {
					URL:    "http://10.10.10.0:",
					Weight: 0,
				},
			},
		},
		{
			desc: "should return a map when configuration keys are defined",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikWeight: "6",
				},
				Containers: []string{"10.10.10.0", "10.10.10.1"},
				Health:     "healthy",
				State:      "active",
			},
			expected: map[string]types.Server{
				"server-0": {
					URL:    "http://10.10.10.0:",
					Weight: 6,
				},
				"server-1": {
					URL:    "http://10.10.10.1:",
					Weight: 6,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getServers(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestWhiteList(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.WhiteList
	}{
		{
			desc: "should return nil when no white list labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when only range",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendWhiteListSourceRange: "10.10.10.10",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: false,
			},
		},
		{
			desc: "should return a struct when range and UseXForwardedFor",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendWhiteListSourceRange:      "10.10.10.10",
					label.TraefikFrontendWhiteListUseXForwardedFor: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: true,
			},
		},
		{
			desc: "should return nil when only UseXForwardedFor",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendWhiteListUseXForwardedFor: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getWhiteList(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.Redirect
	}{

		{
			desc: "should return nil when no redirect labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should use only entry point tag when mix regex redirect and entry point redirect",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRedirectEntryPoint:  "https",
					label.TraefikFrontendRedirectRegex:       "(.*)",
					label.TraefikFrontendRedirectReplacement: "$1",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRedirectEntryPoint: "https",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label (permanent)",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRedirectEntryPoint: "https",
					label.TraefikFrontendRedirectPermanent:  "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
				Permanent:  true,
			},
		},
		{
			desc: "should return a struct when regex redirect labels",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRedirectRegex:       "(.*)",
					label.TraefikFrontendRedirectReplacement: "$1",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
		{
			desc: "should return a struct when regex redirect labels (permanent)",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRedirectRegex:       "(.*)",
					label.TraefikFrontendRedirectReplacement: "$1",
					label.TraefikFrontendRedirectPermanent:   "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
				Permanent:   true,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getRedirect(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRateLimit(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.RateLimit
	}{
		{
			desc: "should return nil when no rate limit labels",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when rate limit labels are defined",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikFrontendRateLimitExtractorFunc:                                        "client.ip",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.RateLimit{
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
		},
		{
			desc: "should return nil when ExtractorFunc is missing",
			service: rancherData{
				Labels: map[string]string{
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getRateLimit(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetHeaders(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected *types.Headers
	}{
		{
			desc: "should return nil when no custom headers options are set",
			service: rancherData{
				Labels: map[string]string{},
				Health: "healthy",
				State:  "active",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when all custom headers options are set",
			service: rancherData{
				Labels: map[string]string{
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
					label.TraefikFrontendSSLRedirect:             "true",
					label.TraefikFrontendSSLTemporaryRedirect:    "true",
					label.TraefikFrontendSTSIncludeSubdomains:    "true",
					label.TraefikFrontendSTSPreload:              "true",
					label.TraefikFrontendForceSTSHeader:          "true",
					label.TraefikFrontendFrameDeny:               "true",
					label.TraefikFrontendContentTypeNosniff:      "true",
					label.TraefikFrontendBrowserXSSFilter:        "true",
					label.TraefikFrontendIsDevelopment:           "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: &types.Headers{
				CustomRequestHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				CustomResponseHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				SSLProxyHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				AllowedHosts:            []string{"foo", "bar", "bor"},
				HostsProxyHeaders:       []string{"foo", "bar", "bor"},
				SSLHost:                 "foo",
				CustomFrameOptionsValue: "foo",
				ContentSecurityPolicy:   "foo",
				PublicKey:               "foo",
				ReferrerPolicy:          "foo",
				CustomBrowserXSSValue:   "foo",
				STSSeconds:              666,
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
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getHeaders(test.service)

			assert.Equal(t, test.expected, actual)
		})
	}
}
