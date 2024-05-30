package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			expectedRule:     "Host(`Foo`)",
			expectedPriority: 3,
		},
		{
			desc: "Multiple Hosts",
			hostnames: []gatev1.Hostname{
				"Foo",
				"Bar",
				"Bir",
			},
			expectedRule:     "(Host(`Foo`) || Host(`Bar`) || Host(`Bir`))",
			expectedPriority: 3,
		},
		{
			desc: "Several Host and wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
				"bar.foo",
				"foo.foo",
			},
			expectedRule:     "(HostRegexp(`^[a-z0-9-\\.]+\\.bar\\.foo$`) || Host(`bar.foo`) || Host(`foo.foo`))",
			expectedPriority: 9,
		},
		{
			desc: "Host with wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
			},
			expectedRule:     "HostRegexp(`^[a-z0-9-\\.]+\\.bar\\.foo$`)",
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

func Test_buildRouterRule(t *testing.T) {
	testCases := []struct {
		desc             string
		routeMatches     []gatev1.HTTPRouteMatch
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectedError    bool
	}{
		{
			desc:             "Empty rule and matches ",
			expectedRule:     "PathPrefix(`/`)",
			expectedPriority: 1,
		},
		{
			desc:             "One Host rule without matches",
			hostnames:        []gatev1.Hostname{"foo.com"},
			expectedRule:     "Host(`foo.com`)",
			expectedPriority: 7,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: ptr.To(gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchPathPrefix),
						Value: ptr.To("/"),
					}),
					Headers: nil,
				},
			},
			expectedRule:     "PathPrefix(`/`)",
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch Type",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: ptr.To(gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchPathPrefix),
						Value: ptr.To("/"),
					}),
					Headers: []gatev1.HTTPHeaderMatch{
						{Name: "foo", Value: "bar"},
					},
				},
			},
			expectedRule:     "PathPrefix(`/`) && Header(`foo`,`bar`)",
			expectedPriority: 91,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch",
			routeMatches: []gatev1.HTTPRouteMatch{
				{Path: nil},
			},
			expectedRule:     "PathPrefix(`/`)",
			expectedPriority: 1,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Type",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  nil,
						Value: ptr.To("/foo/"),
					},
				},
			},
			expectedRule:     "(Path(`/foo`) || PathPrefix(`/foo/`))",
			expectedPriority: 10490,
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Values",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchExact),
						Value: nil,
					},
				},
			},
			expectedRule:     "Path(`/`)",
			expectedPriority: 99990,
		},
		{
			desc: "One Path in matches",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchExact),
						Value: ptr.To("/foo/"),
					},
				},
			},
			expectedRule:     "Path(`/foo/`)",
			expectedPriority: 99990,
		},
		{
			desc: "One Path in matches and another empty",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchExact),
						Value: ptr.To("/foo/"),
					},
				},
				{},
			},
			expectedRule:     "Path(`/foo/`) || PathPrefix(`/`)",
			expectedPriority: 99980,
		},
		{
			desc: "Path OR Header rules",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchExact),
						Value: ptr.To("/foo/"),
					},
				},
				{
					Headers: []gatev1.HTTPHeaderMatch{
						{
							Type:  ptr.To(gatev1.HeaderMatchExact),
							Name:  "my-header",
							Value: "foo",
						},
					},
				},
			},
			expectedRule:     "Path(`/foo/`) || PathPrefix(`/`) && Header(`my-header`,`foo`)",
			expectedPriority: 99980,
		},
		{
			desc: "Path && Header rules",
			routeMatches: []gatev1.HTTPRouteMatch{
				{
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
			},
			expectedRule:     "Path(`/foo/`) && Header(`my-header`,`foo`)",
			expectedPriority: 100090,
		},
		{
			desc:      "Host && Path && Header rules",
			hostnames: []gatev1.Hostname{"foo.com"},
			routeMatches: []gatev1.HTTPRouteMatch{
				{
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
			},
			expectedRule:     "Host(`foo.com`) && (Path(`/foo/`) && Header(`my-header`,`foo`))",
			expectedPriority: 100097,
		},
		{
			desc:      "Host && (Path || Header) rules",
			hostnames: []gatev1.Hostname{"foo.com"},
			routeMatches: []gatev1.HTTPRouteMatch{
				{
					Path: &gatev1.HTTPPathMatch{
						Type:  ptr.To(gatev1.PathMatchExact),
						Value: ptr.To("/foo/"),
					},
				},
				{
					Headers: []gatev1.HTTPHeaderMatch{
						{
							Type:  ptr.To(gatev1.HeaderMatchExact),
							Name:  "my-header",
							Value: "foo",
						},
					},
				},
			},
			expectedRule:     "Host(`foo.com`) && (Path(`/foo/`) || PathPrefix(`/`) && Header(`my-header`,`foo`))",
			expectedPriority: 99987,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildRouterRule(test.hostnames, test.routeMatches)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}
