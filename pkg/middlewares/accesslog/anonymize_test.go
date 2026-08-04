package accesslog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

func TestMaskClientAddr(t *testing.T) {
	anon := &otypes.AccessLogAnonymization{IPv4Subnet: 24, IPv6Subnet: 80}

	assert.Equal(t, "203.0.113.0:8080", maskClientAddr("203.0.113.42:8080", anon))
	assert.Equal(t, "[2001:db8:0:0:abcd::]:8080", maskClientAddr("[2001:db8::abcd:ffff:c0a8:1]:8080", anon))
	assert.Equal(t, "203.0.113.0", maskClientAddr("203.0.113.42", anon))
}

func TestNewHandlerAnonymizationValidation(t *testing.T) {
	testCases := []struct {
		desc      string
		anon      *otypes.AccessLogAnonymization
		expectErr bool
	}{
		{
			desc: "valid subnets",
			anon: &otypes.AccessLogAnonymization{IPv4Subnet: 24, IPv6Subnet: 80},
		},
		{
			desc:      "IPv4 subnet too large",
			anon:      &otypes.AccessLogAnonymization{IPv4Subnet: 33},
			expectErr: true,
		},
		{
			desc:      "IPv6 subnet too large",
			anon:      &otypes.AccessLogAnonymization{IPv6Subnet: 129},
			expectErr: true,
		},
		{
			desc:      "negative subnet",
			anon:      &otypes.AccessLogAnonymization{IPv4Subnet: -1},
			expectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			config := &otypes.AccessLog{Format: CommonFormat, Anonymization: test.anon}
			handler, err := NewHandler(context.Background(), config)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, handler.Close())
		})
	}
}
