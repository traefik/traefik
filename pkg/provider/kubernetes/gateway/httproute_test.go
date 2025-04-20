package gateway

import (
	"fmt"
	"github.com/stretchr/testify/require"
	traefikhttp "github.com/traefik/traefik/v3/pkg/muxer/http"
	"net/http"
	"net/http/httptest"
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
			expectedRule:     "PathPrefix(`/`)",
			expectedPriority: 1,
		},
		{
			desc:             "One Host rule without match",
			hostnames:        []gatev1.Hostname{"foo.com"},
			expectedRule:     "Host(`foo.com`) && PathPrefix(`/`)",
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
			expectedRule:     "PathPrefix(`/`)",
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
			expectedRule:     "PathPrefix(`/`) && Header(`foo`,`bar`)",
			expectedPriority: 101,
		},
		{
			desc:             "One HTTPRouteMatch with nil HTTPPathMatch",
			match:            gatev1.HTTPRouteMatch{Path: nil},
			expectedRule:     "PathPrefix(`/`)",
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
			expectedRule:     "(Path(`/foo`) || PathPrefix(`/foo/`))",
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
			expectedRule:     "Path(`/`)",
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
			expectedRule:     "Path(`/foo/`)",
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
			expectedRule:     "Path(`/foo/`) && Header(`my-header`,`foo`)",
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
			expectedRule:     "Host(`foo.com`) && Path(`/foo/`) && Header(`my-header`,`foo`)",
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

func TestStrictPrefixMatchingRule(t *testing.T) {
	tests := []struct {
		path        string
		requestPath string
		match       bool
	}{ // The tests are taken from https://kubernetes.io/docs/concepts/services-networking/ingress/#examples
		{
			path:        "/foo",
			requestPath: "/foo",
			match:       true,
		},
		{
			path:        "/foo",
			requestPath: "/foo/",
			match:       true,
		},
		{
			path:        "/foo/",
			requestPath: "/foo",
			match:       true,
		},
		{
			path:        "/foo/",
			requestPath: "/foo/",
			match:       true,
		},
		{
			path:        "/aaa/bb",
			requestPath: "/aaa/bbb",
			match:       false,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb",
			match:       true,
		},
		{
			path:        "/aaa/bbb/",
			requestPath: "/aaa/bbb",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb/",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb/ccc",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbbxyz",
			match:       false,
		},
		{
			path:        "/",
			requestPath: "/aaa/ccc",
			match:       true,
		},
		{
			path:        "/aaa",
			requestPath: "/aaa/ccc",
			match:       true,
		},
		{
			path:        "/...",
			requestPath: "/aaa",
			match:       false,
		},
		{
			path:        "/...",
			requestPath: "/.../",
			match:       true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Prefix match case %s", tt.path), func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			muxer, err := traefikhttp.NewMuxer()
			require.NoError(t, err)

			rule := BuildPathPrefixRule(tt.path)
			err = muxer.AddRoute(rule, "", 0, handler)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, http.NoBody)
			muxer.ServeHTTP(w, req)

			if tt.match {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		})
	}
}
