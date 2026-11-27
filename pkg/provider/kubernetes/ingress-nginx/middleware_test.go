package ingressnginx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	netv1 "k8s.io/api/networking/v1"
)

func TestBuildFromToWwwRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		hostname string
		enabled  bool
		allHosts map[string]bool
		expected *middlewareFromToWwwRedirect
	}{
		{
			desc:     "annotation not set",
			hostname: "example.com",
		},
		{
			desc:     "default backend has no hostname",
			hostname: "",
			enabled:  true,
		},
		{
			desc:     "non-www host redirects to www",
			hostname: "example.com",
			enabled:  true,
			expected: &middlewareFromToWwwRedirect{
				ExtraRouterRule: `Host("www.example.com")`,
				TargetHostname:  "example.com",
			},
		},
		{
			desc:     "www host redirects to www",
			hostname: "www.example.com",
			enabled:  true,
			expected: &middlewareFromToWwwRedirect{
				ExtraRouterRule: `Host("example.com")`,
				TargetHostname:  "www.example.com",
			},
		},
		{
			desc:     "www counterpart already served",
			hostname: "example.com",
			enabled:  true,
			allHosts: map[string]bool{"www.example.com": true},
		},
		{
			desc:     "non-www counterpart already served",
			hostname: "www.example.com",
			enabled:  true,
			allHosts: map[string]bool{"example.com": true},
		},
		{
			desc:     "wildcard host",
			hostname: "*.example.com",
			enabled:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loc := &location{Config: IngressConfig{}}
			if test.enabled {
				loc.Config.FromToWwwRedirect = new(true)
			}

			p := &Provider{}
			p.buildFromToWwwRedirect(loc, test.hostname, test.allHosts)

			assert.Equal(t, test.expected, loc.FromToWwwRedirect)
		})
	}
}

func TestBuildLimits(t *testing.T) {
	testCases := []struct {
		desc                     string
		limitIPStrategy          *dynamic.IPStrategy
		expectedRateLimitRPS     *dynamic.RateLimit
		expectedRateLimitRPM     *dynamic.RateLimit
		expectedLimitConnections *dynamic.InFlightReq
	}{
		{
			desc: "without limit IP strategy",
			expectedRateLimitRPS: &dynamic.RateLimit{
				Average: 10,
				Period:  ptypes.Duration(time.Second),
				Burst:   50,
			},
			expectedRateLimitRPM: &dynamic.RateLimit{
				Average: 120,
				Period:  ptypes.Duration(time.Minute),
				Burst:   600,
			},
			expectedLimitConnections: &dynamic.InFlightReq{
				Amount: 5,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{},
				},
			},
		},
		{
			desc:            "with limit IP strategy",
			limitIPStrategy: &dynamic.IPStrategy{Depth: 1},
			expectedRateLimitRPS: &dynamic.RateLimit{
				Average: 10,
				Period:  ptypes.Duration(time.Second),
				Burst:   50,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{Depth: 1},
				},
			},
			expectedRateLimitRPM: &dynamic.RateLimit{
				Average: 120,
				Period:  ptypes.Duration(time.Minute),
				Burst:   600,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{Depth: 1},
				},
			},
			expectedLimitConnections: &dynamic.InFlightReq{
				Amount: 5,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{Depth: 1},
				},
			},
		},
		{
			desc:            "with excluded IPs limit IP strategy",
			limitIPStrategy: &dynamic.IPStrategy{ExcludedIPs: []string{"173.245.48.0/20"}},
			expectedRateLimitRPS: &dynamic.RateLimit{
				Average: 10,
				Period:  ptypes.Duration(time.Second),
				Burst:   50,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{ExcludedIPs: []string{"173.245.48.0/20"}},
				},
			},
			expectedRateLimitRPM: &dynamic.RateLimit{
				Average: 120,
				Period:  ptypes.Duration(time.Minute),
				Burst:   600,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{ExcludedIPs: []string{"173.245.48.0/20"}},
				},
			},
			expectedLimitConnections: &dynamic.InFlightReq{
				Amount: 5,
				SourceCriterion: &dynamic.SourceCriterion{
					IPStrategy: &dynamic.IPStrategy{ExcludedIPs: []string{"173.245.48.0/20"}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var ing netv1.Ingress
			ing.SetAnnotations(map[string]string{
				"nginx.ingress.kubernetes.io/limit-rps":         "10",
				"nginx.ingress.kubernetes.io/limit-rpm":         "120",
				"nginx.ingress.kubernetes.io/limit-connections": "5",
			})

			loc := &location{Config: parseIngressConfig(&ing)}

			p := &Provider{LimitIPStrategy: test.limitIPStrategy}
			p.buildRateLimit(loc)
			p.buildLimitConnections(loc)

			assert.Equal(t, test.expectedRateLimitRPS, loc.RateLimitRPS)
			assert.Equal(t, test.expectedRateLimitRPM, loc.RateLimitRPM)
			assert.Equal(t, test.expectedLimitConnections, loc.LimitConnections)
		})
	}
}

