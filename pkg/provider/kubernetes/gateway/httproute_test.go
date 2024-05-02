package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_hostRule(t *testing.T) {
	testCases := []struct {
		desc         string
		hostnames    []gatev1.Hostname
		expectedRule string
		expectErr    bool
	}{
		{
			desc:         "Empty rule and matches",
			expectedRule: "",
		},
		{
			desc: "One Host",
			hostnames: []gatev1.Hostname{
				"Foo",
			},
			expectedRule: "Host(`Foo`)",
		},
		{
			desc: "Multiple Hosts",
			hostnames: []gatev1.Hostname{
				"Foo",
				"Bar",
				"Bir",
			},
			expectedRule: "(Host(`Foo`) || Host(`Bar`) || Host(`Bir`))",
		},
		{
			desc: "Several Host and wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
				"bar.foo",
				"foo.foo",
			},
			expectedRule: "(HostRegexp(`^[a-z0-9-\\.]+\\.bar\\.foo$`) || Host(`bar.foo`) || Host(`foo.foo`))",
		},
		{
			desc: "Host with wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
			},
			expectedRule: "HostRegexp(`^[a-z0-9-\\.]+\\.bar\\.foo$`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule := hostRule(test.hostnames)
			assert.Equal(t, test.expectedRule, rule)
		})
	}
}

func Test_routerRule(t *testing.T) {
	testCases := []struct {
		desc          string
		routeRule     gatev1.HTTPRouteRule
		hostRule      string
		expectedRule  string
		expectedError bool
	}{
		{
			desc:         "Empty rule and matches",
			expectedRule: "PathPrefix(`/`)",
		},
		{
			desc:         "One Host rule without matches",
			hostRule:     "Host(`foo.com`)",
			expectedRule: "Host(`foo.com`) && PathPrefix(`/`)",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: ptr.To(gatev1.HTTPPathMatch{
							Type:  ptr.To(gatev1.PathMatchPathPrefix),
							Value: ptr.To("/"),
						}),
						Headers: nil,
					},
				},
			},
			expectedRule: "PathPrefix(`/`)",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch Type",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
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
			},
			expectedRule: "PathPrefix(`/`) && Header(`foo`,`bar`)",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{Path: nil},
				},
			},
			expectedRule: "PathPrefix(`/`)",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Type",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  nil,
							Value: ptr.To("/foo/"),
						},
					},
				},
			},
			expectedRule: "(Path(`/foo`) || PathPrefix(`/foo/`))",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Values",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  ptr.To(gatev1.PathMatchExact),
							Value: nil,
						},
					},
				},
			},
			expectedRule: "Path(`/`)",
		},
		{
			desc: "One Path in matches",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  ptr.To(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
				},
			},
			expectedRule: "Path(`/foo/`)",
		},
		{
			desc: "One Path in matches and another empty",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  ptr.To(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
					{},
				},
			},
			expectedRule: "Path(`/foo/`) || PathPrefix(`/`)",
		},
		{
			desc: "Path OR Header rules",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
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
			},
			expectedRule: "Path(`/foo/`) || PathPrefix(`/`) && Header(`my-header`,`foo`)",
		},
		{
			desc: "Path && Header rules",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
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
			},
			expectedRule: "Path(`/foo/`) && Header(`my-header`,`foo`)",
		},
		{
			desc:     "Host && Path && Header rules",
			hostRule: "Host(`foo.com`)",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
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
			},
			expectedRule: "Host(`foo.com`) && Path(`/foo/`) && Header(`my-header`,`foo`)",
		},
		{
			desc:     "Host && (Path || Header) rules",
			hostRule: "Host(`foo.com`)",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
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
			},
			expectedRule: "Host(`foo.com`) && (Path(`/foo/`) || PathPrefix(`/`) && Header(`my-header`,`foo`))",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule := routerRule(test.routeRule, test.hostRule)
			assert.Equal(t, test.expectedRule, rule)
		})
	}
}
