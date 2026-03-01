package requestdecorator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCNAMEFlatten(t *testing.T) {
	testCases := []struct {
		desc           string
		resolvFile     string
		domain         string
		expectedDomain string
	}{
		{
			desc:           "host request is CNAME record",
			resolvFile:     "/etc/resolv.conf",
			domain:         "www.github.com",
			expectedDomain: "github.com",
		},
		{
			desc:           "resolve file not found",
			resolvFile:     "/etc/resolv.oops",
			domain:         "www.github.com",
			expectedDomain: "www.github.com",
		},
		{
			desc:           "host request is not CNAME record",
			resolvFile:     "/etc/resolv.conf",
			domain:         "github.com",
			expectedDomain: "github.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			hostResolver := &Resolver{
				ResolvConfig: test.resolvFile,
				ResolvDepth:  5,
			}

			flatH := hostResolver.CNAMEFlatten(t.Context(), test.domain)
			assert.Equal(t, test.expectedDomain, flatH)
		})
	}
}
