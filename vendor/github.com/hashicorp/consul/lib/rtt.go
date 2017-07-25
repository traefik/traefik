package lib

import (
	"math"
	"time"

	"github.com/hashicorp/serf/coordinate"
)

// ComputeDistance returns the distance between the two network coordinates in
// seconds. If either of the coordinates is nil then this will return positive
// infinity.
func ComputeDistance(a *coordinate.Coordinate, b *coordinate.Coordinate) float64 {
	if a == nil || b == nil {
		return math.Inf(1.0)
	}

	return a.DistanceTo(b).Seconds()
}

// GenerateCoordinate creates a new coordinate with the given distance from the
// origin. This should only be used for tests.
func GenerateCoordinate(rtt time.Duration) *coordinate.Coordinate {
	coord := coordinate.NewCoordinate(coordinate.DefaultConfig())
	coord.Vec[0] = rtt.Seconds()
	coord.Height = 0
	return coord
}
