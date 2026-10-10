package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/ping"
)

func TestIsPingEntryPoint(t *testing.T) {
	testCases := []struct {
		desc           string
		mgr            *ObservabilityMgr
		entryPointName string
		expected       bool
	}{
		{
			desc:           "nil ObservabilityMgr receiver",
			mgr:            nil,
			entryPointName: "websecure",
			expected:       false,
		},
		{
			desc:           "ping config is nil",
			mgr:            &ObservabilityMgr{config: static.Configuration{}},
			entryPointName: "websecure",
			expected:       false,
		},
		{
			desc: "entry point matches configured ping entry point",
			mgr: &ObservabilityMgr{
				config: static.Configuration{
					Ping: &ping.Handler{EntryPoint: "websecure"},
				},
			},
			entryPointName: "websecure",
			expected:       true,
		},
		{
			desc: "entry point does not match configured ping entry point",
			mgr: &ObservabilityMgr{
				config: static.Configuration{
					Ping: &ping.Handler{EntryPoint: "websecure"},
				},
			},
			entryPointName: "web",
			expected:       false,
		},
		{
			desc: "non-ping entry point on same traefik instance",
			mgr: &ObservabilityMgr{
				config: static.Configuration{
					Ping: &ping.Handler{EntryPoint: "ping"},
				},
			},
			entryPointName: "websecure",
			expected:       false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, test.mgr.IsPingEntryPoint(test.entryPointName))
		})
	}
}
