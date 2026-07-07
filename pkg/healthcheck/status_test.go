package healthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthStatusObserve(t *testing.T) {
	testCases := []struct {
		desc    string
		status  healthStatus
		results []bool
		updates []healthStatusResult
		up      bool
	}{
		{
			desc:    "marks healthy target unhealthy after enough failures",
			status:  healthStatus{up: true},
			results: []bool{false, false},
			updates: []healthStatusResult{
				{count: 1, threshold: 2},
				{update: true, count: 2, threshold: 2},
			},
			up: false,
		},
		{
			desc:    "success resets failure count",
			status:  healthStatus{up: true},
			results: []bool{false, true, false},
			updates: []healthStatusResult{
				{count: 1, threshold: 2},
				{update: true},
				{count: 1, threshold: 2},
			},
			up: true,
		},
		{
			desc:    "marks unhealthy target healthy after enough passes",
			status:  healthStatus{up: false},
			results: []bool{true, true},
			updates: []healthStatusResult{
				{count: 1, threshold: 2},
				{update: true, count: 2, threshold: 2},
			},
			up: true,
		},
		{
			desc:    "failure resets pass count",
			status:  healthStatus{up: false},
			results: []bool{true, false, true},
			updates: []healthStatusResult{
				{count: 1, threshold: 2},
				{update: true},
				{count: 1, threshold: 2},
			},
			up: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			status := test.status

			for i, result := range test.results {
				assert.Equal(t, test.updates[i], status.observe(result, 2, 2))
			}

			assert.Equal(t, test.up, status.up)
		})
	}
}
