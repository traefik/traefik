package ingressnginx

import (
	"testing"
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestApplyRateLimitConfiguration(t *testing.T) {
	tests := []struct {
		desc                string
		ingressConfig       ingressConfig
		expectedMiddlewares []string
		expectedRateLimit   *dynamic.RateLimit
	}{
		{
			desc:                "no limit-rpm",
			ingressConfig:       ingressConfig{},
			expectedMiddlewares: nil,
			expectedRateLimit:   nil,
		},
		{
			desc: "invalid limit-rpm",
			ingressConfig: ingressConfig{
				LimitRPM: ptr.To(0),
			},
			expectedMiddlewares: nil,
			expectedRateLimit:   nil,
		},
		{
			desc: "valid limit-rpm",
			ingressConfig: ingressConfig{
				LimitRPM: ptr.To(120),
			},
			expectedMiddlewares: []string{"test-router-rate-limit"},
			expectedRateLimit: &dynamic.RateLimit{
				Average: 120,
				Period:  ptypes.Duration(time.Minute),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conf := &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
				},
			}
			rt := &dynamic.Router{}

			applyRateLimitConfiguration("test-router", test.ingressConfig, rt, conf)

			assert.Equal(t, test.expectedMiddlewares, rt.Middlewares)

			mw, found := conf.HTTP.Middlewares["test-router-rate-limit"]
			if test.expectedRateLimit == nil {
				assert.False(t, found)
				return
			}

			assert.True(t, found)
			assert.Equal(t, test.expectedRateLimit, mw.RateLimit)
		})
	}
}
