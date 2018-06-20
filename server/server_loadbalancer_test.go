package server

import (
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func TestConfigureBackends(t *testing.T) {
	validMethod := "Drr"
	defaultMethod := "wrr"

	testCases := []struct {
		desc               string
		lb                 *types.LoadBalancer
		expectedMethod     string
		expectedStickiness *types.Stickiness
	}{
		{
			desc: "valid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method:     validMethod,
				Stickiness: &types.Stickiness{},
			},
			expectedMethod:     validMethod,
			expectedStickiness: &types.Stickiness{},
		},
		{
			desc: "valid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method:     validMethod,
				Stickiness: nil,
			},
			expectedMethod: validMethod,
		},
		{
			desc: "invalid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method:     "Invalid",
				Stickiness: &types.Stickiness{},
			},
			expectedMethod:     defaultMethod,
			expectedStickiness: &types.Stickiness{},
		},
		{
			desc: "invalid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method:     "Invalid",
				Stickiness: nil,
			},
			expectedMethod: defaultMethod,
		},
		{
			desc:           "missing load balancer",
			lb:             nil,
			expectedMethod: defaultMethod,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			backend := &types.Backend{
				LoadBalancer: test.lb,
			}

			configureBackends(map[string]*types.Backend{
				"backend": backend,
			})

			expected := types.LoadBalancer{
				Method:     test.expectedMethod,
				Stickiness: test.expectedStickiness,
			}

			assert.Equal(t, expected, *backend.LoadBalancer)
		})
	}
}
