package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		endpoints []string
		expErr    string
	}{
		{
			desc:      "no endpoint defined",
			endpoints: nil,
			expErr:    "no endpoint defined",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{}
			provider.SetDefaults()
			provider.Endpoints = test.endpoints

			require.EqualError(t, provider.Init(), test.expErr)
		})
	}
}
