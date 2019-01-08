package label

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestParseErrorPages(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected map[string]*types.ErrorPage
	}{
		{
			desc: "2 errors pages",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageStatus:  "404",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageBackend: "foo_backend",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageQuery:   "foo_query",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageStatus:  "500,600",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageBackend: "bar_backend",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageQuery:   "bar_query",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status:  []string{"404"},
					Query:   "foo_query",
					Backend: "foo_backend",
				},
				"bar": {
					Status:  []string{"500", "600"},
					Query:   "bar_query",
					Backend: "bar_backend",
				},
			},
		},
		{
			desc: "only status field",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageStatus: "404",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status: []string{"404"},
				},
			},
		},
		{
			desc: "invalid field",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + "courgette": "404",
			},
			expected: map[string]*types.ErrorPage{"foo": {}},
		},
		{
			desc:     "no error pages labels",
			labels:   map[string]string{},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pages := ParseErrorPages(test.labels, Prefix+BaseFrontendErrorPage, RegexpFrontendErrorPage)

			assert.EqualValues(t, test.expected, pages)
		})
	}
}

func TestParseRateSets(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected map[string]*types.Rate
	}{
		{
			desc: "2 rate limits",
			labels: map[string]string{
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitPeriod:  "6",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitAverage: "12",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitBurst:   "18",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitPeriod:  "3",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitAverage: "6",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitBurst:   "9",
			},
			expected: map[string]*types.Rate{
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
		{
			desc:     "no rate limits labels",
			labels:   map[string]string{},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rateSets := ParseRateSets(test.labels, Prefix+BaseFrontendRateLimit, RegexpFrontendRateLimit)

			assert.EqualValues(t, test.expected, rateSets)
		})
	}
}

