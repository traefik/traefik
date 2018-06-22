package hostresolver

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCNAMEFlatten(t *testing.T) {
	hostResolver := &HostResolver{
		Enabled:      false,
		ResolvConfig: "/etc/resolv.conf",
		ResolvDepth:  5,
	}

	testCase := []struct {
		desc           string
		domain         string
		expectedDomain string
		isCNAME        bool
	}{
		{
			desc:           "host request is CNAME record",
			domain:         "www.github.com",
			expectedDomain: "github.com",
			isCNAME:        true,
		},
		{
			desc:           "host request is not CNAME record",
			domain:         "github.com",
			expectedDomain: "github.com",
			isCNAME:        false,
		},
	}
	for _, test := range testCase {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			reqH, flatH := hostResolver.CNAMEFlatten(test.domain)
			if test.isCNAME {
				assert.Equal(t, test.domain, reqH)
				assert.Equal(t, test.expectedDomain, flatH)
				assert.NotEqual(t, test.expectedDomain, reqH)
			} else {
				assert.Equal(t, test.domain, reqH)
				assert.Equal(t, test.expectedDomain, flatH)
				assert.Equal(t, test.expectedDomain, reqH)
			}
		})

	}
}
