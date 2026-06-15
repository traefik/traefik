package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

func TestBuildRule(t *testing.T) {
	tests := []struct {
		desc     string
		host     string
		loc      *location
		expected string
	}{
		{
			desc:     "empty host and path returns catch-all",
			host:     "",
			loc:      &location{},
			expected: `PathPrefix("/")`,
		},
		{
			desc:     "host only",
			host:     "foo.localhost",
			loc:      &location{},
			expected: `Host("foo.localhost")`,
		},
		{
			desc: "path only",
			host: "",
			loc: &location{
				Path:     "/foo",
				PathType: ptr.To(netv1.PathTypePrefix),
			},
			expected: `(Path("/foo") || PathPrefix("/foo/"))`,
		},
		{
			desc: "host and path",
			host: "foo.localhost",
			loc: &location{
				Path:     "/foo",
				PathType: ptr.To(netv1.PathTypeExact),
			},
			expected: `Host("foo.localhost") && Path("/foo")`,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, buildRule(test.host, test.loc))
		})
	}
}