func TestWhiteList(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.WhiteList
	}{
		{
			desc:     "should return nil when no white list labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when deprecated label",
			labels: map[string]string{
				TraefikFrontendWhitelistSourceRange: "10.10.10.10",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: false,
			},
		},
		{
			desc: "should return a struct when only range",
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange: "10.10.10.10",
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
			labels: map[string]string{
				TraefikFrontendWhiteListSourceRange:      "10.10.10.10",
				TraefikFrontendWhiteListUseXForwardedFor: "true",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: true,
			},
		},
		{
			desc: "should return a struct when mix deprecated label and new labels",
			labels: map[string]string{
				TraefikFrontendWhitelistSourceRange:      "20.20.20.20",
				TraefikFrontendWhiteListSourceRange:      "10.10.10.10",
				TraefikFrontendWhiteListUseXForwardedFor: "true",
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
			labels: map[string]string{
				TraefikFrontendWhiteListUseXForwardedFor: "true",
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetWhiteList(test.labels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetCircuitBreaker(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.CircuitBreaker
	}{
		{
			desc:     "should return nil when no CB label",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when CB label is set",
			labels: map[string]string{
				TraefikBackendCircuitBreakerExpression: "NetworkErrorRatio() > 0.5",
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

			actual := GetCircuitBreaker(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetLoadBalancer(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.LoadBalancer
	}{
		{
			desc:     "should return nil when no LB labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when labels are set",
			labels: map[string]string{
				TraefikBackendLoadBalancerMethod:               "drr",
				TraefikBackendLoadBalancerSticky:               "true",
				TraefikBackendLoadBalancerStickiness:           "true",
				TraefikBackendLoadBalancerStickinessCookieName: "foo",
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
			labels: map[string]string{
				TraefikBackendLoadBalancerMethod:               "drr",
				TraefikBackendLoadBalancerSticky:               "true",
				TraefikBackendLoadBalancerStickinessCookieName: "foo",
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

			actual := GetLoadBalancer(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetMaxConn(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.MaxConn
	}{
		{
			desc:     "should return nil when no max conn labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return nil when no amount label",
			labels: map[string]string{
				TraefikBackendMaxConnExtractorFunc: "client.ip",
			},
			expected: nil,
		},
		{
			desc: "should return default when no empty extractorFunc label",
			labels: map[string]string{
				TraefikBackendMaxConnExtractorFunc: "",
				TraefikBackendMaxConnAmount:        "666",
			},
			expected: &types.MaxConn{
				ExtractorFunc: "request.host",
				Amount:        666,
			},
		},
		{
			desc: "should return a struct when max conn labels are set",
			labels: map[string]string{
				TraefikBackendMaxConnExtractorFunc: "client.ip",
				TraefikBackendMaxConnAmount:        "666",
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

			actual := GetMaxConn(test.labels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetHealthCheck(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.HealthCheck
	}{
		{
			desc:     "should return nil when no health check labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return nil when no health check Path label",
			labels: map[string]string{
				TraefikBackendHealthCheckPort:     "80",
				TraefikBackendHealthCheckInterval: "6",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when health check labels are set",
			labels: map[string]string{
				TraefikBackendHealthCheckPath:     "/health",
				TraefikBackendHealthCheckPort:     "80",
				TraefikBackendHealthCheckInterval: "6",
				TraefikBackendHealthCheckHeaders:  "Foo:bar || Goo:bir",
				TraefikBackendHealthCheckHostname: "traefik",
				TraefikBackendHealthCheckScheme:   "http",
			},
			expected: &types.HealthCheck{
				Scheme:   "http",
				Path:     "/health",
				Port:     80,
				Interval: "6",
				Hostname: "traefik",
				Headers: map[string]string{
					"Foo": "bar",
					"Goo": "bir",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetHealthCheck(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBuffering(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.Buffering
	}{
		{
			desc:     "should return nil when no buffering labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when buffering labels are set",
			labels: map[string]string{
				TraefikBackendBufferingMaxResponseBodyBytes: "10485760",
				TraefikBackendBufferingMemResponseBodyBytes: "2097152",
				TraefikBackendBufferingMaxRequestBodyBytes:  "10485760",
				TraefikBackendBufferingMemRequestBodyBytes:  "2097152",
				TraefikBackendBufferingRetryExpression:      "IsNetworkError() && Attempts() <= 2",
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

			actual := GetBuffering(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.Redirect
	}{

		{
			desc:     "should return nil when no redirect labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should use only entry point tag when mix regex redirect and entry point redirect",
			labels: map[string]string{
				TraefikFrontendRedirectEntryPoint:  "https",
				TraefikFrontendRedirectRegex:       "(.*)",
				TraefikFrontendRedirectReplacement: "$1",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label",
			labels: map[string]string{
				TraefikFrontendRedirectEntryPoint: "https",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label (permanent)",
			labels: map[string]string{
				TraefikFrontendRedirectEntryPoint: "https",
				TraefikFrontendRedirectPermanent:  "true",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
				Permanent:  true,
			},
		},
		{
			desc: "should return a struct when regex redirect labels",
			labels: map[string]string{
				TraefikFrontendRedirectRegex:       "(.*)",
				TraefikFrontendRedirectReplacement: "$1",
			},
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
		{
			desc: "should return a struct when regex redirect labels (permanent)",
			labels: map[string]string{
				TraefikFrontendRedirectRegex:       "(.*)",
				TraefikFrontendRedirectReplacement: "$1",
				TraefikFrontendRedirectPermanent:   "true",
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

			actual := GetRedirect(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRateLimit(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.RateLimit
	}{
		{
			desc:     "should return nil when no rate limit labels",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when rate limit labels are defined",
			labels: map[string]string{
				TraefikFrontendRateLimitExtractorFunc:                            "client.ip",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitPeriod:  "6",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitAverage: "12",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitBurst:   "18",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitPeriod:  "3",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitAverage: "6",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitBurst:   "9",
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
			labels: map[string]string{
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitPeriod:  "6",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitAverage: "12",
				Prefix + BaseFrontendRateLimit + "foo." + SuffixRateLimitBurst:   "18",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitPeriod:  "3",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitAverage: "6",
				Prefix + BaseFrontendRateLimit + "bar." + SuffixRateLimitBurst:   "9",
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetRateLimit(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetHeaders(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.Headers
	}{
		{
			desc:     "should return nil when no custom headers options are set",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a struct when all custom headers options are set",
			labels: map[string]string{
				TraefikFrontendRequestHeaders:          "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				TraefikFrontendResponseHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				TraefikFrontendSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				TraefikFrontendAllowedHosts:            "foo,bar,bor",
				TraefikFrontendHostsProxyHeaders:       "foo,bar,bor",
				TraefikFrontendSSLHost:                 "foo",
				TraefikFrontendCustomFrameOptionsValue: "foo",
				TraefikFrontendContentSecurityPolicy:   "foo",
				TraefikFrontendPublicKey:               "foo",
				TraefikFrontendReferrerPolicy:          "foo",
				TraefikFrontendCustomBrowserXSSValue:   "foo",
				TraefikFrontendSTSSeconds:              "666",
				TraefikFrontendSSLRedirect:             "true",
				TraefikFrontendSSLForceHost:            "true",
				TraefikFrontendSSLTemporaryRedirect:    "true",
				TraefikFrontendSTSIncludeSubdomains:    "true",
				TraefikFrontendSTSPreload:              "true",
				TraefikFrontendForceSTSHeader:          "true",
				TraefikFrontendFrameDeny:               "true",
				TraefikFrontendContentTypeNosniff:      "true",
				TraefikFrontendBrowserXSSFilter:        "true",
				TraefikFrontendIsDevelopment:           "true",
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
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := GetHeaders(test.labels)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetErrorPages(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected map[string]*types.ErrorPage
	}{
		{
			desc:     "should return nil when no tags",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return a map when tags are present",
			labels: map[string]string{
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageStatus:  "404",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageBackend: "foo_backend",
				Prefix + BaseFrontendErrorPage + "foo." + SuffixErrorPageQuery:   "foo_query",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageStatus:  "500,600",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageBackend: "bar_backend",
				Prefix + BaseFrontendErrorPage + "bar." + SuffixErrorPageQuery:   "bar_query",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status:  []string{"404"},
					Query:   "foo_query",
					Backend: "foo_backend",
				},
				"bar": {
					Status:  []string{"500", "600"},
					Query:   "bar_query",
					Backend: "bar_backend",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := GetErrorPages(test.labels)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetAuth(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.Auth
	}{
		{
			desc:     "should return nil when no tags",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return nil when no real auth",
			labels: map[string]string{
				TraefikFrontendAuthHeaderField: "myHeaderField",
			},
			expected: nil,
		},
		{
			desc: "should return a basic auth",
			labels: map[string]string{
				TraefikFrontendAuthHeaderField:       "myHeaderField",
				TraefikFrontendAuthBasicUsers:        "user:pwd,user2:pwd2",
				TraefikFrontendAuthBasicUsersFile:    "myUsersFile",
				TraefikFrontendAuthBasicRemoveHeader: "true",
			},
			expected: &types.Auth{
				HeaderField: "myHeaderField",
				Basic:       &types.Basic{UsersFile: "myUsersFile", Users: []string{"user:pwd", "user2:pwd2"}, RemoveHeader: true},
			},
		},
		{
			desc: "should return a digest auth",
			labels: map[string]string{
				TraefikFrontendAuthDigestRemoveHeader: "true",
				TraefikFrontendAuthHeaderField:        "myHeaderField",
				TraefikFrontendAuthDigestUsers:        "user:pwd,user2:pwd2",
				TraefikFrontendAuthDigestUsersFile:    "myUsersFile",
			},
			expected: &types.Auth{
				HeaderField: "myHeaderField",
				Digest:      &types.Digest{UsersFile: "myUsersFile", Users: []string{"user:pwd", "user2:pwd2"}, RemoveHeader: true},
			},
		},
		{
			desc: "should return a forward auth",
			labels: map[string]string{
				TraefikFrontendAuthHeaderField:                  "myHeaderField",
				TraefikFrontendAuthForwardAddress:               "myAddress",
				TraefikFrontendAuthForwardTrustForwardHeader:    "true",
				TraefikFrontendAuthForwardTLSCa:                 "ca.crt",
				TraefikFrontendAuthForwardTLSCaOptional:         "true",
				TraefikFrontendAuthForwardTLSInsecureSkipVerify: "true",
				TraefikFrontendAuthForwardTLSKey:                "myKey",
				TraefikFrontendAuthForwardTLSCert:               "myCert",
			},
			expected: &types.Auth{
				HeaderField: "myHeaderField",
				Forward: &types.Forward{
					TrustForwardHeader: true,
					Address:            "myAddress",
					TLS: &types.ClientTLS{
						InsecureSkipVerify: true,
						CA:                 "ca.crt",
						CAOptional:         true,
						Key:                "myKey",
						Cert:               "myCert",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := GetAuth(test.labels)

			assert.Equal(t, test.expected, result)
		})
	}
}
func TestGetPassTLSClientCert(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected *types.TLSClientHeaders
	}{
		{
			desc:     "should return nil when no tags",
			labels:   map[string]string{},
			expected: nil,
		},
		{
			desc: "should return tlsClientHeaders with true pem flag",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertPem: "true",
			},
			expected: &types.TLSClientHeaders{
				PEM: true,
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and NotAfter true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosNotAfter: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotAfter: true,
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and NotBefore true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosNotBefore: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					NotBefore: true,
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and sans true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSans: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Sans: true,
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with commonName true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectCommonName: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						CommonName: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with country true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectCountry: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						Country: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with locality true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectLocality: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						Locality: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with organization true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectOrganization: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						Organization: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with province true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectProvince: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						Province: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with infos and subject with serialNumber true",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber: "true",
			},
			expected: &types.TLSClientHeaders{
				Infos: &types.TLSClientCertificateInfos{
					Subject: &types.TLSCLientCertificateDNInfos{
						SerialNumber: true,
					},
				},
			},
		},
		{
			desc: "should return tlsClientHeaders with all infos",
			labels: map[string]string{
				TraefikFrontendPassTLSClientCertPem:                         "true",
				TraefikFrontendPassTLSClientCertInfosNotAfter:               "true",
				TraefikFrontendPassTLSClientCertInfosNotBefore:              "true",
				TraefikFrontendPassTLSClientCertInfosSans:                   "true",
				TraefikFrontendPassTLSClientCertInfosIssuerCommonName:       "true",
				TraefikFrontendPassTLSClientCertInfosIssuerCountry:          "true",
				TraefikFrontendPassTLSClientCertInfosIssuerDomainComponent:  "true",
				TraefikFrontendPassTLSClientCertInfosIssuerLocality:         "true",
				TraefikFrontendPassTLSClientCertInfosIssuerOrganization:     "true",
				TraefikFrontendPassTLSClientCertInfosIssuerProvince:         "true",
				TraefikFrontendPassTLSClientCertInfosIssuerSerialNumber:     "true",
				TraefikFrontendPassTLSClientCertInfosSubjectCommonName:      "true",
				TraefikFrontendPassTLSClientCertInfosSubjectCountry:         "true",
				TraefikFrontendPassTLSClientCertInfosSubjectDomainComponent: "true",
				TraefikFrontendPassTLSClientCertInfosSubjectLocality:        "true",
				TraefikFrontendPassTLSClientCertInfosSubjectOrganization:    "true",
				TraefikFrontendPassTLSClientCertInfosSubjectProvince:        "true",
				TraefikFrontendPassTLSClientCertInfosSubjectSerialNumber:    "true",
			},
			expected: &types.TLSClientHeaders{
				PEM: true,
				Infos: &types.TLSClientCertificateInfos{
					Sans:      true,
					NotBefore: true,
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
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := GetTLSClientCert(test.labels)

			assert.Equal(t, test.expected, result)
		})
	}
}
