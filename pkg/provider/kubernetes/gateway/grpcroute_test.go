package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_buildGRPCMatchRule(t *testing.T) {
	testCases := []struct {
		desc             string
		match            gatev1.GRPCRouteMatch
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectedError    bool
	}{
		{
			desc:             "Empty rule and matches",
			expectedRule:     "PathPrefix(`/`)",
			expectedPriority: 15,
		},
		{
			desc:             "One Host rule without match",
			hostnames:        []gatev1.Hostname{"foo.com"},
			expectedRule:     "Host(`foo.com`) && PathPrefix(`/`)",
			expectedPriority: 22,
		},
		{
			desc: "One GRPCRouteMatch with no GRPCHeaderMatch",
			match: gatev1.GRPCRouteMatch{
				Method: &gatev1.GRPCMethodMatch{
					Type:    ptr.To(gatev1.GRPCMethodMatchExact),
					Service: ptr.To("foo"),
					Method:  ptr.To("bar"),
				},
			},
			expectedRule:     "PathRegexp(`/foo/bar`)",
			expectedPriority: 22,
		},
		{
			desc: "One GRPCRouteMatch with one GRPCHeaderMatch",
			match: gatev1.GRPCRouteMatch{
				Method: &gatev1.GRPCMethodMatch{
					Type:    ptr.To(gatev1.GRPCMethodMatchExact),
					Service: ptr.To("foo"),
					Method:  ptr.To("bar"),
				},
				Headers: []gatev1.GRPCHeaderMatch{
					{
						Type:  ptr.To(gatev1.GRPCHeaderMatchExact),
						Name:  "foo",
						Value: "bar",
					},
				},
			},
			expectedRule:     "PathRegexp(`/foo/bar`) && Header(`foo`,`bar`)",
			expectedPriority: 45,
		},
		{
			desc:      "One GRPCRouteMatch with one GRPCHeaderMatch and one Host",
			hostnames: []gatev1.Hostname{"foo.com"},
			match: gatev1.GRPCRouteMatch{
				Method: &gatev1.GRPCMethodMatch{
					Type:    ptr.To(gatev1.GRPCMethodMatchExact),
					Service: ptr.To("foo"),
					Method:  ptr.To("bar"),
				},
				Headers: []gatev1.GRPCHeaderMatch{
					{
						Type:  ptr.To(gatev1.GRPCHeaderMatchExact),
						Name:  "foo",
						Value: "bar",
					},
				},
			},
			expectedRule:     "Host(`foo.com`) && PathRegexp(`/foo/bar`) && Header(`foo`,`bar`)",
			expectedPriority: 52,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := buildGRPCMatchRule(test.hostnames, test.match)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}

func Test_buildGRPCMethodRule(t *testing.T) {
	testCases := []struct {
		desc         string
		method       *gatev1.GRPCMethodMatch
		expectedRule string
	}{
		{
			desc:         "Empty",
			expectedRule: "PathPrefix(`/`)",
		},
		{
			desc: "Exact service matching",
			method: &gatev1.GRPCMethodMatch{
				Type:    ptr.To(gatev1.GRPCMethodMatchExact),
				Service: ptr.To("foo"),
			},
			expectedRule: "PathRegexp(`/foo/[^/]+`)",
		},
		{
			desc: "Exact method matching",
			method: &gatev1.GRPCMethodMatch{
				Type:   ptr.To(gatev1.GRPCMethodMatchExact),
				Method: ptr.To("bar"),
			},
			expectedRule: "PathRegexp(`/[^/]+/bar`)",
		},
		{
			desc: "Exact service and method matching",
			method: &gatev1.GRPCMethodMatch{
				Type:    ptr.To(gatev1.GRPCMethodMatchExact),
				Service: ptr.To("foo"),
				Method:  ptr.To("bar"),
			},
			expectedRule: "PathRegexp(`/foo/bar`)",
		},
		{
			desc: "Regexp service matching",
			method: &gatev1.GRPCMethodMatch{
				Type:    ptr.To(gatev1.GRPCMethodMatchRegularExpression),
				Service: ptr.To("[^1-9/]"),
			},
			expectedRule: "PathRegexp(`/[^1-9/]/[^/]+`)",
		},
		{
			desc: "Regexp method matching",
			method: &gatev1.GRPCMethodMatch{
				Type:   ptr.To(gatev1.GRPCMethodMatchRegularExpression),
				Method: ptr.To("[^1-9/]"),
			},
			expectedRule: "PathRegexp(`/[^/]+/[^1-9/]`)",
		},
		{
			desc: "Regexp service and method matching",
			method: &gatev1.GRPCMethodMatch{
				Type:    ptr.To(gatev1.GRPCMethodMatchRegularExpression),
				Service: ptr.To("[^1-9/]"),
				Method:  ptr.To("[^1-9/]"),
			},
			expectedRule: "PathRegexp(`/[^1-9/]/[^1-9/]`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule := buildGRPCMethodRule(test.method)
			assert.Equal(t, test.expectedRule, rule)
		})
	}
}

func Test_buildGRPCHeaderRules(t *testing.T) {
	testCases := []struct {
		desc          string
		headers       []gatev1.GRPCHeaderMatch
		expectedRules []string
	}{
		{
			desc: "Empty",
		},
		{
			desc: "One exact match type",
			headers: []gatev1.GRPCHeaderMatch{
				{
					Type:  ptr.To(gatev1.GRPCHeaderMatchExact),
					Name:  "foo",
					Value: "bar",
				},
			},
			expectedRules: []string{"Header(`foo`,`bar`)"},
		},
		{
			desc: "One regexp match type",
			headers: []gatev1.GRPCHeaderMatch{
				{
					Type:  ptr.To(gatev1.GRPCHeaderMatchRegularExpression),
					Name:  "foo",
					Value: ".*",
				},
			},
			expectedRules: []string{"HeaderRegexp(`foo`,`.*`)"},
		},
		{
			desc: "One exact and regexp match type",
			headers: []gatev1.GRPCHeaderMatch{
				{
					Type:  ptr.To(gatev1.GRPCHeaderMatchExact),
					Name:  "foo",
					Value: "bar",
				},
				{
					Type:  ptr.To(gatev1.GRPCHeaderMatchRegularExpression),
					Name:  "foo",
					Value: ".*",
				},
			},
			expectedRules: []string{
				"Header(`foo`,`bar`)",
				"HeaderRegexp(`foo`,`.*`)",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule := buildGRPCHeaderRules(test.headers)
			assert.Equal(t, test.expectedRules, rule)
		})
	}
}
