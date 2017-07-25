package container

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateMemUsageUnixNoCache(t *testing.T) {
	// Given
	stats := types.MemoryStats{Usage: 500, Stats: map[string]uint64{"cache": 400}}

	// When
	result := calculateMemUsageUnixNoCache(stats)

	// Then
	assert.InDelta(t, 100.0, result, 1e-6)
}

func TestCalculateMemPercentUnixNoCache(t *testing.T) {
	// Given
	someLimit := float64(100.0)
	noLimit := float64(0.0)
	used := float64(70.0)

	// When and Then
	t.Run("Limit is set", func(t *testing.T) {
		result := calculateMemPercentUnixNoCache(someLimit, used)
		assert.InDelta(t, 70.0, result, 1e-6)
	})
	t.Run("No limit, no cgroup data", func(t *testing.T) {
		result := calculateMemPercentUnixNoCache(noLimit, used)
		assert.InDelta(t, 0.0, result, 1e-6)
	})
}
