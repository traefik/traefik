package hostresolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCNAMEFlatten(t *testing.T) {
	testCase := []struct {
		desc           string
		resolvFile     string
		domain         string
		expectedDomain string
		isCNAME        bool
	}{
		{
			desc:           "host request is CNAME record",
			resolvFile:     "/etc/resolv.conf",
			domain:         "www.github.com",
			expectedDomain: "github.com",
			isCNAME:        true,
		},
		{
			desc:           "resolve file not found",
			resolvFile:     "/etc/resolv.oops",
			domain:         "www.github.com",
			expectedDomain: "www.github.com",
			isCNAME:        false,
		},
		{
			desc:           "host request is not CNAME record",
			resolvFile:     "/etc/resolv.conf",
			domain:         "github.com",
			expectedDomain: "github.com",
			isCNAME:        false,
		},
	}

	for _, test := range testCase {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			hostResolver := &Resolver{
				ResolvConfig: test.resolvFile,
				ResolvDepth:  5,
			}

			reqH, flatH := hostResolver.CNAMEFlatten(test.domain)
			assert.Equal(t, test.domain, reqH)
			assert.Equal(t, test.expectedDomain, flatH)

			if test.isCNAME {
				assert.NotEqual(t, test.expectedDomain, reqH)
			} else {
				assert.Equal(t, test.expectedDomain, reqH)
			}
		})
	}
}
