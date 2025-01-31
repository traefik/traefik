package dynamic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			ipv6Subnet:  intPtr(0),
		},
		{
			desc:        "Subnet greater that 128",
			expectError: true,
			ipv6Subnet:  intPtr(129),
		},
		{
			desc:       "Valid subnet",
			ipv6Subnet: intPtr(128),
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

func intPtr(value int) *int {
	return &value
}
