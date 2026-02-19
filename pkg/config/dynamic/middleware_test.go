package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func Test_GetStrategy_ipv6Subnet(t *testing.T) {
	testCases := []struct {
		desc        string
		expectError bool
		ipv6Subnet  *int
	}{
		{
			desc: "Nil subnet",
		},
		{
			desc:        "Zero subnet",
			expectError: true,
			ipv6Subnet:  ptr.To(0),
		},
		{
			desc:        "Subnet greater that 128",
			expectError: true,
			ipv6Subnet:  ptr.To(129),
		},
		{
			desc:       "Valid subnet",
			ipv6Subnet: ptr.To(128),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			strategy := IPStrategy{
				IPv6Subnet: test.ipv6Subnet,
			}

			get, err := strategy.Get()
			if test.expectError {
				require.Error(t, err)
				assert.Nil(t, get)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, get)
			}
		})
	}
}

func TestHasSecureHeadersDefined(t *testing.T) {
	testCases := []struct {
		desc     string
		headers  *Headers
		expected bool
	}{
		{
			desc:     "Nil headers",
			headers:  nil,
			expected: false,
		},
		{
			desc:     "Empty headers",
			headers:  &Headers{},
			expected: false,
		},
		{
			desc: "STSSeconds set to non-zero",
			headers: &Headers{
				STSSeconds: ptr.To(int64(42)),
			},
			expected: true,
		},
		{
			desc: "STSSeconds set to zero",
			headers: &Headers{
				STSSeconds: ptr.To(int64(0)),
			},
			expected: true,
		},
		{
			desc: "STSSeconds nil (not set)",
			headers: &Headers{
				FrameDeny: true,
			},
			expected: true,
		},
		{
			desc: "Only ForceSTSHeader",
			headers: &Headers{
				ForceSTSHeader: true,
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := test.headers.HasSecureHeadersDefined()
			assert.Equal(t, test.expected, result)
		})
	}
}
