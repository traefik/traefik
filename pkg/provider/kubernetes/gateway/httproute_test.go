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

func Test_convertSessionPersistence(t *testing.T) {
	testCases := []struct {
		desc           string
		sessionPersist *gatev1.SessionPersistence
		wantNil        bool
		wantCookie     bool
		wantHeader     bool
		wantName       string
		wantMaxAge     int
	}{
		{
			desc:           "nil session persistence",
			sessionPersist: nil,
			wantNil:        true,
		},
		{
			desc:           "default cookie type (nil Type)",
			sessionPersist: &gatev1.SessionPersistence{},
			wantNil:        false,
			wantCookie:     true,
			wantHeader:     false,
		},
		{
			desc: "explicit cookie type",
			sessionPersist: &gatev1.SessionPersistence{
				Type: ptr.To(gatev1.CookieBasedSessionPersistence),
			},
			wantNil:    false,
			wantCookie: true,
			wantHeader: false,
		},
		{
			desc: "header type",
			sessionPersist: &gatev1.SessionPersistence{
				Type: ptr.To(gatev1.HeaderBasedSessionPersistence),
			},
			wantNil:    false,
			wantCookie: false,
			wantHeader: true,
		},
		{
			desc: "cookie with session name",
			sessionPersist: &gatev1.SessionPersistence{
				SessionName: ptr.To("my-session"),
				Type:        ptr.To(gatev1.CookieBasedSessionPersistence),
			},
			wantNil:    false,
			wantCookie: true,
			wantHeader: false,
			wantName:   "my-session",
		},
		{
			desc: "header with session name",
			sessionPersist: &gatev1.SessionPersistence{
				SessionName: ptr.To("X-My-Session"),
				Type:        ptr.To(gatev1.HeaderBasedSessionPersistence),
			},
			wantNil:    false,
			wantCookie: false,
			wantHeader: true,
			wantName:   "X-My-Session",
		},
		{
			desc: "cookie with permanent lifetime and timeout",
			sessionPersist: &gatev1.SessionPersistence{
				SessionName:     ptr.To("my-cookie"),
				Type:            ptr.To(gatev1.CookieBasedSessionPersistence),
				AbsoluteTimeout: ptr.To(gatev1.Duration("1h")),
				CookieConfig: &gatev1.CookieConfig{
					LifetimeType: ptr.To(gatev1.PermanentCookieLifetimeType),
				},
			},
			wantNil:    false,
			wantCookie: true,
			wantHeader: false,
			wantName:   "my-cookie",
			wantMaxAge: 3600,
		},
		{
			desc: "cookie with session lifetime ignores timeout",
			sessionPersist: &gatev1.SessionPersistence{
				SessionName:     ptr.To("my-cookie"),
				Type:            ptr.To(gatev1.CookieBasedSessionPersistence),
				AbsoluteTimeout: ptr.To(gatev1.Duration("1h")),
				CookieConfig: &gatev1.CookieConfig{
					LifetimeType: ptr.To(gatev1.SessionCookieLifetimeType),
				},
			},
			wantNil:    false,
			wantCookie: true,
			wantHeader: false,
			wantName:   "my-cookie",
			wantMaxAge: 0, // Session cookie has no MaxAge
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := convertSessionPersistence(test.sessionPersist)

			if test.wantNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)

			if test.wantCookie {
				assert.NotNil(t, result.Cookie)
				assert.Nil(t, result.Header)
				if test.wantName != "" {
					assert.Equal(t, test.wantName, result.Cookie.Name)
				}
				assert.Equal(t, test.wantMaxAge, result.Cookie.MaxAge)
			}

			if test.wantHeader {
				assert.Nil(t, result.Cookie)
				assert.NotNil(t, result.Header)
				if test.wantName != "" {
					assert.Equal(t, test.wantName, result.Header.Name)
				}
			}
		})
	}
}
