package docker

import (
	"testing"

	eventtypes "github.com/moby/moby/api/types/events"
	"github.com/stretchr/testify/assert"
)

func TestShouldHandleEvent(t *testing.T) {
	tests := []struct {
		desc   string
		action eventtypes.Action
		want   bool
	}{
		{
			desc:   "start event triggers reload",
			action: eventtypes.ActionStart,
			want:   true,
		},
		{
			desc:   "die event triggers reload",
			action: eventtypes.ActionDie,
			want:   true,
		},
		{
			desc:   "health_status: healthy does NOT trigger reload",
			action: eventtypes.ActionHealthStatusHealthy,
			want:   false,
		},
		{
			desc:   "health_status: unhealthy triggers reload",
			action: eventtypes.ActionHealthStatusUnhealthy,
			want:   true,
		},
		{
			desc:   "health_status: running triggers reload",
			action: eventtypes.ActionHealthStatusRunning,
			want:   true,
		},
		{
			desc:   "other action triggers reload",
			action: eventtypes.ActionStop,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := shouldHandleEvent(tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}