func TestBuildCORS(t *testing.T) {
	testCases := []struct {
		desc              string
		allowOrigin       string
		expectedList      []string
		expectedRegexList []string
	}{
		{
			desc:         "no annotation defaults to any origin",
			expectedList: []string{"*"},
		},
		{
			desc:         "any origin",
			allowOrigin:  "*",
			expectedList: []string{"*"},
		},
		{
			desc:         "exact origin",
			allowOrigin:  "https://example.com",
			expectedList: []string{"https://example.com"},
		},
		{
			desc:              "single-level wildcard origin",
			allowOrigin:       "https://*.example.com",
			expectedRegexList: []string{`^(?i)https://[A-Za-z0-9-]+\.example\.com$`},
		},
		{
			desc:              "single-level wildcard origin with a port",
			allowOrigin:       "https://*.example.com:8443",
			expectedRegexList: []string{`^(?i)https://[A-Za-z0-9-]+\.example\.com:8443$`},
		},
		{
			desc:              "single-level wildcard origin on a subdomain",
			allowOrigin:       "https://*.foo.example.com",
			expectedRegexList: []string{`^(?i)https://[A-Za-z0-9-]+\.foo\.example\.com$`},
		},
		{
			desc:              "exact and wildcard origins",
			allowOrigin:       "https://*.example.com, https://exact.example.org",
			expectedList:      []string{"https://exact.example.org"},
			expectedRegexList: []string{`^(?i)https://[A-Za-z0-9-]+\.example\.com$`},
		},
		{
			desc:        "unsupported wildcard not followed by a label separator",
			allowOrigin: "https://*example.com",
		},
		{
			desc:        "unsupported wildcard in the middle of the host",
			allowOrigin: "https://foo.*.example.com",
		},
		{
			desc:        "unsupported multiple wildcards",
			allowOrigin: "https://*.*.example.com",
		},
		{
			desc:        "unsupported wildcard without a scheme",
			allowOrigin: "*.example.com",
		},
		{
			desc:        "unsupported wildcard with an uppercase scheme",
			allowOrigin: "HTTPS://*.example.com",
		},
		{
			desc:        "unsupported wildcard with a path",
			allowOrigin: "https://*.example.com/foo",
		},
		{
			desc:         "unsupported wildcard alongside an exact origin",
			allowOrigin:  "https://*example.com, https://exact.example.org",
			expectedList: []string{"https://exact.example.org"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			annotations := map[string]string{"nginx.ingress.kubernetes.io/enable-cors": "true"}
			if test.allowOrigin != "" {
				annotations["nginx.ingress.kubernetes.io/cors-allow-origin"] = test.allowOrigin
			}

			var ing netv1.Ingress
			ing.SetAnnotations(annotations)

			loc := &location{Config: parseIngressConfig(&ing)}

			var p Provider
			p.buildCORS(loc)

			require.NotNil(t, loc.CORS)
			assert.Equal(t, test.expectedList, loc.CORS.AccessControlAllowOriginList)
			assert.Equal(t, test.expectedRegexList, loc.CORS.AccessControlAllowOriginListRegex)
		})
	}
}
