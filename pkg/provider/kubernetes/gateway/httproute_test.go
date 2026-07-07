package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_buildHostRule(t *testing.T) {
	testCases := []struct {
		desc             string
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectErr        bool
	}{
		{
			desc:         "Empty (should not happen)",
			expectedRule: "",
		},
		{
			desc: "One Host",
			hostnames: []gatev1.Hostname{
				"Foo",
			},
			expectedRule:     `Host("Foo")`,
			expectedPriority: 3,
		},
		{
			desc: "Multiple Hosts",
			hostnames: []gatev1.Hostname{
				"Foo",
				"Bar",
				"Bir",
			},
			expectedRule:     `(Host("Foo") || Host("Bar") || Host("Bir"))`,
			expectedPriority: 3,
		},
		{
			desc: "Several Host and wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
				"bar.foo",
				"foo.foo",
			},
			expectedRule:     `(HostRegexp("^[a-z0-9-\\.]+\\.bar\\.foo$") || Host("bar.foo") || Host("foo.foo"))`,
			expectedPriority: 9,
		},
		{
			desc: "Host with wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
			},
			expectedRule:     `HostRegexp("^[a-z0-9-\\.]+\\.bar\\.foo$")`,
			expectedPriority: 9,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildHostRule(test.hostnames)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}

func Test_buildMatchRule(t *testing.T) {
	testCases := []struct {
		desc             string
		match            gatev1.HTTPRouteMatch
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectedError    bool
	}{
		{
			desc:             "Empty rule and matches",
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc:             "One Host rule without match",
			hostnames:        []gatev1.Hostname{"foo.com"},
			expectedRule:     `Host("foo.com") && PathPrefix("/")`,
			expectedPriority: 8,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch",
			match: gatev1.HTTPRouteMatch{
				Path: ptr.To(gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchPathPrefix),
					Value: ptr.To("/"),
				}),
				Headers: nil,
			},
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch Type",
			match: gatev1.HTTPRouteMatch{
				Path: ptr.To(gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchPathPrefix),
					Value: ptr.To("/"),
				}),
				Headers: []gatev1.HTTPHeaderMatch{
					{Name: "foo", Value: "bar"},
				},
			},
			expectedRule:     `PathPrefix("/") && Header("foo","bar")`,
			expectedPriority: 101,
		},
		{
			desc:             "One HTTPRouteMatch with nil HTTPPathMatch",
			match:            gatev1.HTTPRouteMatch{Path: nil},
			expectedRule:     `PathPrefix("/")`,
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Type",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  nil,
					Value: ptr.To("/foo/"),
				},
			},
			expectedRule:     `(Path("/foo") || PathPrefix("/foo/"))`,
			expectedPriority: 10500,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Values",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchExact),
					Value: nil,
				},
			},
			expectedRule:     `Path("/")`,
			expectedPriority: 100000,
		},
		{
			desc: "One Path",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchExact),
					Value: ptr.To("/foo/"),
				},
			},
			expectedRule:     `Path("/foo/")`,
			expectedPriority: 100000,
		},
		{
			desc: "Path && Header",
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchExact),
					Value: ptr.To("/foo/"),
				},
				Headers: []gatev1.HTTPHeaderMatch{
					{
						Type:  ptr.To(gatev1.HeaderMatchExact),
						Name:  "my-header",
						Value: "foo",
					},
				},
			},
			expectedRule:     `Path("/foo/") && Header("my-header","foo")`,
			expectedPriority: 100100,
		},
		{
			desc:      "Host && Path && Header",
			hostnames: []gatev1.Hostname{"foo.com"},
			match: gatev1.HTTPRouteMatch{
				Path: &gatev1.HTTPPathMatch{
					Type:  ptr.To(gatev1.PathMatchExact),
					Value: ptr.To("/foo/"),
				},
				Headers: []gatev1.HTTPHeaderMatch{
					{
						Type:  ptr.To(gatev1.HeaderMatchExact),
						Name:  "my-header",
						Value: "foo",
					},
				},
			},
			expectedRule:     `Host("foo.com") && Path("/foo/") && Header("my-header","foo")`,
			expectedPriority: 100107,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildMatchRule(test.hostnames, test.match)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}

func Test_createCORS(t *testing.T) {
	testCases := []struct {
		desc     string
		filter   *gatev1.HTTPCORSFilter
		expected *dynamic.Middleware
	}{
		{
			desc:   "Empty filter",
			filter: &gatev1.HTTPCORSFilter{},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlMaxAge: ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "Plain origins",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins: []gatev1.CORSOrigin{"https://foo.example.com", "https://bar.example.com"},
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowOriginList: []string{"https://foo.example.com", "https://bar.example.com"},
					AccessControlMaxAge:          ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "Wildcard origin becomes catch-all regex",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins: []gatev1.CORSOrigin{"*"},
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowOriginListRegex: []string{`.*`},
					AccessControlMaxAge:               ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "Origin with bare wildcard host",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins: []gatev1.CORSOrigin{"https://*"},
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowOriginListRegex: []string{`^https://.*$`},
					AccessControlMaxAge:               ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "Origin with wildcard subdomain becomes regex",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins: []gatev1.CORSOrigin{"https://*.example.com"},
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowOriginListRegex: []string{`^https://.*\.example\.com$`},
					AccessControlMaxAge:               ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "Mixed plain and wildcard origins",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins: []gatev1.CORSOrigin{"https://foo.example.com", "https://*.example.com"},
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowOriginList:      []string{"https://foo.example.com"},
					AccessControlAllowOriginListRegex: []string{`^https://.*\.example\.com$`},
					AccessControlMaxAge:               ptr.To(int64(0)),
				},
			},
		},
		{
			desc: "All fields set",
			filter: &gatev1.HTTPCORSFilter{
				AllowOrigins:     []gatev1.CORSOrigin{"https://foo.example.com"},
				AllowCredentials: ptr.To(true),
				AllowMethods:     []gatev1.HTTPMethodWithWildcard{"GET", "POST"},
				AllowHeaders:     []gatev1.HTTPHeaderName{"X-Foo", "X-Bar"},
				ExposeHeaders:    []gatev1.HTTPHeaderName{"X-Baz"},
				MaxAge:           600,
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowCredentials: true,
					AccessControlAllowOriginList:  []string{"https://foo.example.com"},
					AccessControlAllowMethods:     []string{"GET", "POST"},
					AccessControlAllowHeaders:     []string{"X-Foo", "X-Bar"},
					AccessControlExposeHeaders:    []string{"X-Baz"},
					AccessControlMaxAge:           ptr.To(int64(600)),
				},
			},
		},
		{
			desc: "AllowCredentials explicitly false",
			filter: &gatev1.HTTPCORSFilter{
				AllowCredentials: ptr.To(false),
			},
			expected: &dynamic.Middleware{
				Headers: &dynamic.Headers{
					AccessControlAllowCredentials: false,
					AccessControlMaxAge:           ptr.To(int64(0)),
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, createCORS(test.filter))
		})
	}
}
